package models

type PermissionDiff struct {
	Kind        string `json:"kind"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required"`
	State       string `json:"state"`
}

type UpdateChange struct {
	Kind    string `json:"kind"`
	Target  string `json:"target"`
	Before  string `json:"before,omitempty"`
	After   string `json:"after,omitempty"`
	Summary string `json:"summary"`
}

type UpdatePreview struct {
	AppID                   string               `json:"appId"`
	CurrentVersion          string               `json:"currentVersion"`
	NewVersion              string               `json:"newVersion"`
	CurrentCommit           string               `json:"currentCommit"`
	NewCommit               string               `json:"newCommit"`
	NewChecksum             string               `json:"newChecksum"`
	CurrentManifest         Manifest             `json:"currentManifest"`
	NewManifest             Manifest             `json:"newManifest"`
	Images                  []string             `json:"images"`
	AddedImages             []string             `json:"addedImages"`
	RemovedImages           []string             `json:"removedImages"`
	Ports                   []PlannedPort        `json:"ports"`
	AddedPorts              []PlannedPort        `json:"addedPorts"`
	RemovedPorts            []PlannedPort        `json:"removedPorts"`
	Volumes                 []PlannedVolume      `json:"volumes"`
	AddedVolumes            []PlannedVolume      `json:"addedVolumes"`
	RemovedVolumes          []PlannedVolume      `json:"removedVolumes"`
	Environment             []PlannedConfig      `json:"environment"`
	Permissions             []ManifestPermission `json:"permissions"`
	PermissionDiff          []PermissionDiff     `json:"permissionDiff"`
	Risks                   []Risk               `json:"risks"`
	Changes                 []UpdateChange       `json:"changes"`
	CanUpdate               bool                 `json:"canUpdate"`
	BlockedReasons          []string             `json:"blockedReasons"`
	MigrationDisclosures    []string             `json:"migrationDisclosures"`
	RollbackDisclosures     []string             `json:"rollbackDisclosures"`
	RequiresAcknowledgement bool                 `json:"requiresAcknowledgement"`
}

type UpdateApplyRequest struct {
	CommitSHA         string `json:"commitSha"`
	Checksum          string `json:"checksum"`
	Confirmed         bool   `json:"confirmed"`
	AcknowledgedRisks bool   `json:"acknowledgedRisks"`
}
