package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/opendash-project/opendash/internal/catalog"
	"github.com/opendash-project/opendash/internal/githubsource"
	"github.com/opendash-project/opendash/internal/models"
	"github.com/opendash-project/opendash/internal/store"
)

// Service manages GitHub manifest sources and update previews.
type Service struct {
	Store    store.SourceManager
	Resolver *githubsource.Resolver
	AppsRoot string
}

// NewService creates a Service. If resolver is nil, a default GitHub resolver is used.
func NewService(st store.SourceManager, resolver *githubsource.Resolver, appsRoot string) *Service {
	if resolver == nil {
		resolver = githubsource.NewResolver()
	}
	return &Service{Store: st, Resolver: resolver, AppsRoot: appsRoot}
}

// Preview resolves a GitHub URL and returns the validated manifest without storing it.
func (svc *Service) Preview(ctx context.Context, rawURL string) (*models.SourcePreview, error) {
	u, err := githubsource.ParseURL(rawURL)
	if err != nil {
		return nil, err
	}
	commit, checksum, manifest, err := svc.Resolver.Resolve(ctx, u)
	if err != nil {
		return nil, err
	}
	trust := models.SourceUntrusted
	existing, err := svc.Store.GetManifestSourceByURL(ctx, rawURL)
	if err == nil && existing != nil {
		trust = existing.Trust
	}
	return &models.SourcePreview{
		SourceURL: rawURL,
		Owner:     u.Owner,
		Repo:      u.Repo,
		Path:      u.Path,
		Ref:       u.Ref,
		CommitSHA: commit,
		Checksum:  checksum,
		Trust:     trust,
		Manifest:  *manifest,
	}, nil
}

