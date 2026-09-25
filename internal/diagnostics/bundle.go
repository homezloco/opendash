package diagnostics

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"runtime/debug"
	"sort"
	"time"

	"github.com/opendash-project/opendash/internal/models"
	opdruntime "github.com/opendash-project/opendash/internal/runtime"
)

// SupportBundle is a redacted, portable snapshot of OpenDash state for support
// and incident analysis. It intentionally omits or redacts secret values.
type SupportBundle struct {
	GeneratedAt    time.Time               `json:"generatedAt"`
	Version        string                  `json:"version"`
	GoOS           string                  `json:"goos"`
	GoArch         string                  `json:"goarch"`
	Warnings       []string                `json:"warnings,omitempty"`
	Summary        *models.SystemSummary   `json:"summary,omitempty"`
	Readiness      *opdruntime.Readiness   `json:"readiness,omitempty"`
	ReadinessError string                  `json:"readinessError,omitempty"`
	Findings       []Finding               `json:"findings"`
	Attention      []models.AttentionItem  `json:"attention,omitempty"`
	Apps           []BundleApp             `json:"apps,omitempty"`
	Catalog        []models.CatalogApp     `json:"catalog,omitempty"`
	Sources        []models.ManifestSource `json:"sources,omitempty"`
	Backups        []models.BackupArchive  `json:"backups,omitempty"`
	Activity       []models.ActivityEvent  `json:"activity,omitempty"`
	Operations     []BundleOperation       `json:"operations,omitempty"`
}

// BundleApp augments an app instance with redacted configuration values.
type BundleApp struct {
	models.AppInstance
	RedactedConfig map[string]any `json:"redactedConfig,omitempty"`
}

