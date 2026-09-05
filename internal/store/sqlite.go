package store

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/opendash-project/opendash/internal/data"
	"github.com/opendash-project/opendash/internal/models"
	_ "modernc.org/sqlite"
)

//go:embed migrations.sql
var migrations string

type SQLiteStore struct {
	db         *sql.DB
	dataSource string
}

func OpenSQLite(ctx context.Context, path string, seedDemo bool) (*SQLiteStore, error) {
	if path == "" {
		return nil, errors.New("database path is required")
	}
	clean, err := filepath.Abs(filepath.Clean(path))
	if err != nil || filepath.Ext(clean) == "" {
		return nil, errors.New("database path must name a file")
	}
	if err := os.MkdirAll(filepath.Dir(clean), 0700); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}
	if info, statErr := os.Lstat(clean); statErr == nil && info.Mode()&os.ModeSymlink != 0 {
		return nil, errors.New("database path must not be a symbolic link")
	} else if statErr != nil && !errors.Is(statErr, os.ErrNotExist) {
		return nil, fmt.Errorf("inspect database file: %w", statErr)
	}
	file, err := os.OpenFile(clean, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, fmt.Errorf("create database file: %w", err)
	}
	if err = file.Chmod(0600); err != nil {
		file.Close()
		return nil, fmt.Errorf("secure database file: %w", err)
	}
	if err = file.Close(); err != nil {
		return nil, fmt.Errorf("close database file: %w", err)
	}
	dsn := (&url.URL{Scheme: "file", Path: clean}).String() + "?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	db.SetMaxOpenConns(1)
	s := &SQLiteStore{db: db, dataSource: "sqlite"}
	if err := s.migrate(ctx); err != nil {
		db.Close()
		return nil, err
	}
	if seedDemo {
		if err := s.seed(ctx); err != nil {
			db.Close()
			return nil, err
		}
		s.dataSource = "demo-sqlite"
	}
	return s, nil
}

func (s *SQLiteStore) migrate(ctx context.Context) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migration: %w", err)
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS schema_migrations (version INTEGER PRIMARY KEY)"); err != nil {
		return fmt.Errorf("initialize migrations: %w", err)
	}
	var version int
	err = tx.QueryRowContext(ctx, "SELECT COALESCE(MAX(version), 0) FROM schema_migrations").Scan(&version)
	if err != nil {
		return fmt.Errorf("read schema version: %w", err)
	}
	migrationBlocks := parseMigrations(migrations)
	if version > len(migrationBlocks) {
		return fmt.Errorf("database schema version %d is newer than supported version %d", version, len(migrationBlocks))
	}
	for i := version; i < len(migrationBlocks); i++ {
		if _, err = tx.ExecContext(ctx, migrationBlocks[i]); err != nil {
			return fmt.Errorf("apply migration %d: %w", i+1, err)
		}
		if _, err = tx.ExecContext(ctx, "INSERT INTO schema_migrations(version) VALUES (?)", i+1); err != nil {
			return fmt.Errorf("record migration %d: %w", i+1, err)
		}
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit migrations: %w", err)
	}
	return nil
}

func parseMigrations(sql string) []string {
	var blocks []string
	var current []string
	for _, line := range strings.Split(sql, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "-- migration:") {
			if len(current) > 0 {
				blocks = append(blocks, strings.Join(current, "\n"))
			}
			current = nil
			continue
		}
		current = append(current, line)
	}
	if len(current) > 0 {
		blocks = append(blocks, strings.Join(current, "\n"))
	}
	return blocks
}

func (s *SQLiteStore) seed(ctx context.Context) error {
	values := map[string]any{"summary": data.Summary, "attention": data.AttentionItems, "apps": data.Apps, "catalog": data.CatalogApps, "protection": data.Protection, "activity": data.ActivityEvents}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for kind, value := range values {
		payload, err := json.Marshal(value)
		if err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, "INSERT OR IGNORE INTO dashboard_data(kind,payload) VALUES (?,?)", kind, payload); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *SQLiteStore) Close() error       { return s.db.Close() }
