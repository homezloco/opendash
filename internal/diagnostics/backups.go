package diagnostics

import (
	"fmt"
	"time"

	"github.com/opendash-project/opendash/internal/models"
)

type BackupSignals struct {
	AppID             string
	BackupRequired    bool
	LatestBackup      *models.BackupArchive
	InstalledManifest string
	CurrentManifest   string
	RestoreIntegrity  *bool
	ImportSources     []models.ManifestSource
}

// EvaluateBackupSignals keeps backup/recovery checks deterministic and usable
// by API, scheduler, and tests without requiring Docker.
func EvaluateBackupSignals(v BackupSignals, now time.Time, maxAge time.Duration) []Finding {
	out := []Finding{}
	add := func(rule string, status models.HealthStatus, confidence Confidence, message string) {
		out = append(out, Finding{Rule: rule, Status: status, Confidence: confidence, Evidence: []Evidence{{Source: "backup", Message: message, Timestamp: now.UTC()}}})
	}
	if v.LatestBackup == nil {
		status := models.HealthWarning
		if v.BackupRequired {
			status = models.HealthCritical
			add("missing-required-backup", status, ConfidenceCertain, fmt.Sprintf("app %s requires a backup but has none", v.AppID))
		} else {
			add("no-recent-backup", status, ConfidenceHigh, fmt.Sprintf("app %s has no backup", v.AppID))
		}
	} else if maxAge > 0 && now.Sub(v.LatestBackup.CreatedAt) > maxAge {
		add("no-recent-backup", models.HealthWarning, ConfidenceHigh, fmt.Sprintf("latest backup for %s is older than %s", v.AppID, maxAge))
	}
	if v.InstalledManifest != "" && v.CurrentManifest != "" && v.InstalledManifest != v.CurrentManifest {
		add("config-drift", models.HealthWarning, ConfidenceHigh, "installed manifest differs from the recorded manifest")
	}
	if v.RestoreIntegrity != nil && !*v.RestoreIntegrity {
		add("restore-integrity-mismatch", models.HealthCritical, ConfidenceCertain, "backup archive checksum or restore verification did not match")
	}
	for _, source := range v.ImportSources {
		if source.Trust == models.SourceChanged || source.Trust == models.SourceUntrusted {
			add("untrusted-recovery-source", models.HealthWarning, ConfidenceCertain, fmt.Sprintf("recovery source %s is not trusted", source.SourceURL))
		}
	}
	return out
}
