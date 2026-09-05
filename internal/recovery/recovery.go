package recovery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/opendash-project/opendash/internal/models"
	"github.com/opendash-project/opendash/internal/store"
)

type Inventory struct {
	Version    int                     `json:"version"`
	ExportedAt time.Time               `json:"exportedAt"`
	Apps       []models.AppInstance    `json:"apps"`
	Sources    []models.ManifestSource `json:"sources"`
	Backups    []models.BackupArchive  `json:"backups"`
}
type ImportConflict struct {
	AppID   string `json:"appId"`
	Kind    string `json:"kind"`
	Message string `json:"message"`
}
type ImportPreview struct {
	Valid            bool             `json:"valid"`
	Apps             int              `json:"apps"`
	Sources          int              `json:"sources"`
	Backups          int              `json:"backups"`
	UntrustedSources []string         `json:"untrustedSources"`
	Conflicts        []ImportConflict `json:"conflicts"`
	Strategy         string           `json:"strategy"`
}
type Service struct {
	apps    store.AppManager
	sources store.SourceManager
	backups store.BackupManager
	now     func() time.Time
}

func NewService(a store.AppManager, s store.SourceManager, b store.BackupManager) *Service {
	return &Service{apps: a, sources: s, backups: b, now: time.Now}
}
func (s *Service) Export(ctx context.Context) (*Inventory, error) {
	apps, e := s.apps.ListAppInstances(ctx)
	if e != nil {
		return nil, e
	}
	sources, e := s.sources.ListManifestSources(ctx)
	if e != nil {
		return nil, e
	}
	inv := &Inventory{Version: 1, ExportedAt: s.now().UTC(), Apps: apps, Sources: sources, Backups: []models.BackupArchive{}}
	for _, a := range apps {
		v, e := s.backups.ListBackupArchives(ctx, a.ID)
		if e != nil {
			return nil, e
		}
		inv.Backups = append(inv.Backups, v...)
	}
	return inv, nil
}
func (s *Service) Preview(ctx context.Context, payload []byte) (*ImportPreview, error) {
	if len(payload) > 8<<20 {
		return nil, errors.New("inventory exceeds 8 MiB")
	}
	var inv Inventory
	if e := json.Unmarshal(payload, &inv); e != nil {
		return nil, fmt.Errorf("invalid inventory: %w", e)
	}
	if inv.Version != 1 {
		return nil, errors.New("unsupported inventory version")
	}
	p := &ImportPreview{Valid: true, Apps: len(inv.Apps), Sources: len(inv.Sources), Backups: len(inv.Backups), Strategy: "new-ids-or-reconcile", Conflicts: []ImportConflict{}, UntrustedSources: []string{}}
	for _, src := range inv.Sources {
		if src.Trust != models.SourceOfficial && src.Trust != models.SourceReviewed && src.Trust != models.SourceUserTrusted {
			p.UntrustedSources = append(p.UntrustedSources, src.SourceURL)
		}
	}
	for _, a := range inv.Apps {
		if existing, e := s.apps.GetAppInstance(ctx, a.ID); e == nil {
			p.Conflicts = append(p.Conflicts, ImportConflict{AppID: a.ID, Kind: "app-id", Message: "existing app " + existing.Name + " requires reconciliation or a new ID"})
		}
	}
	return p, nil
}

// Apply intentionally requires an explicit strategy and imports metadata only;
// it never runs source or manifest hooks.
func (s *Service) Apply(context.Context, Inventory, string) error {
	return errors.New("recovery apply is not enabled")
}
