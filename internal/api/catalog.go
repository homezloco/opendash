package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/opendash-project/opendash/internal/catalog"
	"github.com/opendash-project/opendash/internal/models"
	"github.com/opendash-project/opendash/internal/store"
)

func (h *Handlers) CatalogAppDetail(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		RespondError(w, http.StatusBadRequest, "missing_id", "app id is required")
		return
	}
	m, err := catalog.LoadManifest(catalogManifestPath(h.catalogRoot, id))
	if err != nil {
		if os.IsNotExist(err) {
			RespondError(w, http.StatusNotFound, "not_found", "catalog app not found")
			return
		}
		slog.Error("load catalog manifest failed", "error", err, "request_id", requestIDFromContext(r.Context()))
		RespondError(w, http.StatusInternalServerError, "catalog_error", err.Error())
		return
	}
	RespondJSON(w, http.StatusOK, toCatalogAppDetail(m))
}

func (h *Handlers) InstallPlan(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		RespondError(w, http.StatusBadRequest, "missing_id", "app id is required")
		return
	}
	m, err := catalog.LoadManifest(catalogManifestPath(h.catalogRoot, id))
	if err != nil {
		RespondError(w, http.StatusNotFound, "not_found", "catalog app not found")
		return
	}
	manager, ok := h.store.(store.AppManager)
	if !ok {
		RespondError(w, http.StatusNotImplemented, "not_implemented", "app management not available")
		return
	}
	existing, err := manager.ListAppInstances(r.Context())
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "store_error", err.Error())
		return
	}
	overrides := parseConfigOverrides(r)
	plan, err := catalog.BuildInstallPlan(m, h.appsRoot, overrides, existing)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "invalid_manifest", err.Error())
		return
	}
	RespondJSON(w, http.StatusOK, plan)
}

