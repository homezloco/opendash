package store

import (
	"context"
	"fmt"

	"github.com/opendash-project/opendash/internal/data"
	"github.com/opendash-project/opendash/internal/diagnostics"
	"github.com/opendash-project/opendash/internal/models"
)

type Store interface {
	GetSummary(ctx context.Context) (*models.SystemSummary, error)
	ListAttention(ctx context.Context) ([]models.AttentionItem, error)
	ListApps(ctx context.Context) ([]models.InstalledApp, error)
	GetApp(ctx context.Context, id string) (*models.InstalledApp, error)
	ListCatalog(ctx context.Context) ([]models.CatalogApp, error)
	GetProtection(ctx context.Context) (*models.ProtectionOverview, error)
	ListActivity(ctx context.Context) ([]models.ActivityEvent, error)
}

type AppManager interface {
	CreateAppInstance(ctx context.Context, instance *models.AppInstance) error
	GetAppInstance(ctx context.Context, id string) (*models.AppInstance, error)
	GetAppInstanceByCatalogID(ctx context.Context, catalogID string) (*models.AppInstance, error)
	ListAppInstances(ctx context.Context) ([]models.AppInstance, error)
	UpdateAppInstanceStatus(ctx context.Context, id string, status models.AppInstanceStatus, health models.HealthStatus) error
	UpdateAppInstanceStatusAndEndpoints(ctx context.Context, id string, status models.AppInstanceStatus, health models.HealthStatus, endpoints []models.AppEndpoint) error
	DeleteAppInstance(ctx context.Context, id string) error

	CreateOperation(ctx context.Context, op *models.Operation) error
	GetOperation(ctx context.Context, id string) (*models.Operation, error)
	GetActiveOperationForApp(ctx context.Context, appID string) (*models.Operation, error)
	UpdateOperation(ctx context.Context, op *models.Operation) error
	ListOperationsForApp(ctx context.Context, appID string) ([]models.Operation, error)
	RecoverOperations(ctx context.Context) error

	CreateRevision(ctx context.Context, revision *models.Revision) error
	GetLatestRevision(ctx context.Context, appID string) (*models.Revision, error)
	ApplySuccessfulUpdate(ctx context.Context, instance *models.AppInstance, revision *models.Revision, source *models.ManifestSource, expectedCommit, expectedChecksum string) error
}

type SourceManager interface {
	CreateManifestSource(ctx context.Context, source *models.ManifestSource) error
	GetManifestSource(ctx context.Context, id string) (*models.ManifestSource, error)
	GetManifestSourceByURL(ctx context.Context, url string) (*models.ManifestSource, error)
	ListManifestSources(ctx context.Context) ([]models.ManifestSource, error)
	DeleteManifestSource(ctx context.Context, id string) error
	UpdateManifestSourceTrust(ctx context.Context, id string, trust models.SourceTrust) error
	UpdateManifestSourceManifest(ctx context.Context, source *models.ManifestSource) error
}

type MemoryStore struct{}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{}
}

func (s *MemoryStore) GetSummary(ctx context.Context) (*models.SystemSummary, error) {
	apps := data.Apps
	protection := data.Protection
	summary := *data.Summary
	summary.AppsRunning = 0
	summary.AppsTotal = len(apps)
	summary.UpdatesAvailable = 0
	for _, app := range apps {
		if app.Status == models.AppRunning {
			summary.AppsRunning++
		}
		if app.Version != app.LatestVersion {
			summary.UpdatesAvailable++
		}
	}
	summary.Health = diagnostics.EvaluateHealth(apps, protection)
	summary.ProtectionStatus = protection.Status
	return &summary, nil
}

func (s *MemoryStore) ListAttention(ctx context.Context) ([]models.AttentionItem, error) {
	return data.AttentionItems, nil
}

func (s *MemoryStore) ListApps(ctx context.Context) ([]models.InstalledApp, error) {
	return data.Apps, nil
}

func (s *MemoryStore) GetApp(ctx context.Context, id string) (*models.InstalledApp, error) {
	for i := range data.Apps {
		if data.Apps[i].ID == id {
			app := data.Apps[i]
			return &app, nil
		}
	}
	return nil, fmt.Errorf("app not found: %s", id)
}

