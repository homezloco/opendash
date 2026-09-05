package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/opendash-project/opendash/internal/models"
)

func (s *SQLiteStore) CreateBackupJob(ctx context.Context, v *models.BackupJob) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO backup_jobs(id,app_id,status,error,created_at,updated_at) VALUES(?,?,?,?,?,?)`, v.ID, v.AppID, v.Status, nullable(v.Error), v.CreatedAt.Unix(), v.UpdatedAt.Unix())
	return err
}
func (s *SQLiteStore) UpdateBackupJob(ctx context.Context, v *models.BackupJob) error {
	_, err := s.db.ExecContext(ctx, `UPDATE backup_jobs SET status=?,error=?,updated_at=? WHERE id=?`, v.Status, nullable(v.Error), v.UpdatedAt.Unix(), v.ID)
	return err
}
func (s *SQLiteStore) CreateBackupArchive(ctx context.Context, v *models.BackupArchive) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO backup_archives(id,job_id,app_id,path,size,sha256,manifest_sha256,created_at) VALUES(?,?,?,?,?,?,?,?)`, v.ID, v.JobID, v.AppID, v.Path, v.Size, v.SHA256, v.ManifestSHA256, v.CreatedAt.Unix())
	return err
}
func (s *SQLiteStore) GetBackupArchive(ctx context.Context, id string) (*models.BackupArchive, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id,job_id,app_id,path,size,sha256,manifest_sha256,created_at,verified_at,integrity_ok FROM backup_archives WHERE id=?`, id)
	return scanArchive(row)
}
func (s *SQLiteStore) ListBackupArchives(ctx context.Context, appID string) ([]models.BackupArchive, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,job_id,app_id,path,size,sha256,manifest_sha256,created_at,verified_at,integrity_ok FROM backup_archives WHERE app_id=? ORDER BY created_at DESC`, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.BackupArchive{}
	for rows.Next() {
		v, err := scanArchive(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *v)
	}
	return out, rows.Err()
}

type scanner interface{ Scan(...any) error }

func scanArchive(row scanner) (*models.BackupArchive, error) {
	var v models.BackupArchive
	var created int64
	var verified sql.NullInt64
	var ok sql.NullBool
	if err := row.Scan(&v.ID, &v.JobID, &v.AppID, &v.Path, &v.Size, &v.SHA256, &v.ManifestSHA256, &created, &verified, &ok); err != nil {
		return nil, err
	}
	v.CreatedAt = time.Unix(created, 0).UTC()
	if verified.Valid {
		v.VerifiedAt = time.Unix(verified.Int64, 0).UTC()
	}
	if ok.Valid {
		b := ok.Bool
		v.IntegrityOK = &b
	}
	return &v, nil
}
func (s *SQLiteStore) DeleteBackupArchive(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM backup_archives WHERE id=?`, id)
	return err
}
func (s *SQLiteStore) MarkBackupVerified(ctx context.Context, id string, at time.Time, ok bool) error {
	_, err := s.db.ExecContext(ctx, `UPDATE backup_archives SET verified_at=?,integrity_ok=? WHERE id=?`, at.Unix(), ok, id)
	return err
}
func (s *SQLiteStore) UpsertBackupSchedule(ctx context.Context, v *models.BackupSchedule) error {
	var next any
	if !v.NextRunAt.IsZero() {
		next = v.NextRunAt.Unix()
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO backup_schedules(id,app_id,enabled,interval_hours,retention_count,retention_days,next_run_at,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?) ON CONFLICT(app_id) DO UPDATE SET enabled=excluded.enabled,interval_hours=excluded.interval_hours,retention_count=excluded.retention_count,retention_days=excluded.retention_days,next_run_at=excluded.next_run_at,updated_at=excluded.updated_at`, v.ID, v.AppID, v.Enabled, v.IntervalHours, v.RetentionCount, v.RetentionDays, next, v.CreatedAt.Unix(), v.UpdatedAt.Unix())
	return err
}
func (s *SQLiteStore) GetBackupSchedule(ctx context.Context, appID string) (*models.BackupSchedule, error) {
	var v models.BackupSchedule
	var enabled bool
	var next sql.NullInt64
	var created, updated int64
	err := s.db.QueryRowContext(ctx, `SELECT id,app_id,enabled,interval_hours,retention_count,retention_days,next_run_at,created_at,updated_at FROM backup_schedules WHERE app_id=?`, appID).Scan(&v.ID, &v.AppID, &enabled, &v.IntervalHours, &v.RetentionCount, &v.RetentionDays, &next, &created, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("backup schedule not found: %s", appID)
	}
	if err != nil {
		return nil, err
	}
	v.Enabled = enabled
	v.CreatedAt = time.Unix(created, 0).UTC()
	v.UpdatedAt = time.Unix(updated, 0).UTC()
	if next.Valid {
		v.NextRunAt = time.Unix(next.Int64, 0).UTC()
	}
	return &v, nil
}
func nullable(v string) any {
	if v == "" {
		return nil
	}
	return v
}