func (h *Handlers) InstallApp(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		RespondError(w, http.StatusBadRequest, "missing_id", "app id is required")
		return
	}
	manager, ok := h.store.(store.AppManager)
	if !ok {
		RespondError(w, http.StatusNotImplemented, "not_implemented", "app management not available")
		return
	}
	m, err := catalog.LoadManifest(catalogManifestPath(h.catalogRoot, id))
	if err != nil {
		RespondError(w, http.StatusNotFound, "not_found", "catalog app not found")
		return
	}
	ctx := r.Context()
	existing, err := manager.ListAppInstances(ctx)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "store_error", err.Error())
		return
	}
	overrides := parseConfigOverrides(r)
	plan, err := catalog.BuildInstallPlan(m, h.appsRoot, overrides, existing)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "invalid_manifest", err.Error())
		return
	}
	if len(plan.Conflicts) > 0 {
		RespondError(w, http.StatusConflict, "conflict", conflictSummary(plan.Conflicts))
		return
	}
	if active, _ := manager.GetActiveOperationForApp(ctx, m.ID); active != nil {
		RespondError(w, http.StatusConflict, "operation_in_progress", "an operation is already in progress for this app")
		return
	}

	hostPorts, err := catalog.HostPorts(m, nil)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "port_allocation_failed", err.Error())
		return
	}
	// Intra-process port reservation: released once the operation finishes.
	catalog.ReserveHostPorts(hostPorts)

	configValues, _, err := catalog.ApplyConfig(m, overrides)
	if err != nil {
		catalog.ReleaseHostPorts(hostPorts)
		RespondError(w, http.StatusBadRequest, "config_error", err.Error())
		return
	}
	composePath, composeData, envData, err := catalog.RenderComposeFiles(m, plan, hostPorts, configValues, plan.ProjectPath)
	if err != nil {
		catalog.ReleaseHostPorts(hostPorts)
		RespondError(w, http.StatusInternalServerError, "render_error", err.Error())
		return
	}

	now := time.Now().UTC()
	appID := uuid.NewString()
	instance := &models.AppInstance{
		ID:          appID,
		CatalogID:   m.ID,
		Name:        m.Name,
		Version:     m.Version,
		ProjectName: plan.ProjectName,
		InstallPath: plan.ProjectPath,
		ComposePath: composePath,
		Status:      models.AppInstanceInstalling,
		Health:      models.HealthUnknown,
		Endpoints:   endpointModels(m, hostPorts),
		Storage:     storageModels(plan),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := manager.CreateAppInstance(ctx, instance); err != nil {
		catalog.ReleaseHostPorts(hostPorts)
		RespondError(w, http.StatusConflict, "already_installed", err.Error())
		return
	}
	rev := &models.Revision{
		ID:          uuid.NewString(),
		AppID:       appID,
		Manifest:    *m,
		ConfigJSON:  string(mustJSON(configValues)),
		ComposePath: composePath,
		CreatedAt:   now,
	}
	if err := manager.CreateRevision(ctx, rev); err != nil {
		// Best-effort rollback so a failed revision does not leave a phantom app.
		_ = manager.DeleteAppInstance(ctx, appID)
		catalog.ReleaseHostPorts(hostPorts)
		RespondError(w, http.StatusInternalServerError, "revision_error", err.Error())
		return
	}

	op := &models.Operation{
		ID:        uuid.NewString(),
		AppID:     appID,
		Kind:      models.OperationInstall,
		Status:    models.OperationPending,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := manager.CreateOperation(ctx, op); err != nil {
		catalog.ReleaseHostPorts(hostPorts)
		RespondError(w, http.StatusInternalServerError, "operation_error", err.Error())
		return
	}

	go h.runOperation(op, instance, func() { catalog.ReleaseHostPorts(hostPorts) }, func(ctx context.Context) (*models.AppInstanceStatus, error) {
		_, err := h.runtime.Install(ctx, m.ID, plan.ProjectName, plan.ProjectPath, composePath, composeData, envData)
		if err != nil {
			return nil, err
		}
		status := models.AppInstanceRunning
		return &status, nil
	})

	RespondJSON(w, http.StatusAccepted, op)
}

func (h *Handlers) AppOperation(w http.ResponseWriter, r *http.Request) {
	manager, ok := h.store.(store.AppManager)
	if !ok {
		RespondError(w, http.StatusNotImplemented, "not_implemented", "app management not available")
		return
	}
	id := r.PathValue("id")
	if id == "" {
		RespondError(w, http.StatusBadRequest, "missing_id", "operation id is required")
		return
	}
	op, err := manager.GetOperation(r.Context(), id)
	if err != nil {
		RespondError(w, http.StatusNotFound, "not_found", err.Error())
		return
	}
	RespondJSON(w, http.StatusOK, op)
}

func (h *Handlers) runOperation(op *models.Operation, instance *models.AppInstance, cleanup func(), fn func(context.Context) (*models.AppInstanceStatus, error)) {
	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
		if cleanup != nil {
			defer cleanup()
		}
		manager := h.store.(store.AppManager)
		ctx, cancel := context.WithTimeout(h.baseCtx, 15*time.Minute)
		defer cancel()

		secretVals := h.secretValuesForApp(ctx, op.AppID)

		update := func(status models.OperationStatus, output, errMsg string) {
			op.Status = status
			op.Output = redactSecrets(output, secretVals)
			op.Error = redactSecrets(errMsg, secretVals)
			op.UpdatedAt = time.Now().UTC()
			if uErr := manager.UpdateOperation(ctx, op); uErr != nil {
				slog.Error("failed to update operation", "operation_id", op.ID, "error", uErr)
			}
		}

		defer func() {
			if r := recover(); r != nil {
				slog.Error("operation panic", "operation_id", op.ID, "error", r, "stack", string(debug.Stack()))
				update(models.OperationFailed, "", fmt.Sprintf("internal error: %v", r))
				if instance != nil {
					_ = manager.UpdateAppInstanceStatus(ctx, instance.ID, models.AppInstanceError, models.HealthUnknown)
				}
			}
		}()

		update(models.OperationRunning, "", "")
		newStatus, err := fn(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				update(models.OperationCancelled, "", err.Error())
			} else {
				update(models.OperationFailed, "", err.Error())
			}
			if instance != nil {
				_ = manager.UpdateAppInstanceStatus(ctx, instance.ID, models.AppInstanceError, models.HealthUnknown)
			}
			return
		}
		update(models.OperationCompleted, "", "")
		if instance != nil && newStatus != nil {
			_ = manager.UpdateAppInstanceStatus(ctx, instance.ID, *newStatus, models.HealthUnknown)
		}
		if instance != nil {
			_ = h.observeAndUpdateInstance(ctx, manager, instance)
		}
	}()
}

func (h *Handlers) AppStart(w http.ResponseWriter, r *http.Request) {
	h.composeAction(w, r, models.OperationStart, func(ctx context.Context, inst *models.AppInstance) error {
		_, err := h.runtime.Start(ctx, inst.ProjectName, inst.ComposePath)
		return err
	}, models.AppInstanceRunning)
}

