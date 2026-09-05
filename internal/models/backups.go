package models

import "time"

type BackupStatus string

const (
	BackupPending   BackupStatus = "pending"
	BackupRunning   BackupStatus = "running"
	BackupCompleted BackupStatus = "completed"
	BackupFailed    BackupStatus = "failed"
)

type BackupJob struct {
	ID        string       `json:"id"`
	AppID     string       `json:"appId"`
	Status    BackupStatus `json:"status"`
	Error     string       `json:"error,omitempty"`
	CreatedAt time.Time    `json:"createdAt"`
	UpdatedAt time.Time    `json:"updatedAt"`
}

type BackupArchive struct {
	ID             string    `json:"id"`
	JobID          string    `json:"jobId"`
	AppID          string    `json:"appId"`
	Path           string    `json:"-"`
	Size           int64     `json:"size"`
	SHA256         string    `json:"sha256"`
	ManifestSHA256 string    `json:"manifestSha256,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
	VerifiedAt     time.Time `json:"verifiedAt,omitempty"`
	IntegrityOK    *bool     `json:"integrityOk,omitempty"`
}

type BackupSchedule struct {
	ID             string    `json:"id"`
	AppID          string    `json:"appId"`
	Enabled        bool      `json:"enabled"`
	IntervalHours  int       `json:"intervalHours"`
	RetentionCount int       `json:"retentionCount"`
	RetentionDays  int       `json:"retentionDays"`
	NextRunAt      time.Time `json:"nextRunAt,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type ManifestBackup struct {
	Volumes         []string `json:"volumes"`
	IncludeConfig   bool     `json:"includeConfig,omitempty"`
	PauseContainers bool     `json:"pauseContainers,omitempty"`
	Required        bool     `json:"required,omitempty"`
	MigrationRisk   string   `json:"migrationRisk,omitempty"`
	RollbackSafe    bool     `json:"rollbackSafe,omitempty"`
}

type BackupMetadata struct {
	Version       int               `json:"version"`
	App           AppInstance       `json:"app"`
	Manifest      Manifest          `json:"manifest"`
	Source        *ManifestSource   `json:"source,omitempty"`
	Compose       string            `json:"compose,omitempty"`
	Environment   map[string]string `json:"environment,omitempty"`
	Volumes       []string          `json:"volumes"`
	Images        []string          `json:"images"`
	MigrationRisk string            `json:"migrationRisk,omitempty"`
	RollbackSafe  bool              `json:"rollbackSafe"`
	CreatedAt     time.Time         `json:"createdAt"`
}

type RestorePreview struct {
	BackupID      string   `json:"backupId"`
	AppID         string   `json:"appId"`
	Volumes       []string `json:"volumes"`
	Images        []string `json:"images"`
	Manifest      Manifest `json:"manifest"`
	Compatible    bool     `json:"compatible"`
	Compatibility []string `json:"compatibility"`
	MigrationRisk string   `json:"migrationRisk,omitempty"`
	RollbackSafe  bool     `json:"rollbackSafe"`
	SourceTrusted bool     `json:"sourceTrusted"`
	IntegrityOK   bool     `json:"integrityOk"`
}

type VerificationResult struct {
	BackupID    string    `json:"backupId"`
	ProjectName string    `json:"projectName"`
	Healthy     bool      `json:"healthy"`
	CleanedUp   bool      `json:"cleanedUp"`
	VerifiedAt  time.Time `json:"verifiedAt"`
	Error       string    `json:"error,omitempty"`
}