// BundleOperation is a sanitized operation record. Output is omitted because
// it may contain unredacted application logs and secret values.
type BundleOperation struct {
	ID        string    `json:"id"`
	AppID     string    `json:"appId"`
	Kind      string    `json:"kind"`
	Status    string    `json:"status"`
	Error     string    `json:"error,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// BundleStore is the subset of store methods BuildBundle needs. It is defined
// here so the diagnostics package does not depend on the store package.
type BundleStore interface {
	GetSummary(ctx context.Context) (*models.SystemSummary, error)
	GetProtection(ctx context.Context) (*models.ProtectionOverview, error)
	ListAttention(ctx context.Context) ([]models.AttentionItem, error)
	ListCatalog(ctx context.Context) ([]models.CatalogApp, error)
	ListActivity(ctx context.Context) ([]models.ActivityEvent, error)
	ListApps(ctx context.Context) ([]models.InstalledApp, error)
	ListAppInstances(ctx context.Context) ([]models.AppInstance, error)
	GetLatestRevision(ctx context.Context, appID string) (*models.Revision, error)
	ListBackupArchives(ctx context.Context, appID string) ([]models.BackupArchive, error)
	ListManifestSources(ctx context.Context) ([]models.ManifestSource, error)
	ListOperationsForApp(ctx context.Context, appID string) ([]models.Operation, error)
}

// BuildBundle collects a redacted support bundle. It tolerates individual query
// failures and reports them as bundle warnings so a partial bundle can still be
// examined.
func BuildBundle(ctx context.Context, st BundleStore, rt opdruntime.Runtime) (*SupportBundle, error) {
	b := &SupportBundle{
		GeneratedAt: time.Now().UTC(),
		Version:     buildVersion(),
		GoOS:        runtime.GOOS,
		GoArch:      runtime.GOARCH,
		Findings:    []Finding{},
	}

	if summary, err := st.GetSummary(ctx); err == nil {
		b.Summary = summary
	} else {
		b.Warnings = append(b.Warnings, fmt.Sprintf("summary: %s", err))
	}

	installed, err := st.ListApps(ctx)
	if err != nil {
		b.Warnings = append(b.Warnings, fmt.Sprintf("apps: %s", err))
		installed = nil
	}

	if ready, err := rt.Ready(ctx); err == nil {
		b.Readiness = ready
	} else {
		b.ReadinessError = err.Error()
		b.Warnings = append(b.Warnings, fmt.Sprintf("readiness: %s", err))
	}

	apps, err := st.ListAppInstances(ctx)
	if err != nil {
		b.Warnings = append(b.Warnings, fmt.Sprintf("app instances: %s", err))
	}

	if protection, err := st.GetProtection(ctx); err == nil {
		b.Findings = Evaluate(installed, protection)
	} else {
		b.Warnings = append(b.Warnings, fmt.Sprintf("protection: %s", err))
	}

	if attention, err := st.ListAttention(ctx); err == nil {
		b.Attention = attention
	} else {
		b.Warnings = append(b.Warnings, fmt.Sprintf("attention: %s", err))
	}

	if catalog, err := st.ListCatalog(ctx); err == nil {
		b.Catalog = catalog
	} else {
		b.Warnings = append(b.Warnings, fmt.Sprintf("catalog: %s", err))
	}

	if sources, err := st.ListManifestSources(ctx); err == nil {
		b.Sources = sources
	} else {
		b.Warnings = append(b.Warnings, fmt.Sprintf("sources: %s", err))
	}

	if activity, err := st.ListActivity(ctx); err == nil {
		b.Activity = activity
	} else {
		b.Warnings = append(b.Warnings, fmt.Sprintf("activity: %s", err))
	}

	for _, app := range apps {
		ba := BundleApp{AppInstance: app}
		if rev, err := st.GetLatestRevision(ctx, app.ID); err == nil && rev.ConfigJSON != "" {
			ba.RedactedConfig = redactConfigJSON(rev.ConfigJSON)
		} else if err != nil {
			b.Warnings = append(b.Warnings, fmt.Sprintf("revision for %s: %s", app.ID, err))
		}
		b.Apps = append(b.Apps, ba)

		if backups, err := st.ListBackupArchives(ctx, app.ID); err == nil {
			b.Backups = append(b.Backups, backups...)
		}

		if ops, err := st.ListOperationsForApp(ctx, app.ID); err == nil {
			for _, op := range ops {
				b.Operations = append(b.Operations, BundleOperation{
					ID:        op.ID,
					AppID:     op.AppID,
					Kind:      string(op.Kind),
					Status:    string(op.Status),
					Error:     op.Error,
					CreatedAt: op.CreatedAt,
					UpdatedAt: op.UpdatedAt,
				})
			}
		}
	}

	sort.Slice(b.Backups, func(i, j int) bool { return b.Backups[i].CreatedAt.After(b.Backups[j].CreatedAt) })
	sort.Slice(b.Operations, func(i, j int) bool { return b.Operations[i].CreatedAt.After(b.Operations[j].CreatedAt) })

	if len(b.Findings) == 0 {
		b.Findings = append(b.Findings, Finding{
			Rule:       "healthy",
			Status:     models.HealthHealthy,
			Confidence: ConfidenceCertain,
			Evidence: []Evidence{
				{Source: "bundle", Message: "No diagnostic findings collected", Timestamp: b.GeneratedAt},
			},
		})
	}

	return b, nil
}

func buildVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" {
		return info.Main.Version
	}
	return "dev"
}

// redactConfigJSON parses a JSON configuration payload and returns a map where
// every leaf value is replaced with "REDACTED". Keys are preserved so the
// support bundle still reveals which settings are in use.
func redactConfigJSON(raw string) map[string]any {
	var root any
	if err := json.Unmarshal([]byte(raw), &root); err != nil {
		return map[string]any{"_parse_error": err.Error()}
	}
	out := redactValue(root)
	m, ok := out.(map[string]any)
	if !ok {
		m = map[string]any{"_redacted": out}
	}
	return m
}

func redactValue(v any) any {
	switch val := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(val))
		for k, child := range val {
			out[k] = redactValue(child)
		}
		return out
	case []any:
		out := make([]any, len(val))
		for i, child := range val {
			out[i] = redactValue(child)
		}
		return out
	default:
		return "REDACTED"
	}
}