func (h *Handlers) AppStop(w http.ResponseWriter, r *http.Request) {
	h.composeAction(w, r, models.OperationStop, func(ctx context.Context, inst *models.AppInstance) error {
		_, err := h.runtime.Stop(ctx, inst.ProjectName, inst.ComposePath)
		return err
	}, models.AppInstanceStopped)
}

func (h *Handlers) AppRestart(w http.ResponseWriter, r *http.Request) {
	h.composeAction(w, r, models.OperationRestart, func(ctx context.Context, inst *models.AppInstance) error {
		_, err := h.runtime.Restart(ctx, inst.ProjectName, inst.ComposePath)
		return err
	}, models.AppInstanceRunning)
}

func (h *Handlers) AppLogs(w http.ResponseWriter, r *http.Request) {
	inst, ok := h.getInstance(w, r)
	if !ok {
		return
	}
	lines := 100
	if raw := r.URL.Query().Get("lines"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n >= 1 && n <= 10000 {
			lines = n
		}
	}
	res, err := h.runtime.LogTail(r.Context(), inst.ProjectName, inst.ComposePath, lines)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "logs_failed", err.Error())
		return
	}
	RespondJSON(w, http.StatusOK, map[string]string{"output": res.Output})
}

func (h *Handlers) AppUninstallPreview(w http.ResponseWriter, r *http.Request) {
	inst, ok := h.getInstance(w, r)
	if !ok {
		return
	}
	preview, err := h.runtime.UninstallPreview(r.Context(), inst.ProjectName, inst.ComposePath)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "uninstall_preview_failed", err.Error())
		return
	}
	RespondJSON(w, http.StatusOK, preview)
}

