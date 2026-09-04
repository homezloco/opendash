package diagnostics

import (
	"fmt"
	"time"

	"github.com/opendash-project/opendash/internal/models"
)

type Confidence string

const (
	ConfidenceCertain Confidence = "certain"
	ConfidenceHigh    Confidence = "high"
	ConfidenceMedium  Confidence = "medium"
	ConfidenceLow     Confidence = "low"
)

type Evidence struct {
	Source    string    `json:"source"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

type Finding struct {
	Rule       string              `json:"rule"`
	Status     models.HealthStatus `json:"status"`
	Confidence Confidence          `json:"confidence"`
	Evidence   []Evidence          `json:"evidence"`
}

func EvaluateHealth(apps []models.InstalledApp, protection *models.ProtectionOverview) models.HealthStatus {
	findings := Evaluate(apps, protection)
	for _, f := range findings {
		if f.Status == models.HealthCritical {
			return models.HealthCritical
		}
	}
	for _, f := range findings {
		if f.Status == models.HealthWarning {
			return models.HealthWarning
		}
	}
	return models.HealthHealthy
}

func Evaluate(apps []models.InstalledApp, protection *models.ProtectionOverview) []Finding {
	var findings []Finding
	now := time.Now().UTC()

	criticalApps := 0
	warningApps := 0
	stoppedApps := 0
	runningApps := 0
	updates := 0
	for _, app := range apps {
		switch app.Status {
		case models.AppError:
			criticalApps++
		case models.AppStopped:
			stoppedApps++
		case models.AppRunning:
			runningApps++
		}
		if app.Health == models.HealthCritical {
			criticalApps++
		} else if app.Health == models.HealthWarning {
			warningApps++
		}
		if app.Version != app.LatestVersion {
			updates++
		}
	}

	if criticalApps > 0 {
		findings = append(findings, Finding{
			Rule:       "app-critical",
			Status:     models.HealthCritical,
			Confidence: ConfidenceCertain,
			Evidence: []Evidence{
				{Source: "apps", Message: fmt.Sprintf("%d app(s) report critical/error status", criticalApps), Timestamp: now},
			},
		})
	}
	if warningApps > 0 {
		findings = append(findings, Finding{
			Rule:       "app-warning",
			Status:     models.HealthWarning,
			Confidence: ConfidenceHigh,
			Evidence: []Evidence{
				{Source: "apps", Message: fmt.Sprintf("%d app(s) report warning health", warningApps), Timestamp: now},
			},
		})
	}
	if len(apps) > 0 && runningApps == 0 {
		findings = append(findings, Finding{
			Rule:       "no-running-apps",
			Status:     models.HealthCritical,
			Confidence: ConfidenceHigh,
			Evidence: []Evidence{
				{Source: "apps", Message: "No apps are currently running", Timestamp: now},
			},
		})
	}
	if updates >= 3 {
		findings = append(findings, Finding{
			Rule:       "updates-pending",
			Status:     models.HealthWarning,
			Confidence: ConfidenceMedium,
			Evidence: []Evidence{
				{Source: "apps", Message: fmt.Sprintf("%d updates are available", updates), Timestamp: now},
			},
		})
	}
	if stoppedApps > 0 {
		findings = append(findings, Finding{
			Rule:       "stopped-apps",
			Status:     models.HealthWarning,
			Confidence: ConfidenceLow,
			Evidence: []Evidence{
				{Source: "apps", Message: fmt.Sprintf("%d app(s) are stopped", stoppedApps), Timestamp: now},
			},
		})
	}

	if protection != nil {
		if protection.Status != models.ProtectionArmed {
			findings = append(findings, Finding{
				Rule:       "protection-not-armed",
				Status:     models.HealthWarning,
				Confidence: ConfidenceHigh,
				Evidence: []Evidence{
					{Source: "protection", Message: fmt.Sprintf("Protection status is %s", protection.Status), Timestamp: now},
				},
			})
		}
		var failedRules int
		for _, rule := range protection.Rules {
			if !rule.Enabled || rule.Status == models.HealthCritical {
				failedRules++
			}
		}
		if failedRules > 0 {
			findings = append(findings, Finding{
				Rule:       "protection-rule-failure",
				Status:     models.HealthWarning,
				Confidence: ConfidenceMedium,
				Evidence: []Evidence{
					{Source: "protection", Message: fmt.Sprintf("%d protection rule(s) are disabled or critical", failedRules), Timestamp: now},
				},
			})
		}
	}

	if len(findings) == 0 {
		findings = append(findings, Finding{
			Rule:       "healthy",
			Status:     models.HealthHealthy,
			Confidence: ConfidenceCertain,
			Evidence: []Evidence{
				{Source: "system", Message: "All diagnostics rules passed", Timestamp: now},
			},
		})
	}

	return findings
}