func (s *SQLiteStore) DB() *sql.DB        { return s.db }
func (s *SQLiteStore) DataSource() string { return s.dataSource }
func (s *SQLiteStore) get(ctx context.Context, kind string, dst any) error {
	var payload []byte
	if err := s.db.QueryRowContext(ctx, "SELECT payload FROM dashboard_data WHERE kind=?", kind).Scan(&payload); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return json.Unmarshal([]byte(emptyValue(kind)), dst)
		}
		return err
	}
	return json.Unmarshal(payload, dst)
}
func emptyValue(kind string) string {
	if kind == "summary" {
		return `{"protectionStatus":"offline","health":"unknown"}`
	}
	if kind == "protection" {
		return `{"status":"offline","rules":[]}`
	}
	return `[]`
}
func (s *SQLiteStore) GetSummary(ctx context.Context) (*models.SystemSummary, error) {
	var v models.SystemSummary
	if err := s.get(ctx, "summary", &v); err != nil {
		return nil, err
	}
	if s.hasTable(ctx, "app_instances") {
		instances, err := s.ListAppInstances(ctx)
		if err != nil {
			return nil, err
		}
		if len(instances) > 0 {
			running := 0
			for _, inst := range instances {
				if inst.Status == models.AppInstanceRunning {
					running++
				}
			}
			v.AppsTotal = len(instances)
			v.AppsRunning = running
		}
	}
	return &v, nil
}
func (s *SQLiteStore) ListAttention(ctx context.Context) ([]models.AttentionItem, error) {
	var v []models.AttentionItem
	return v, s.get(ctx, "attention", &v)
}
func (s *SQLiteStore) ListApps(ctx context.Context) ([]models.InstalledApp, error) {
	if !s.hasTable(ctx, "app_instances") {
		var v []models.InstalledApp
		return v, s.get(ctx, "apps", &v)
	}
	instances, err := s.ListAppInstances(ctx)
	if err != nil {
		return nil, err
	}
	if len(instances) == 0 {
		var v []models.InstalledApp
		return v, s.get(ctx, "apps", &v)
	}
	apps := make([]models.InstalledApp, 0, len(instances))
	for _, inst := range instances {
		apps = append(apps, s.instanceToInstalledApp(ctx, inst))
	}
	return apps, nil
}
func (s *SQLiteStore) GetApp(ctx context.Context, id string) (*models.InstalledApp, error) {
	if !s.hasTable(ctx, "app_instances") {
		v, err := s.ListApps(ctx)
		if err != nil {
			return nil, err
		}
		for i := range v {
			if v[i].ID == id {
				return &v[i], nil
			}
		}
		return nil, fmt.Errorf("app not found: %s", id)
	}
	inst, err := s.GetAppInstance(ctx, id)
	if err != nil {
		v, err2 := s.ListApps(ctx)
		if err2 != nil {
			return nil, err2
		}
		for i := range v {
			if v[i].ID == id {
				return &v[i], nil
			}
		}
		return nil, fmt.Errorf("app not found: %s", id)
	}
	app := s.instanceToInstalledApp(ctx, *inst)
	return &app, nil
}

func (s *SQLiteStore) instanceToInstalledApp(ctx context.Context, inst models.AppInstance) models.InstalledApp {
	trust := models.SourceTrust("")
	if inst.SourceID != "" {
		src, err := s.GetManifestSource(ctx, inst.SourceID)
		if err == nil {
			trust = src.Trust
		}
	}
	return models.InstalledApp{
		ID:            inst.ID,
		Name:          inst.Name,
		Icon:          "",
		Description:   "",
		Status:        models.AppStatus(inst.Status),
		Health:        inst.Health,
		Version:       inst.Version,
		LatestVersion: inst.Version,
		Trust:         trust,
		Category:      "",
		Endpoints:     inst.Endpoints,
		Services:      nil,
		Storage:       inst.Storage,
		UpdatedAt:     inst.UpdatedAt,
		InstalledAt:   inst.CreatedAt,
	}
}

func (s *SQLiteStore) hasTable(ctx context.Context, name string) bool {
	var count int
	_ = s.db.QueryRowContext(ctx, "SELECT count(*) FROM sqlite_master WHERE type='table' AND name=?", name).Scan(&count)
	return count > 0
}
func (s *SQLiteStore) ListCatalog(ctx context.Context) ([]models.CatalogApp, error) {
	var v []models.CatalogApp
	return v, s.get(ctx, "catalog", &v)
}
func (s *SQLiteStore) GetProtection(ctx context.Context) (*models.ProtectionOverview, error) {
	var v models.ProtectionOverview
	return &v, s.get(ctx, "protection", &v)
}
func (s *SQLiteStore) ListActivity(ctx context.Context) ([]models.ActivityEvent, error) {
	var v []models.ActivityEvent
	return v, s.get(ctx, "activity", &v)
}