func (h *Handlers) AppUninstall(w http.ResponseWriter, r *http.Request) {
	inst, ok := h.getInstance(w, r)
	if !ok {
		return
	}
	manager := h.store.(store.AppManager)
	if active, _ := manager.GetActiveOperationForApp(r.Context(), inst.ID); active != nil {
		RespondError(w, http.StatusConflict, "operation_in_progress", "an operation is already in progress for this app")
		return
	}
	op := &models.Operation{
		ID:        uuid.NewString(),
		AppID:     inst.ID,
		Kind:      models.OperationUninstall,
		Status:    models.OperationPending,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	if err := manager.CreateOperation(r.Context(), op); err != nil {
		RespondError(w, http.StatusInternalServerError, "operation_error", err.Error())
		return
	}
	go h.runOperation(op, nil, nil, func(ctx context.Context) (*models.AppInstanceStatus, error) {
		_, err := h.runtime.Uninstall(ctx, inst.ProjectName, inst.ComposePath)
		if err != nil {
			return nil, err
		}
		if dErr := manager.DeleteAppInstance(ctx, inst.ID); dErr != nil {
			slog.Error("failed to delete app instance record after uninstall", "app_id", inst.ID, "error", dErr)
		}
		return nil, nil
	})
	RespondJSON(w, http.StatusAccepted, op)
}

func (h *Handlers) composeAction(w http.ResponseWriter, r *http.Request, kind models.OperationKind, fn func(context.Context, *models.AppInstance) error, successStatus models.AppInstanceStatus) {
	inst, ok := h.getInstance(w, r)
	if !ok {
		return
	}
	manager := h.store.(store.AppManager)
	if active, _ := manager.GetActiveOperationForApp(r.Context(), inst.ID); active != nil {
		RespondError(w, http.StatusConflict, "operation_in_progress", "an operation is already in progress for this app")
		return
	}
	op := &models.Operation{
		ID:        uuid.NewString(),
		AppID:     inst.ID,
		Kind:      kind,
		Status:    models.OperationPending,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	if err := manager.CreateOperation(r.Context(), op); err != nil {
		RespondError(w, http.StatusInternalServerError, "operation_error", err.Error())
		return
	}
	go h.runOperation(op, inst, nil, func(ctx context.Context) (*models.AppInstanceStatus, error) {
		if err := fn(ctx, inst); err != nil {
			return nil, err
		}
		return &successStatus, nil
	})
	RespondJSON(w, http.StatusAccepted, op)
}

func (h *Handlers) getInstance(w http.ResponseWriter, r *http.Request) (*models.AppInstance, bool) {
	manager, ok := h.store.(store.AppManager)
	if !ok {
		RespondError(w, http.StatusNotImplemented, "not_implemented", "app management not available")
		return nil, false
	}
	id := r.PathValue("id")
	if id == "" {
		RespondError(w, http.StatusBadRequest, "missing_id", "app id is required")
		return nil, false
	}
	inst, err := manager.GetAppInstance(r.Context(), id)
	if err != nil {
		RespondError(w, http.StatusNotFound, "not_found", err.Error())
		return nil, false
	}
	return inst, true
}

func catalogManifestPath(root, id string) string {
	return root + "/apps/" + id + "/app.json"
}

func parseConfigOverrides(r *http.Request) map[string]string {
	overrides := make(map[string]string)
	if r.Body == nil {
		return overrides
	}
	defer r.Body.Close()
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return overrides
	}
	configRaw, ok := body["config"].(map[string]any)
	if !ok {
		return overrides
	}
	for k, v := range configRaw {
		overrides[k] = fmt.Sprintf("%v", v)
	}
	return overrides
}

func toCatalogAppDetail(m *models.Manifest) models.CatalogAppDetail {
	return models.CatalogAppDetail{
		ID:          m.ID,
		Name:        m.Name,
		Version:     m.Version,
		Description: m.Description,
		Category:    m.Category,
		Tags:        m.Tags,
		Icon:        m.Icon,
		Website:     m.Website,
		Source:      m.Source,
		Maintainer:  m.Maintainer,
		License:     m.License,
		Compose:     m.Compose,
		Endpoints:   m.Endpoints,
		Storage:     m.Storage,
		Secrets:     m.Secrets,
		Config:      m.Config,
		Permissions: m.Permissions,
	}
}

func endpointModels(m *models.Manifest, hostPorts map[int]int) []models.AppEndpoint {
	var out []models.AppEndpoint
	for _, ep := range m.Endpoints {
		hp := hostPorts[ep.Port]
		url := fmt.Sprintf("http://127.0.0.1:%d%s", hp, ep.Path)
		if ep.Kind == models.EndpointUDP || ep.Kind == models.EndpointTCP {
			url = fmt.Sprintf("127.0.0.1:%d", hp)
		}
		out = append(out, models.AppEndpoint{
			Label: ep.Label,
			URL:   url,
			Kind:  ep.Kind,
		})
	}
	return out
}

func storageModels(plan *models.InstallPlan) []models.AppStorageVolume {
	var out []models.AppStorageVolume
	for _, vol := range plan.Volumes {
		out = append(out, models.AppStorageVolume{
			Name:      vol.Name,
			MountPath: vol.MountPath,
			UsedGb:    0,
			TotalGb:   0,
		})
	}
	return out
}

func conflictSummary(conflicts []models.Conflict) string {
	var parts []string
	for _, c := range conflicts {
		parts = append(parts, c.Description)
	}
	return strings.Join(parts, "; ")
}

func mustJSON(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}

func (h *Handlers) secretValuesForApp(ctx context.Context, catalogID string) map[string]string {
	manager, ok := h.store.(store.AppManager)
	if !ok {
		return nil
	}
	rev, err := manager.GetLatestRevision(ctx, catalogID)
	if err != nil {
		return nil
	}
	values := make(map[string]string)
	if err := json.Unmarshal([]byte(rev.ConfigJSON), &values); err != nil {
		return nil
	}
	secrets := make(map[string]string)
	for _, cfg := range rev.Manifest.Config {
		if cfg.Secret {
			if v, ok := values[cfg.Key]; ok && v != "" {
				secrets[v] = ""
			}
		}
	}
	for _, sec := range rev.Manifest.Secrets {
		if v, ok := values[sec.Name]; ok && v != "" {
			secrets[v] = ""
		}
	}
	return secrets
}

func redactSecrets(output string, secrets map[string]string) string {
	if len(secrets) == 0 || output == "" {
		return output
	}
	type pair struct {
		value string
		mask  string
	}
	var pairs []pair
	for v := range secrets {
		if len(v) >= 3 {
			pairs = append(pairs, pair{value: v, mask: "***"})
		}
	}
	// Replace longest values first to avoid partial replacements.
	sort.Slice(pairs, func(i, j int) bool { return len(pairs[i].value) > len(pairs[j].value) })
	for _, p := range pairs {
		output = strings.ReplaceAll(output, p.value, p.mask)
	}
	return output
}
