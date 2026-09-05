package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/opendash-project/opendash/internal/catalog"
	"github.com/opendash-project/opendash/internal/models"
	"github.com/opendash-project/opendash/internal/store"
)

type sourcePreviewRequest struct {
	URL string `json:"url"`
}

type sourceAddRequest struct {
	URL       string `json:"url"`
	Confirmed bool   `json:"confirmed"`
}

func (h *Handlers) SourcePreview(w http.ResponseWriter, r *http.Request) {
	if h.sourceService == nil {
		RespondError(w, http.StatusNotImplemented, "not_implemented", "source management not available")
		return
	}
	var req sourcePreviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}
	preview, err := h.sourceService.Preview(r.Context(), req.URL)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "source_preview_failed", err.Error())
		return
	}
	RespondJSON(w, http.StatusOK, preview)
}

func (h *Handlers) AddSource(w http.ResponseWriter, r *http.Request) {
	if h.sourceService == nil {
		RespondError(w, http.StatusNotImplemented, "not_implemented", "source management not available")
		return
	}
	var req sourceAddRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}
	src, err := h.sourceService.Add(r.Context(), req.URL, req.Confirmed)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "source_add_failed", err.Error())
		return
	}
	RespondJSON(w, http.StatusCreated, src)
}

func (h *Handlers) ListSources(w http.ResponseWriter, r *http.Request) {
	if h.sourceService == nil {
		RespondJSON(w, http.StatusOK, []models.SourceListItem{})
		return
	}
	items, err := h.sourceService.List(r.Context())
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "store_error", err.Error())
		return
	}
	RespondJSON(w, http.StatusOK, items)
}

func (h *Handlers) RemoveSource(w http.ResponseWriter, r *http.Request) {
	if h.sourceService == nil {
		RespondError(w, http.StatusNotImplemented, "not_implemented", "source management not available")
		return
	}
	id := r.PathValue("id")
	if id == "" {
		RespondError(w, http.StatusBadRequest, "missing_id", "source id is required")
		return
	}
	manager, ok := h.store.(store.AppManager)
	if ok {
		instances, err := manager.ListAppInstances(r.Context())
		if err == nil {
			for _, inst := range instances {
				if inst.SourceID == id {
					RespondError(w, http.StatusConflict, "source_in_use", "remove the installed app before deleting this source")
					return
				}
			}
		}
	}
	if err := h.sourceService.Remove(r.Context(), id); err != nil {
		RespondError(w, http.StatusNotFound, "not_found", err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) AppApplyUpdate(w http.ResponseWriter, r *http.Request) {
	manager, ok := h.store.(store.AppManager)
	if !ok || h.sourceService == nil {
		RespondError(w, http.StatusNotImplemented, "not_implemented", "source updates not available")
		return
	}
	var req models.UpdateApplyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}
	if !req.Confirmed || req.CommitSHA == "" || req.Checksum == "" {
		RespondError(w, http.StatusBadRequest, "confirmation_required", "exact preview commit and checksum plus explicit confirmation are required")
		return
	}
	inst, err := manager.GetAppInstance(r.Context(), r.PathValue("id"))
	if err != nil {
		RespondError(w, http.StatusNotFound, "not_found", err.Error())
		return
	}
	if active, _ := manager.GetActiveOperationForApp(r.Context(), inst.ID); active != nil {
		RespondError(w, http.StatusConflict, "operation_in_progress", "an operation is already in progress for this app")
		return
	}
	preview, err := h.sourceService.UpdatePreview(r.Context(), inst, manager)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "update_preview_failed", err.Error())
		return
	}
	if preview.NewCommit != req.CommitSHA || preview.NewChecksum != req.Checksum {
		RespondError(w, http.StatusConflict, "stale_preview", "source no longer matches the reviewed preview")
		return
	}
	if !preview.CanUpdate {
		RespondError(w, http.StatusConflict, "update_blocked", "update preview is blocked or contains no update")
		return
	}
	if preview.RequiresAcknowledgement && !req.AcknowledgedRisks {
		RespondError(w, http.StatusBadRequest, "risk_acknowledgement_required", "acknowledge permission, volume, migration, and rollback risks")
		return
	}
	if preview.NewManifest.ID != preview.CurrentManifest.ID {
		RespondError(w, http.StatusBadRequest, "identity_change", "updates cannot change the app id")
		return
	}
	now := time.Now().UTC()
	op := &models.Operation{ID: uuid.NewString(), AppID: inst.ID, Kind: models.OperationUpdate, Status: models.OperationPending, CreatedAt: now, UpdatedAt: now}
	if err := manager.CreateOperation(r.Context(), op); err != nil {
		RespondError(w, http.StatusConflict, "operation_error", err.Error())
		return
	}
	h.runSourceUpdate(op, inst, preview)
	RespondJSON(w, http.StatusAccepted, op)
}