func (s *SQLiteStore) CreateAppInstance(ctx context.Context, instance *models.AppInstance) error {
	endpoints, err := json.Marshal(instance.Endpoints)
	if err != nil {
		return err
	}
	storage, err := json.Marshal(instance.Storage)
	if err != nil {
		return err
	}
	now := time.Now().Unix()
	srcID := sql.NullString{String: instance.SourceID, Valid: instance.SourceID != ""}
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO app_instances(id,catalog_id,source_id,name,version,project_name,install_path,compose_path,status,health,endpoints_json,storage_json,created_at,updated_at)
		 VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		instance.ID, instance.CatalogID, srcID, instance.Name, instance.Version, instance.ProjectName, instance.InstallPath, instance.ComposePath, instance.Status, instance.Health, endpoints, storage, now, now)
	return err
}

func (s *SQLiteStore) GetAppInstance(ctx context.Context, id string) (*models.AppInstance, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id,catalog_id,source_id,name,version,project_name,install_path,compose_path,status,health,endpoints_json,storage_json,created_at,updated_at FROM app_instances WHERE id=?`, id)
	return s.scanAppInstance(row)
}

func (s *SQLiteStore) GetAppInstanceByCatalogID(ctx context.Context, catalogID string) (*models.AppInstance, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id,catalog_id,source_id,name,version,project_name,install_path,compose_path,status,health,endpoints_json,storage_json,created_at,updated_at FROM app_instances WHERE catalog_id=?`, catalogID)
	return s.scanAppInstance(row)
}

func (s *SQLiteStore) ListAppInstances(ctx context.Context) ([]models.AppInstance, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id,catalog_id,source_id,name,version,project_name,install_path,compose_path,status,health,endpoints_json,storage_json,created_at,updated_at FROM app_instances ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var instances []models.AppInstance
	for rows.Next() {
		inst, err := s.scanAppInstance(rows)
		if err != nil {
			return nil, err
		}
		instances = append(instances, *inst)
	}
	return instances, rows.Err()
}

func (s *SQLiteStore) scanAppInstance(scanner interface{ Scan(dest ...any) error }) (*models.AppInstance, error) {
	var inst models.AppInstance
	var sourceID sql.NullString
	var endpointsJSON, storageJSON []byte
	var createdAt, updatedAt int64
	err := scanner.Scan(&inst.ID, &inst.CatalogID, &sourceID, &inst.Name, &inst.Version, &inst.ProjectName, &inst.InstallPath, &inst.ComposePath, &inst.Status, &inst.Health, &endpointsJSON, &storageJSON, &createdAt, &updatedAt)
	if sourceID.Valid {
		inst.SourceID = sourceID.String
	}
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("app instance not found: %s", inst.ID)
		}
		return nil, err
	}
	if err := json.Unmarshal(endpointsJSON, &inst.Endpoints); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(storageJSON, &inst.Storage); err != nil {
		return nil, err
	}
	inst.CreatedAt = time.Unix(createdAt, 0).UTC()
	inst.UpdatedAt = time.Unix(updatedAt, 0).UTC()
	return &inst, nil
}

func (s *SQLiteStore) UpdateAppInstanceStatus(ctx context.Context, id string, status models.AppInstanceStatus, health models.HealthStatus) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE app_instances SET status=?, health=?, updated_at=? WHERE id=?`,
		status, health, time.Now().Unix(), id)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("app instance not found: %s", id)
	}
	return nil
}

func (s *SQLiteStore) UpdateAppInstanceStatusAndEndpoints(ctx context.Context, id string, status models.AppInstanceStatus, health models.HealthStatus, endpoints []models.AppEndpoint) error {
	endpointsJSON, err := json.Marshal(endpoints)
	if err != nil {
		return err
	}
	res, err := s.db.ExecContext(ctx,
		`UPDATE app_instances SET status=?, health=?, endpoints_json=?, updated_at=? WHERE id=?`,
		status, health, endpointsJSON, time.Now().Unix(), id)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("app instance not found: %s", id)
	}
	return nil
}

func (s *SQLiteStore) DeleteAppInstance(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM app_instances WHERE id=?`, id)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("app instance not found: %s", id)
	}
	return nil
}