func (s *MemoryStore) ListCatalog(ctx context.Context) ([]models.CatalogApp, error) {
	return data.CatalogApps, nil
}

func (s *MemoryStore) GetProtection(ctx context.Context) (*models.ProtectionOverview, error) {
	return data.Protection, nil
}

func (s *MemoryStore) ListActivity(ctx context.Context) ([]models.ActivityEvent, error) {
	return data.ActivityEvents, nil
}

func (s *MemoryStore) CreateAppInstance(ctx context.Context, instance *models.AppInstance) error {
	return fmt.Errorf("not implemented in memory store")
}

func (s *MemoryStore) GetAppInstance(ctx context.Context, id string) (*models.AppInstance, error) {
	return nil, fmt.Errorf("app instance not found: %s", id)
}

func (s *MemoryStore) GetAppInstanceByCatalogID(ctx context.Context, catalogID string) (*models.AppInstance, error) {
	return nil, fmt.Errorf("app instance not found: %s", catalogID)
}

func (s *MemoryStore) ListAppInstances(ctx context.Context) ([]models.AppInstance, error) {
	return nil, nil
}

func (s *MemoryStore) UpdateAppInstanceStatus(ctx context.Context, id string, status models.AppInstanceStatus, health models.HealthStatus) error {
	return fmt.Errorf("not implemented in memory store")
}

func (s *MemoryStore) UpdateAppInstanceStatusAndEndpoints(ctx context.Context, id string, status models.AppInstanceStatus, health models.HealthStatus, endpoints []models.AppEndpoint) error {
	return fmt.Errorf("not implemented in memory store")
}

func (s *MemoryStore) DeleteAppInstance(ctx context.Context, id string) error {
	return nil
}

func (s *MemoryStore) CreateOperation(ctx context.Context, op *models.Operation) error {
	return fmt.Errorf("not implemented in memory store")
}

func (s *MemoryStore) GetOperation(ctx context.Context, id string) (*models.Operation, error) {
	return nil, fmt.Errorf("operation not found: %s", id)
}

func (s *MemoryStore) GetActiveOperationForApp(ctx context.Context, appID string) (*models.Operation, error) {
	return nil, nil
}

func (s *MemoryStore) UpdateOperation(ctx context.Context, op *models.Operation) error {
	return fmt.Errorf("not implemented in memory store")
}

func (s *MemoryStore) ListOperationsForApp(ctx context.Context, appID string) ([]models.Operation, error) {
	return nil, nil
}

func (s *MemoryStore) RecoverOperations(ctx context.Context) error {
	return nil
}

func (s *MemoryStore) CreateRevision(ctx context.Context, revision *models.Revision) error {
	return fmt.Errorf("not implemented in memory store")
}

func (s *MemoryStore) GetLatestRevision(ctx context.Context, appID string) (*models.Revision, error) {
	return nil, fmt.Errorf("revision not found: %s", appID)
}

func (s *MemoryStore) ApplySuccessfulUpdate(ctx context.Context, instance *models.AppInstance, revision *models.Revision, source *models.ManifestSource, expectedCommit, expectedChecksum string) error {
	return fmt.Errorf("not implemented in memory store")
}

func (s *MemoryStore) CreateManifestSource(ctx context.Context, source *models.ManifestSource) error {
	return fmt.Errorf("not implemented in memory store")
}

func (s *MemoryStore) GetManifestSource(ctx context.Context, id string) (*models.ManifestSource, error) {
	return nil, fmt.Errorf("source not found: %s", id)
}

func (s *MemoryStore) GetManifestSourceByURL(ctx context.Context, url string) (*models.ManifestSource, error) {
	return nil, fmt.Errorf("source not found")
}

func (s *MemoryStore) ListManifestSources(ctx context.Context) ([]models.ManifestSource, error) {
	return nil, nil
}

func (s *MemoryStore) DeleteManifestSource(ctx context.Context, id string) error {
	return fmt.Errorf("not implemented in memory store")
}

func (s *MemoryStore) UpdateManifestSourceTrust(ctx context.Context, id string, trust models.SourceTrust) error {
	return fmt.Errorf("not implemented in memory store")
}

func (s *MemoryStore) UpdateManifestSourceManifest(ctx context.Context, source *models.ManifestSource) error {
	return fmt.Errorf("not implemented in memory store")
}