func (h *Handlers) runSourceUpdate(op *models.Operation, inst *models.AppInstance, preview *models.UpdatePreview) {
	h.runOperation(op, inst, nil, func(ctx context.Context) (*models.AppInstanceStatus, error) {
		manager := h.store.(store.AppManager)
		src, err := h.sourceService.Store.GetManifestSource(ctx, inst.SourceID)
		if err != nil {
			return nil, err
		}
		if src.CommitSHA != preview.CurrentCommit {
			return nil, fmt.Errorf("source changed before update started")
		}
		oldRev, err := manager.GetLatestRevision(ctx, inst.ID)
		if err != nil {
			return nil, err
		}
		oldCompose, err := os.ReadFile(oldRev.ComposePath)
		if err != nil {
			return nil, fmt.Errorf("snapshot prior compose: %w", err)
		}
		oldEnv, _ := os.ReadFile(oldRev.ComposePath + ".env")
		config := map[string]string{}
		if err := json.Unmarshal([]byte(oldRev.ConfigJSON), &config); err != nil {
			return nil, err
		}
		existing, err := manager.ListAppInstances(ctx)
		if err != nil {
			return nil, err
		}
		filtered := make([]models.AppInstance, 0, len(existing))
		for _, app := range existing {
			if app.ID != inst.ID {
				filtered = append(filtered, app)
			}
		}
		plan, err := catalog.BuildInstallPlan(&preview.NewManifest, h.appsRoot, config, filtered)
		if err != nil {
			return nil, err
		}
		plan.ProjectName, plan.ProjectPath = inst.ProjectName, inst.InstallPath
		ports := map[int]int{}
		for _, manifestEP := range oldRev.Manifest.Endpoints {
			for _, installedEP := range inst.Endpoints {
				if installedEP.Label == manifestEP.Label {
					ports[manifestEP.Port] = endpointHostPort(installedEP.URL)
				}
			}
		}
		revisionID := uuid.NewString()
		revisionPath := filepath.Join(inst.InstallPath, "revisions", revisionID)
		composePath, composeData, envData, err := catalog.RenderComposeFiles(&preview.NewManifest, plan, ports, config, revisionPath)
		if err != nil {
			return nil, err
		}
		_, runErr := h.runtime.Install(ctx, preview.NewManifest.ID, inst.ProjectName, revisionPath, composePath, composeData, envData)
		if runErr == nil {
			runErr = h.waitForUpdateHealth(ctx, inst.ProjectName, composePath)
		}
		if runErr != nil {
			if preview.NewManifest.Upgrade != nil && preview.NewManifest.Upgrade.RollbackSafe {
				if _, rbErr := h.runtime.Install(ctx, oldRev.Manifest.ID, inst.ProjectName, inst.InstallPath, oldRev.ComposePath, oldCompose, oldEnv); rbErr != nil {
					return nil, fmt.Errorf("update failed: %v; Compose rollback failed: %v; database/volume contents were not rolled back", runErr, rbErr)
				}
				return nil, fmt.Errorf("update failed: %v; prior Compose configuration restored; database/volume contents were not rolled back", runErr)
			}
			return nil, fmt.Errorf("update failed: %v; automatic rollback was not declared safe; database/volume contents were not rolled back", runErr)
		}
		manifestJSON, _ := json.Marshal(preview.NewManifest)
		newSrc := *src
		newSrc.CommitSHA, newSrc.Checksum, newSrc.ManifestJSON = preview.NewCommit, preview.NewChecksum, string(manifestJSON)
		newInst := *inst
		newInst.Name, newInst.Version, newInst.ComposePath = preview.NewManifest.Name, preview.NewManifest.Version, composePath
		newInst.Storage = storageModels(plan)
		newInst.Endpoints = endpointModels(&preview.NewManifest, ports)
		newInst.Status, newInst.Health = models.AppInstanceRunning, models.HealthHealthy
		rev := &models.Revision{ID: revisionID, AppID: inst.ID, Manifest: preview.NewManifest, ConfigJSON: oldRev.ConfigJSON, ComposePath: composePath, CreatedAt: time.Now().UTC()}
		if err := manager.ApplySuccessfulUpdate(ctx, &newInst, rev, &newSrc, src.CommitSHA, src.Checksum); err != nil {
			return nil, err
		}
		*inst = newInst
		status := models.AppInstanceRunning
		return &status, nil
	})
}