func (s *SQLiteStore) CreateOperation(ctx context.Context, op *models.Operation) error {
	now := time.Now().Unix()
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO operations(id,app_id,kind,status,error,output,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?)`,
		op.ID, op.AppID, op.Kind, op.Status, op.Error, op.Output, now, now)
	return err
}

func (s *SQLiteStore) GetOperation(ctx context.Context, id string) (*models.Operation, error) {
	var op models.Operation
	var createdAt, updatedAt int64
	err := s.db.QueryRowContext(ctx,
		`SELECT id,app_id,kind,status,error,output,created_at,updated_at FROM operations WHERE id=?`, id).
		Scan(&op.ID, &op.AppID, &op.Kind, &op.Status, &op.Error, &op.Output, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("operation not found: %s", id)
		}
		return nil, err
	}
	op.CreatedAt = time.Unix(createdAt, 0).UTC()
	op.UpdatedAt = time.Unix(updatedAt, 0).UTC()
	return &op, nil
}

func (s *SQLiteStore) GetActiveOperationForApp(ctx context.Context, appID string) (*models.Operation, error) {
	var op models.Operation
	var createdAt, updatedAt int64
	err := s.db.QueryRowContext(ctx,
		`SELECT id,app_id,kind,status,error,output,created_at,updated_at FROM operations WHERE app_id=? AND status IN ('pending','running') ORDER BY created_at DESC LIMIT 1`, appID).
		Scan(&op.ID, &op.AppID, &op.Kind, &op.Status, &op.Error, &op.Output, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	op.CreatedAt = time.Unix(createdAt, 0).UTC()
	op.UpdatedAt = time.Unix(updatedAt, 0).UTC()
	return &op, nil
}

func (s *SQLiteStore) UpdateOperation(ctx context.Context, op *models.Operation) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE operations SET status=?, error=?, output=?, updated_at=? WHERE id=?`,
		op.Status, op.Error, op.Output, time.Now().Unix(), op.ID)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("operation not found: %s", op.ID)
	}
	return nil
}

func (s *SQLiteStore) ListOperationsForApp(ctx context.Context, appID string) ([]models.Operation, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id,app_id,kind,status,error,output,created_at,updated_at FROM operations WHERE app_id=? ORDER BY created_at DESC`, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ops []models.Operation
	for rows.Next() {
		op, err := s.scanOperation(rows)
		if err != nil {
			return nil, err
		}
		ops = append(ops, *op)
	}
	return ops, rows.Err()
}

func (s *SQLiteStore) scanOperation(scanner interface{ Scan(dest ...any) error }) (*models.Operation, error) {
	var op models.Operation
	var createdAt, updatedAt int64
	err := scanner.Scan(&op.ID, &op.AppID, &op.Kind, &op.Status, &op.Error, &op.Output, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	op.CreatedAt = time.Unix(createdAt, 0).UTC()
	op.UpdatedAt = time.Unix(updatedAt, 0).UTC()
	return &op, nil
}

// RecoverOperations marks any operations left pending or running after a crash
// as failed so that subsequent operations on the same app are not blocked.
func (s *SQLiteStore) RecoverOperations(ctx context.Context) error {
	now := time.Now().Unix()
	_, err := s.db.ExecContext(ctx,
		`UPDATE operations SET status=?, error=?, output=?, updated_at=? WHERE status IN (?,?)`,
		models.OperationFailed, "interrupted by restart", "", now,
		models.OperationPending, models.OperationRunning)
	return err
}

func (s *SQLiteStore) CreateRevision(ctx context.Context, revision *models.Revision) error {
	manifest, err := json.Marshal(revision.Manifest)
	if err != nil {
		return err
	}
	now := time.Now().Unix()
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO revisions(id,app_id,manifest_json,config_json,compose_path,created_at) VALUES(?,?,?,?,?,?)`,
		revision.ID, revision.AppID, manifest, revision.ConfigJSON, revision.ComposePath, now)
	return err
}

