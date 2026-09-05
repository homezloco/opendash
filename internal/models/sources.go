package models

import "time"

type SourceTrust string

const (
	SourceOfficial    SourceTrust = "official"
	SourceReviewed    SourceTrust = "reviewed"
	SourceUserTrusted SourceTrust = "user-trusted"
	SourceChanged     SourceTrust = "changed"
	SourceUntrusted   SourceTrust = "untrusted"
)

type ManifestSource struct {
	ID           string      `json:"id"`
	SourceURL    string      `json:"sourceUrl"`
	Owner        string      `json:"owner"`
	Repo         string      `json:"repo"`
	Path         string      `json:"path"`
	Ref          string      `json:"ref"`
	CommitSHA    string      `json:"commitSha"`
	ManifestJSON string      `json:"-"`
	Checksum     string      `json:"checksum"`
	Trust        SourceTrust `json:"trust"`
	CreatedAt    time.Time   `json:"createdAt"`
	UpdatedAt    time.Time   `json:"updatedAt"`
}

type SourcePreview struct {
	SourceURL string      `json:"sourceUrl"`
	Owner     string      `json:"owner"`
	Repo      string      `json:"repo"`
	Path      string      `json:"path"`
	Ref       string      `json:"ref"`
	CommitSHA string      `json:"commitSha"`
	Checksum  string      `json:"checksum"`
	Trust     SourceTrust `json:"trust"`
	Manifest  Manifest    `json:"manifest"`
}

type SourceListItem struct {
	ID        string      `json:"id"`
	SourceURL string      `json:"sourceUrl"`
	Owner     string      `json:"owner"`
	Repo      string      `json:"repo"`
	Path      string      `json:"path"`
	Ref       string      `json:"ref"`
	CommitSHA string      `json:"commitSha"`
	Checksum  string      `json:"checksum"`
	Trust     SourceTrust `json:"trust"`
	AppID     string      `json:"appId"`
	Name      string      `json:"name"`
	Version   string      `json:"version"`
	UpdatedAt time.Time   `json:"updatedAt"`
}