func endpointHostPort(raw string) int {
	hostPort := raw
	if parsed, err := url.Parse(raw); err == nil && parsed.Host != "" {
		hostPort = parsed.Host
	}
	_, port, err := net.SplitHostPort(hostPort)
	if err != nil {
		return 0
	}
	value, _ := strconv.Atoi(port)
	return value
}

func (h *Handlers) waitForUpdateHealth(ctx context.Context, project, compose string) error {
	deadline := time.NewTimer(h.updateHealthTimeout)
	defer deadline.Stop()
	tick := time.NewTicker(h.updateHealthPoll)
	defer tick.Stop()
	for {
		obs, err := h.runtime.Observe(ctx, project, compose)
		if err == nil && obs.Running && obs.Healthy {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-deadline.C:
			return fmt.Errorf("health gate timed out")
		case <-tick.C:
		}
	}
}

func (h *Handlers) AppUpdatePlan(w http.ResponseWriter, r *http.Request) {
	if h.sourceService == nil {
		RespondError(w, http.StatusNotImplemented, "not_implemented", "source updates not available")
		return
	}
	manager, ok := h.store.(store.AppManager)
	if !ok {
		RespondError(w, http.StatusNotImplemented, "not_implemented", "app management not available")
		return
	}
	id := r.PathValue("id")
	if id == "" {
		RespondError(w, http.StatusBadRequest, "missing_id", "app id is required")
		return
	}
	inst, err := manager.GetAppInstance(r.Context(), id)
	if err != nil {
		RespondError(w, http.StatusNotFound, "not_found", err.Error())
		return
	}
	preview, err := h.sourceService.UpdatePreview(r.Context(), inst, manager)
	if err != nil {
		slog.Error("update preview failed", "error", err, "app_id", id)
		RespondError(w, http.StatusBadRequest, "update_preview_failed", err.Error())
		return
	}
	RespondJSON(w, http.StatusOK, preview)
}

// loadManifestByID returns the manifest for a catalog or a stored GitHub
// source. The second return value is the source ID, or "" for the local
// official catalog.
func (h *Handlers) loadManifestByID(ctx context.Context, id string) (*models.Manifest, string, error) {
	m, err := catalog.LoadManifest(catalogManifestPath(h.catalogRoot, id))
	if err == nil {
		return m, "", nil
	}
	if !os.IsNotExist(err) {
		return nil, "", err
	}
	if h.sourceService == nil {
		return nil, "", err
	}
	sm, ok := h.store.(store.SourceManager)
	if !ok {
		return nil, "", err
	}
	src, err := sm.GetManifestSource(ctx, id)
	if err != nil {
		return nil, "", err
	}
	var sourceManifest models.Manifest
	if err := json.Unmarshal([]byte(src.ManifestJSON), &sourceManifest); err != nil {
		return nil, "", err
	}
	return &sourceManifest, src.ID, nil
}

// toCatalogAppDetailWithTrust builds a catalog detail from a manifest and trust state.
func toCatalogAppDetailWithTrust(m *models.Manifest, trust models.SourceTrust) models.CatalogAppDetail {
	d := toCatalogAppDetail(m)
	d.Trust = trust
	return d
}