func (s *SQLiteStore) ApplySuccessfulUpdate(ctx context.Context, instance *models.AppInstance, revision *models.Revision, source *models.ManifestSource, expectedCommit, expectedChecksum string) error {
	manifest, err := json.Marshal(revision.Manifest)
	if err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	res, err := tx.ExecContext(ctx, `UPDATE manifest_sources SET commit_sha=?,manifest_json=?,checksum=?,updated_at=? WHERE id=? AND commit_sha=? AND checksum=?`, source.CommitSHA, source.ManifestJSON, source.Checksum, time.Now().Unix(), source.ID, expectedCommit, expectedChecksum)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n != 1 {
		return fmt.Errorf("source changed while update was running")
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO revisions(id,app_id,manifest_json,config_json,compose_path,created_at) VALUES(?,?,?,?,?,?)`, revision.ID, revision.AppID, manifest, revision.ConfigJSON, revision.ComposePath, time.Now().Unix())
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `UPDATE app_instances SET name=?,version=?,compose_path=?,status=?,health=?,endpoints_json=?,storage_json=?,updated_at=? WHERE id=?`, instance.Name, instance.Version, instance.ComposePath, instance.Status, instance.Health, mustStoreJSON(instance.Endpoints), mustStoreJSON(instance.Storage), time.Now().Unix(), instance.ID)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func mustStoreJSON(v any) []byte { b, _ := json.Marshal(v); return b }

func (s *SQLiteStore) GetLatestRevision(ctx context.Context, appID string) (*models.Revision, error) {
	var rev models.Revision
	var manifestJSON, configJSON []byte
	var createdAt int64
	err := s.db.QueryRowContext(ctx,
		`SELECT id,app_id,manifest_json,config_json,compose_path,created_at FROM revisions WHERE app_id=? ORDER BY created_at DESC LIMIT 1`, appID).
		Scan(&rev.ID, &rev.AppID, &manifestJSON, &configJSON, &rev.ComposePath, &createdAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("revision not found: %s", appID)
		}
		return nil, err
	}
	if err := json.Unmarshal(manifestJSON, &rev.Manifest); err != nil {
		return nil, err
	}
	rev.ConfigJSON = string(configJSON)
	rev.CreatedAt = time.Unix(createdAt, 0).UTC()
	return &rev, nil
}

func (s *SQLiteStore) CreateManifestSource(ctx context.Context, source *models.ManifestSource) error {
	now := time.Now().Unix()
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO manifest_sources(id,source_url,owner,repo,path,ref,commit_sha,manifest_json,checksum,trust,created_at,updated_at)
		 VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`,
		source.ID, source.SourceURL, source.Owner, source.Repo, source.Path, source.Ref, source.CommitSHA, source.ManifestJSON, source.Checksum, source.Trust, now, now)
	return err
}

func (s *SQLiteStore) GetManifestSource(ctx context.Context, id string) (*models.ManifestSource, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id,source_url,owner,repo,path,ref,commit_sha,manifest_json,checksum,trust,created_at,updated_at FROM manifest_sources WHERE id=?`, id)
	return s.scanManifestSource(row)
}

func (s *SQLiteStore) GetManifestSourceByURL(ctx context.Context, url string) (*models.ManifestSource, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id,source_url,owner,repo,path,ref,commit_sha,manifest_json,checksum,trust,created_at,updated_at FROM manifest_sources WHERE source_url=?`, url)
	return s.scanManifestSource(row)
}

func (s *SQLiteStore) ListManifestSources(ctx context.Context) ([]models.ManifestSource, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id,source_url,owner,repo,path,ref,commit_sha,manifest_json,checksum,trust,created_at,updated_at FROM manifest_sources ORDER BY updated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var sources []models.ManifestSource
	for rows.Next() {
		src, err := s.scanManifestSource(rows)
		if err != nil {
			return nil, err
		}
		sources = append(sources, *src)
	}
	return sources, rows.Err()
}

func (s *SQLiteStore) DeleteManifestSource(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM manifest_sources WHERE id=?`, id)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("manifest source not found: %s", id)
	}
	return nil
}

func (s *SQLiteStore) UpdateManifestSourceTrust(ctx context.Context, id string, trust models.SourceTrust) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE manifest_sources SET trust=?, updated_at=? WHERE id=?`,
		trust, time.Now().Unix(), id)
	return err
}

func (s *SQLiteStore) UpdateManifestSourceManifest(ctx context.Context, source *models.ManifestSource) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE manifest_sources SET commit_sha=?, manifest_json=?, checksum=?, trust=?, updated_at=? WHERE id=?`,
		source.CommitSHA, source.ManifestJSON, source.Checksum, source.Trust, time.Now().Unix(), source.ID)
	return err
}

func (s *SQLiteStore) scanManifestSource(scanner interface{ Scan(dest ...any) error }) (*models.ManifestSource, error) {
	var src models.ManifestSource
	var createdAt, updatedAt int64
	err := scanner.Scan(&src.ID, &src.SourceURL, &src.Owner, &src.Repo, &src.Path, &src.Ref, &src.CommitSHA, &src.ManifestJSON, &src.Checksum, &src.Trust, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("manifest source not found")
		}
		return nil, err
	}
	src.CreatedAt = time.Unix(createdAt, 0).UTC()
	src.UpdatedAt = time.Unix(updatedAt, 0).UTC()
	return &src, nil
}

var _ AppManager = (*SQLiteStore)(nil)
var _ SourceManager = (*SQLiteStore)(nil)
