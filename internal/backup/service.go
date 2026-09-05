package backup

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/opendash-project/opendash/internal/models"
	"github.com/opendash-project/opendash/internal/store"
)

var safeVolume = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]{0,127}$`)

type VolumeArchiver interface {
	ArchiveVolume(context.Context, string, string) error
}
type DrillVerifier interface {
	Verify(context.Context, string, models.BackupMetadata) (bool, error)
}
type Service struct {
	store    store.BackupManager
	apps     store.AppManager
	sources  store.SourceManager
	root     string
	archiver VolumeArchiver
	verifier DrillVerifier
	now      func() time.Time
}
type CreateRequest struct {
	RetentionCount int `json:"retentionCount"`
	RetentionDays  int `json:"retentionDays"`
}

func NewService(st store.BackupManager, apps store.AppManager, sources store.SourceManager, root string, archiver VolumeArchiver, verifier DrillVerifier) *Service {
	if root == "" {
		root = "./data/backups"
	}
	return &Service{store: st, apps: apps, sources: sources, root: root, archiver: archiver, verifier: verifier, now: time.Now}
}
func ID() string { b := make([]byte, 12); _, _ = rand.Read(b); return hex.EncodeToString(b) }
func Plan(manifest models.Manifest) ([]string, error) {
	if manifest.Backup == nil || len(manifest.Backup.Volumes) == 0 {
		return nil, errors.New("manifest does not declare backup volumes")
	}
	out := append([]string(nil), manifest.Backup.Volumes...)
	for _, v := range out {
		if !safeVolume.MatchString(v) {
			return nil, fmt.Errorf("unsafe volume name %q", v)
		}
	}
	sort.Strings(out)
	return out, nil
}
func (s *Service) Create(ctx context.Context, appID string, req CreateRequest) (*models.BackupArchive, error) {
	app, err := s.apps.GetAppInstance(ctx, appID)
	if err != nil {
		return nil, err
	}
	rev, err := s.apps.GetLatestRevision(ctx, appID)
	if err != nil {
		return nil, fmt.Errorf("latest manifest: %w", err)
	}
	volumes, err := Plan(rev.Manifest)
	if err != nil {
		return nil, err
	}
	now := s.now().UTC()
	job := &models.BackupJob{ID: ID(), AppID: appID, Status: models.BackupRunning, CreatedAt: now, UpdatedAt: now}
	if err = s.store.CreateBackupJob(ctx, job); err != nil {
		return nil, err
	}
	fail := func(e error) (*models.BackupArchive, error) {
		job.Status = models.BackupFailed
		job.Error = e.Error()
		job.UpdatedAt = s.now().UTC()
		_ = s.store.UpdateBackupJob(ctx, job)
		return nil, e
	}
	if err = s.EnforceRetention(ctx, appID, req.RetentionCount, req.RetentionDays); err != nil {
		return fail(err)
	}
	dir := filepath.Join(s.root, appID)
	if err = os.MkdirAll(dir, 0700); err != nil {
		return fail(err)
	}
	stage, err := os.MkdirTemp(dir, ".creating-")
	if err != nil {
		return fail(err)
	}
	defer os.RemoveAll(stage)
	for _, v := range volumes {
		if s.archiver == nil {
			return fail(errors.New("volume archiver is unavailable"))
		}
		if err = s.archiver.ArchiveVolume(ctx, v, filepath.Join(stage, "volume-"+v+".tar")); err != nil {
			return fail(fmt.Errorf("archive volume %s: %w", v, err))
		}
	}
	compose := ""
	if rev.Manifest.Backup != nil && rev.Manifest.Backup.IncludeConfig {
		compose = readOptional(app.ComposePath)
	}
	meta := models.BackupMetadata{Version: 1, App: *app, Manifest: rev.Manifest, Compose: compose, Volumes: volumes, Images: images(rev.Manifest), CreatedAt: now}
	if rev.Manifest.Backup != nil {
		meta.MigrationRisk = rev.Manifest.Backup.MigrationRisk
		meta.RollbackSafe = rev.Manifest.Backup.RollbackSafe
	}
	if app.SourceID != "" && s.sources != nil {
		meta.Source, _ = s.sources.GetManifestSource(ctx, app.SourceID)
	}
	payload, _ := json.MarshalIndent(meta, "", "  ")
	if err = os.WriteFile(filepath.Join(stage, "metadata.json"), payload, 0600); err != nil {
		return fail(err)
	}
	path := filepath.Join(dir, now.Format("20060102T150405Z")+"-"+job.ID+".tar.gz")
	if err = pack(stage, path); err != nil {
		return fail(err)
	}
	sum, size, err := fileHash(path)
	if err != nil {
		return fail(err)
	}
	mh := sha256.Sum256(payload)
	arc := &models.BackupArchive{ID: ID(), JobID: job.ID, AppID: appID, Path: path, Size: size, SHA256: sum, ManifestSHA256: hex.EncodeToString(mh[:]), CreatedAt: now}
	if err = s.store.CreateBackupArchive(ctx, arc); err != nil {
		os.Remove(path)
		return fail(err)
	}
	job.Status = models.BackupCompleted
	job.UpdatedAt = s.now().UTC()
	_ = s.store.UpdateBackupJob(ctx, job)
	_ = s.EnforceRetention(ctx, appID, req.RetentionCount, req.RetentionDays)
	return arc, nil
}
func (s *Service) Preview(ctx context.Context, id string) (*models.RestorePreview, error) {
	arc, err := s.store.GetBackupArchive(ctx, id)
	if err != nil {
		return nil, err
	}
	sum, _, err := fileHash(arc.Path)
	if err != nil {
		return nil, err
	}
	meta, err := readMetadata(arc.Path)
	if err != nil {
		return nil, err
	}
	trusted := meta.Source == nil || sourceTrusted(meta.Source.Trust)
	if meta.Source != nil && s.sources != nil {
		current, e := s.sources.GetManifestSource(ctx, meta.Source.ID)
		trusted = e == nil && sourceTrusted(current.Trust) && current.Checksum == meta.Source.Checksum
	}
	compatible := meta.Version == 1
	reasons := []string{}
	if !compatible {
		reasons = append(reasons, "unsupported backup format")
	}
	return &models.RestorePreview{BackupID: id, AppID: arc.AppID, Volumes: meta.Volumes, Images: meta.Images, Manifest: meta.Manifest, Compatible: compatible, Compatibility: reasons, MigrationRisk: meta.MigrationRisk, RollbackSafe: meta.RollbackSafe, SourceTrusted: trusted, IntegrityOK: sum == arc.SHA256}, nil
}
func (s *Service) Verify(ctx context.Context, id string) (*models.VerificationResult, error) {
	arc, err := s.store.GetBackupArchive(ctx, id)
	if err != nil {
		return nil, err
	}
	meta, err := readMetadata(arc.Path)
	if err != nil {
		return nil, err
	}
	project := "opendash-verify-" + id[:min(12, len(id))]
	result := &models.VerificationResult{BackupID: id, ProjectName: project, VerifiedAt: s.now().UTC(), CleanedUp: true}
	if s.verifier == nil {
		return nil, errors.New("isolated verification is unavailable")
	}
	result.Healthy, err = s.verifier.Verify(ctx, project, meta)
	if err != nil {
		result.Error = err.Error()
	}
	_ = s.store.MarkBackupVerified(ctx, id, result.VerifiedAt, result.Healthy)
	return result, err
}
func (s *Service) EnforceRetention(ctx context.Context, appID string, count, days int) error {
	arcs, err := s.store.ListBackupArchives(ctx, appID)
	if err != nil {
		return err
	}
	cut := time.Time{}
	if days > 0 {
		cut = s.now().AddDate(0, 0, -days)
	}
	for i, a := range arcs {
		if (count > 0 && i >= count) || (!cut.IsZero() && a.CreatedAt.Before(cut)) {
			if err = os.Remove(a.Path); err != nil && !errors.Is(err, os.ErrNotExist) {
				return err
			}
			if err = s.store.DeleteBackupArchive(ctx, a.ID); err != nil {
				return err
			}
		}
	}
	return nil
}
func images(m models.Manifest) []string {
	var out []string
	if m.Compose.Inline != nil {
		for _, s := range m.Compose.Inline.Services {
			out = append(out, s.Image)
		}
	}
	sort.Strings(out)
	return out
}
func readOptional(path string) string {
	b, e := os.ReadFile(path)
	if e != nil {
		return ""
	}
	return string(b)
}
func pack(dir, dst string) error {
	f, e := os.OpenFile(dst, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		return e
	}
	defer f.Close()
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)
	e = filepath.Walk(dir, func(path string, info os.FileInfo, e error) error {
		if e != nil {
			return e
		}
		if info.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(dir, path)
		h, e := tar.FileInfoHeader(info, "")
		if e != nil {
			return e
		}
		h.Name = filepath.ToSlash(rel)
		if e = tw.WriteHeader(h); e != nil {
			return e
		}
		in, e := os.Open(path)
		if e != nil {
			return e
		}
		defer in.Close()
		_, e = io.Copy(tw, in)
		return e
	})
	if x := tw.Close(); e == nil {
		e = x
	}
	if x := gz.Close(); e == nil {
		e = x
	}
	return e
}
func readMetadata(path string) (models.BackupMetadata, error) {
	var m models.BackupMetadata
	f, e := os.Open(path)
	if e != nil {
		return m, e
	}
	defer f.Close()
	gz, e := gzip.NewReader(f)
	if e != nil {
		return m, e
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		h, e := tr.Next()
		if errors.Is(e, io.EOF) {
			break
		}
		if e != nil {
			return m, e
		}
		if h.Name == "metadata.json" {
			e = json.NewDecoder(io.LimitReader(tr, 4<<20)).Decode(&m)
			return m, e
		}
	}
	return m, errors.New("backup metadata missing")
}
func fileHash(path string) (string, int64, error) {
	f, e := os.Open(path)
	if e != nil {
		return "", 0, e
	}
	defer f.Close()
	h := sha256.New()
	n, e := io.Copy(h, f)
	return hex.EncodeToString(h.Sum(nil)), n, e
}
func sourceTrusted(v models.SourceTrust) bool {
	return v == models.SourceOfficial || v == models.SourceReviewed || v == models.SourceUserTrusted
}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

var _ = strings.Builder{}
