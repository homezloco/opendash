package diagnostics_test

import (
	"testing"

	"github.com/opendash-project/opendash/internal/diagnostics"
	"github.com/opendash-project/opendash/internal/models"
)

func healthyApp(id string) models.InstalledApp {
	return models.InstalledApp{
		ID:            id,
		Status:        models.AppRunning,
		Health:        models.HealthHealthy,
		Version:       "1",
		LatestVersion: "1",
	}
}

func TestEvaluateHealthy(t *testing.T) {
	apps := []models.InstalledApp{healthyApp("a"), healthyApp("b")}
	protection := &models.ProtectionOverview{
		Status: models.ProtectionArmed,
		Rules: []models.ProtectionRule{
			{Enabled: true, Status: models.HealthHealthy},
		},
	}

	findings := diagnostics.Evaluate(apps, protection)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Status != models.HealthHealthy {
		t.Fatalf("expected healthy, got %s", findings[0].Status)
	}
	if findings[0].Confidence != diagnostics.ConfidenceCertain {
		t.Fatalf("expected certain confidence")
	}
}

func TestEvaluateCriticalApp(t *testing.T) {
	apps := []models.InstalledApp{
		healthyApp("a"),
		{ID: "bad", Status: models.AppError, Health: models.HealthCritical, Version: "1", LatestVersion: "1"},
	}
	findings := diagnostics.Evaluate(apps, &models.ProtectionOverview{Status: models.ProtectionArmed})

	found := false
	for _, f := range findings {
		if f.Rule == "app-critical" && f.Status == models.HealthCritical {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected critical app finding, got %+v", findings)
	}
}

func TestEvaluateProtectionNotArmed(t *testing.T) {
	apps := []models.InstalledApp{healthyApp("a")}
	protection := &models.ProtectionOverview{Status: models.ProtectionMonitoring}

	findings := diagnostics.Evaluate(apps, protection)
	found := false
	for _, f := range findings {
		if f.Rule == "protection-not-armed" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected protection-not-armed finding")
	}
}

func TestEvaluateHealthReturnsCritical(t *testing.T) {
	apps := []models.InstalledApp{
		{ID: "bad", Status: models.AppError, Health: models.HealthCritical, Version: "1", LatestVersion: "1"},
	}
	got := diagnostics.EvaluateHealth(apps, &models.ProtectionOverview{Status: models.ProtectionArmed})
	if got != models.HealthCritical {
		t.Fatalf("expected critical, got %s", got)
	}
}

func TestEvaluateHealthReturnsWarning(t *testing.T) {
	apps := []models.InstalledApp{
		{ID: "warn", Status: models.AppRunning, Health: models.HealthWarning, Version: "1", LatestVersion: "2"},
	}
	got := diagnostics.EvaluateHealth(apps, &models.ProtectionOverview{Status: models.ProtectionArmed})
	if got != models.HealthWarning {
		t.Fatalf("expected warning, got %s", got)
	}
}