// Add persists a manifest source after an explicit preview and confirmation.
func (svc *Service) Add(ctx context.Context, rawURL string, confirmed bool) (*models.ManifestSource, error) {
	if !confirmed {
		return nil, fmt.Errorf("source must be previewed and confirmed before adding")
	}
	u, err := githubsource.ParseURL(rawURL)
	if err != nil {
		return nil, err
	}
	if existing, err := svc.Store.GetManifestSourceByURL(ctx, rawURL); err == nil && existing != nil {
		return nil, fmt.Errorf("source already exists")
	}
	commit, checksum, manifest, err := svc.Resolver.Resolve(ctx, u)
	if err != nil {
		return nil, err
	}
	manifestJSON, err := json.Marshal(manifest)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	src := &models.ManifestSource{
		ID:           uuid.NewString(),
		SourceURL:    rawURL,
		Owner:        u.Owner,
		Repo:         u.Repo,
		Path:         u.Path,
		Ref:          u.Ref,
		CommitSHA:    commit,
		ManifestJSON: string(manifestJSON),
		Checksum:     checksum,
		Trust:        models.SourceUserTrusted,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := svc.Store.CreateManifestSource(ctx, src); err != nil {
		return nil, err
	}
	return src, nil
}

// List returns a summary of all stored manifest sources.
func (svc *Service) List(ctx context.Context) ([]models.SourceListItem, error) {
	sources, err := svc.Store.ListManifestSources(ctx)
	if err != nil {
		return nil, err
	}
	var out []models.SourceListItem
	for _, src := range sources {
		var m models.Manifest
		_ = json.Unmarshal([]byte(src.ManifestJSON), &m)
		out = append(out, models.SourceListItem{
			ID:        src.ID,
			SourceURL: src.SourceURL,
			Owner:     src.Owner,
			Repo:      src.Repo,
			Path:      src.Path,
			Ref:       src.Ref,
			CommitSHA: src.CommitSHA,
			Checksum:  src.Checksum,
			Trust:     src.Trust,
			AppID:     m.ID,
			Name:      m.Name,
			Version:   m.Version,
			UpdatedAt: src.UpdatedAt,
		})
	}
	return out, nil
}

// Remove deletes a manifest source. It will not delete installed app records.
func (svc *Service) Remove(ctx context.Context, id string) error {
	return svc.Store.DeleteManifestSource(ctx, id)
}

// GetManifest loads the cached manifest for a source.
func (svc *Service) GetManifest(ctx context.Context, sourceID string) (*models.Manifest, error) {
	src, err := svc.Store.GetManifestSource(ctx, sourceID)
	if err != nil {
		return nil, err
	}
	var m models.Manifest
	if err := json.Unmarshal([]byte(src.ManifestJSON), &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// UpdatePreview resolves the latest manifest for an installed source app and
// produces a safe update preview. Drift in commit or checksum will mark the
// source as changed and block updates.
func (svc *Service) UpdatePreview(ctx context.Context, inst *models.AppInstance, am store.AppManager) (*models.UpdatePreview, error) {
	if inst.SourceID == "" {
		return nil, fmt.Errorf("app is not from a GitHub source")
	}
	src, err := svc.Store.GetManifestSource(ctx, inst.SourceID)
	if err != nil {
		return nil, err
	}
	rev, err := am.GetLatestRevision(ctx, inst.ID)
	if err != nil {
		return nil, fmt.Errorf("current revision not found")
	}
	u := &githubsource.URL{
		Owner: src.Owner,
		Repo:  src.Repo,
		Ref:   src.Ref,
		Path:  src.Path,
	}
	newCommit, newChecksum, newManifest, err := svc.Resolver.Resolve(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve latest source: %w", err)
	}

	current := rev.Manifest
	preview := &models.UpdatePreview{
		AppID:           inst.ID,
		CurrentVersion:  current.Version,
		NewVersion:      newManifest.Version,
		CurrentCommit:   src.CommitSHA,
		NewCommit:       newCommit,
		NewChecksum:     newChecksum,
		CurrentManifest: current,
		NewManifest:     *newManifest,
		CanUpdate:       false,
	}

	// A moving branch/tag resolving to a new immutable commit is normal update
	// discovery. Only inconsistent bytes for the same commit indicate integrity
	// drift and downgrade trust.
	if newCommit == src.CommitSHA && newChecksum != src.Checksum {
		if src.Trust != models.SourceChanged && src.Trust != models.SourceUntrusted {
			src.Trust = models.SourceChanged
			_ = svc.Store.UpdateManifestSourceTrust(ctx, src.ID, models.SourceChanged)
		}
		preview.BlockedReasons = append(preview.BlockedReasons, "source returned inconsistent content for the stored immutable commit")
	}
	if newCommit != src.CommitSHA {
		preview.Changes = append(preview.Changes, models.UpdateChange{Kind: "commit", Target: "commit", Before: src.CommitSHA, After: newCommit, Summary: "new source commit discovered"})
	}
	if newChecksum != src.Checksum {
		preview.Changes = append(preview.Changes, models.UpdateChange{Kind: "checksum", Target: "checksum", Before: src.Checksum, After: newChecksum, Summary: "manifest content changed"})
	}

	if current.Version != newManifest.Version {
		preview.Changes = append(preview.Changes, models.UpdateChange{
			Kind:    "version",
			Target:  "version",
			Before:  current.Version,
			After:   newManifest.Version,
			Summary: fmt.Sprintf("version %s -> %s", current.Version, newManifest.Version),
		})
	}

	// Load saved config values for plan previews.
	overrides := map[string]string{}
	if rev.ConfigJSON != "" {
		_ = json.Unmarshal([]byte(rev.ConfigJSON), &overrides)
	}

	// Build current and new install plans with the same overrides and with the
	// current instance filtered from conflicts.
	existing, err := am.ListAppInstances(ctx)
	if err != nil {
		return nil, err
	}
	filtered := make([]models.AppInstance, 0, len(existing))
	for _, e := range existing {
		if e.ID != inst.ID {
			filtered = append(filtered, e)
		}
	}
	currentPlan, err := catalog.BuildInstallPlan(&current, svc.AppsRoot, overrides, filtered)
	if err != nil {
		return nil, fmt.Errorf("build current plan: %w", err)
	}
	newPlan, err := catalog.BuildInstallPlan(newManifest, svc.AppsRoot, overrides, filtered)
	if err != nil {
		return nil, fmt.Errorf("build new plan: %w", err)
	}

	preview.Images = newPlan.Images
	preview.AddedImages = diffStrings(newPlan.Images, currentPlan.Images)
	preview.RemovedImages = diffStrings(currentPlan.Images, newPlan.Images)

	preview.Ports = newPlan.Ports
	preview.AddedPorts = diffPorts(newPlan.Ports, currentPlan.Ports)
	preview.RemovedPorts = diffPorts(currentPlan.Ports, newPlan.Ports)

	preview.Volumes = newPlan.Volumes
	preview.AddedVolumes = diffVolumes(newPlan.Volumes, currentPlan.Volumes)
	preview.RemovedVolumes = diffVolumes(currentPlan.Volumes, newPlan.Volumes)

	preview.Environment = newPlan.Environment
	preview.Permissions = newPlan.Permissions
	preview.Risks = newPlan.Risks
	preview.PermissionDiff = permissionDiff(current.Permissions, newManifest.Permissions)
	preview.Changes = append(preview.Changes, imageChangeSummaries(preview.AddedImages, preview.RemovedImages)...)
	preview.Changes = append(preview.Changes, volumeChangeSummaries(preview.AddedVolumes, preview.RemovedVolumes)...)
	preview.Changes = append(preview.Changes, portChangeSummaries(preview.AddedPorts, preview.RemovedPorts)...)
	preview.Changes = append(preview.Changes, permissionChangeSummaries(preview.PermissionDiff)...)

	preview.CanUpdate = len(preview.BlockedReasons) == 0 && (newCommit != src.CommitSHA || newChecksum != src.Checksum)
	preview.MigrationDisclosures = []string{
		"Compose containers are recreated while existing project identity and retained data volumes are reused.",
		"A Compose rollback does not roll back database or volume contents.",
	}
	rollbackSafe := newManifest.Upgrade != nil && newManifest.Upgrade.RollbackSafe
	migrationRisk := newManifest.Upgrade != nil && newManifest.Upgrade.MigrationRisk != ""
	if migrationRisk {
		preview.MigrationDisclosures = append(preview.MigrationDisclosures, newManifest.Upgrade.MigrationRisk)
	}
	preview.RollbackDisclosures = []string{fmt.Sprintf("The prior generated revision remains pinned to commit %s.", src.CommitSHA)}
	if rollbackSafe {
		preview.RollbackDisclosures = append(preview.RollbackDisclosures, "On runtime or health-gate failure OpenDash will attempt to re-apply the prior Compose configuration; data is not reverted.")
	} else {
		preview.RollbackDisclosures = append(preview.RollbackDisclosures, "The publisher does not declare Compose rollback safe; automatic rollback will not be attempted.")
	}
	if newManifest.Upgrade != nil && newManifest.Upgrade.RollbackNotes != "" {
		preview.RollbackDisclosures = append(preview.RollbackDisclosures, newManifest.Upgrade.RollbackNotes)
	}
	permissionIncrease := false
	for _, d := range preview.PermissionDiff {
		if d.State == "added" || (d.State == "changed" && d.Required) {
			permissionIncrease = true
			break
		}
	}
	preview.RequiresAcknowledgement = permissionIncrease || len(preview.RemovedVolumes) > 0 || migrationRisk || !rollbackSafe

	return preview, nil
}

func diffStrings(a, b []string) []string {
	m := make(map[string]bool)
	for _, v := range b {
		m[v] = true
	}
	var out []string
	for _, v := range a {
		if !m[v] {
			out = append(out, v)
		}
	}
	sort.Strings(out)
	return out
}

func portKey(p models.PlannedPort) string {
	return fmt.Sprintf("%d/%s", p.ContainerPort, p.Kind)
}

func diffPorts(a, b []models.PlannedPort) []models.PlannedPort {
	m := make(map[string]bool)
	for _, p := range b {
		m[portKey(p)] = true
	}
	var out []models.PlannedPort
	for _, p := range a {
		if !m[portKey(p)] {
			out = append(out, p)
		}
	}
	return out
}

func volumeKey(v models.PlannedVolume) string {
	return v.Name
}

func diffVolumes(a, b []models.PlannedVolume) []models.PlannedVolume {
	m := make(map[string]bool)
	for _, v := range b {
		m[volumeKey(v)] = true
	}
	var out []models.PlannedVolume
	for _, v := range a {
		if !m[volumeKey(v)] {
			out = append(out, v)
		}
	}
	return out
}

func permissionDiff(current, next []models.ManifestPermission) []models.PermissionDiff {
	cur := make(map[string]models.ManifestPermission)
	for _, p := range current {
		cur[p.Kind] = p
	}
	nxt := make(map[string]models.ManifestPermission)
	for _, p := range next {
		nxt[p.Kind] = p
	}
	var out []models.PermissionDiff
	for _, p := range next {
		state := "added"
		c, ok := cur[p.Kind]
		if ok {
			if c.Required == p.Required {
				state = "unchanged"
			} else {
				state = "changed"
			}
		}
		out = append(out, models.PermissionDiff{
			Kind:        p.Kind,
			Description: p.Description,
			Required:    p.Required,
			State:       state,
		})
	}
	for _, p := range current {
		if _, ok := nxt[p.Kind]; !ok {
			out = append(out, models.PermissionDiff{
				Kind:        p.Kind,
				Description: p.Description,
				Required:    p.Required,
				State:       "removed",
			})
		}
	}
	return out
}

func imageChangeSummaries(added, removed []string) []models.UpdateChange {
	var out []models.UpdateChange
	for _, img := range added {
		out = append(out, models.UpdateChange{Kind: "image", Target: img, Summary: fmt.Sprintf("image %q added", img)})
	}
	for _, img := range removed {
		out = append(out, models.UpdateChange{Kind: "image", Target: img, Summary: fmt.Sprintf("image %q removed", img)})
	}
	return out
}

func volumeChangeSummaries(added, removed []models.PlannedVolume) []models.UpdateChange {
	var out []models.UpdateChange
	for _, v := range added {
		out = append(out, models.UpdateChange{Kind: "volume", Target: v.Name, Summary: fmt.Sprintf("volume %q added", v.Name)})
	}
	for _, v := range removed {
		out = append(out, models.UpdateChange{Kind: "volume", Target: v.Name, Summary: fmt.Sprintf("volume %q removed", v.Name)})
	}
	return out
}

func portChangeSummaries(added, removed []models.PlannedPort) []models.UpdateChange {
	var out []models.UpdateChange
	for _, p := range added {
		out = append(out, models.UpdateChange{Kind: "port", Target: portKey(p), Summary: fmt.Sprintf("port %q added", portKey(p))})
	}
	for _, p := range removed {
		out = append(out, models.UpdateChange{Kind: "port", Target: portKey(p), Summary: fmt.Sprintf("port %q removed", portKey(p))})
	}
	return out
}

func permissionChangeSummaries(diff []models.PermissionDiff) []models.UpdateChange {
	var out []models.UpdateChange
	for _, d := range diff {
		if d.State == "unchanged" {
			continue
		}
		summary := fmt.Sprintf("permission %q %s", d.Kind, d.State)
		if d.State == "changed" {
			summary += fmt.Sprintf(" (required=%v)", d.Required)
		}
		out = append(out, models.UpdateChange{
			Kind:    "permission",
			Target:  d.Kind,
			Summary: summary,
		})
	}
	return out
}

func sortAndDedupStrings(in []string) []string {
	m := make(map[string]bool)
	for _, v := range in {
		m[v] = true
	}
	out := make([]string, 0, len(m))
	for v := range m {
		out = append(out, v)
	}
	sort.Strings(out)
	return out
}

func stringSet(in []string) map[string]bool {
	m := make(map[string]bool)
	for _, v := range in {
		m[v] = true
	}
	return m
}

func stringList(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for v := range m {
		out = append(out, v)
	}
	sort.Strings(out)
	return out
}
