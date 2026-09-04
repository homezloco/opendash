package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/opendash-project/opendash/internal/auth"
	"github.com/opendash-project/opendash/internal/catalog"
	"github.com/opendash-project/opendash/internal/models"
	"github.com/opendash-project/opendash/internal/runtime"
	"github.com/opendash-project/opendash/internal/store"
)

type Handlers struct {
	store           store.Store
	runtime         runtime.Runtime
	auth            *auth.Service
	authEnabled     bool
	secureCookies   bool
	sessionLifetime time.Duration
	limiter         *loginLimiter
	dataSource      string
	catalogRoot     string
	appsRoot        string
	baseCtx         context.Context
	wg              sync.WaitGroup
}

func NewHandlers(store store.Store, runtime runtime.Runtime) *Handlers {
	source := "demo"
	if provider, ok := store.(interface{ DataSource() string }); ok {
		source = provider.DataSource()
	}
	return &Handlers{
		store:       store,
		runtime:     runtime,
		dataSource:  source,
		catalogRoot: "./catalog",
		appsRoot:    "./data/apps",
		baseCtx:     context.Background(),
	}
}

func (h *Handlers) ConfigureCatalog(catalogRoot, appsRoot string) {
	if catalogRoot != "" {
		h.catalogRoot = catalogRoot
	}
	if appsRoot != "" {
		h.appsRoot = appsRoot
	}
}

// SetLifecycle installs the context used for background operations. When the
// lifecycle context is cancelled, running operations are marked cancelled.
func (h *Handlers) SetLifecycle(ctx context.Context) {
	h.baseCtx = ctx
}

// Wait blocks until all background operation goroutines have completed. It
// should be called during graceful shutdown.
func (h *Handlers) Wait() {
	h.wg.Wait()
}

func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	RespondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handlers) Readiness(w http.ResponseWriter, r *http.Request) {
	ready, err := h.runtime.Ready(r.Context())
	if err != nil {
		slog.Error("readiness check failed", "error", err, "request_id", requestIDFromContext(r.Context()))
		RespondError(w, http.StatusServiceUnavailable, "readiness_error", "readiness check failed")
		return
	}
	status := http.StatusOK
	if !ready.Docker || !ready.Compose {
		status = http.StatusServiceUnavailable
	}
	RespondJSON(w, status, ready)
}

func (h *Handlers) Summary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.store.GetSummary(r.Context())
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "store_error", err.Error())
		return
	}
	w.Header().Set("X-OpenDash-Data-Source", h.dataSource)
	RespondJSON(w, http.StatusOK, summary)
}

func (h *Handlers) Attention(w http.ResponseWriter, r *http.Request) {
	items, err := h.store.ListAttention(r.Context())
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "store_error", err.Error())
		return
	}
	w.Header().Set("X-OpenDash-Data-Source", h.dataSource)
	RespondJSON(w, http.StatusOK, items)
}

func (h *Handlers) Apps(w http.ResponseWriter, r *http.Request) {
	apps, err := h.store.ListApps(r.Context())
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "store_error", err.Error())
		return
	}
	for i := range apps {
		if updated, ok := h.observeApp(r.Context(), &apps[i]); ok {
			apps[i] = updated
		}
	}
	w.Header().Set("X-OpenDash-Data-Source", h.dataSource)
	RespondJSON(w, http.StatusOK, apps)
}

func (h *Handlers) AppDetail(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		RespondError(w, http.StatusBadRequest, "missing_id", "app id is required")
		return
	}
	app, err := h.store.GetApp(r.Context(), id)
	if err != nil {
		RespondError(w, http.StatusNotFound, "not_found", err.Error())
		return
	}
	if updated, ok := h.observeApp(r.Context(), app); ok {
		*app = updated
	}
	w.Header().Set("X-OpenDash-Data-Source", h.dataSource)
	RespondJSON(w, http.StatusOK, app)
}

func (h *Handlers) Catalog(w http.ResponseWriter, r *http.Request) {
	catalog, err := h.listCatalog(r.Context())
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "store_error", err.Error())
		return
	}
	w.Header().Set("X-OpenDash-Data-Source", h.dataSource)
	RespondJSON(w, http.StatusOK, catalog)
}

func (h *Handlers) listCatalog(ctx context.Context) ([]models.CatalogApp, error) {
	idx, err := catalog.LoadIndex(h.catalogRoot)
	if err == nil {
		installed := make(map[string]bool)
		if manager, ok := h.store.(store.AppManager); ok {
			instances, err := manager.ListAppInstances(ctx)
			if err == nil {
				for _, inst := range instances {
					installed[inst.CatalogID] = true
				}
			}
		}
		apps := make([]models.CatalogApp, 0, len(idx.Apps))
		for _, entry := range idx.Apps {
			apps = append(apps, models.CatalogApp{
				ID:          entry.ID,
				Name:        entry.Name,
				Icon:        entry.Icon,
				Description: entry.Description,
				Category:    entry.Category,
				Version:     entry.Version,
				Installed:   installed[entry.ID],
			})
		}
		return apps, nil
	}
	return h.store.ListCatalog(ctx)
}

func (h *Handlers) Protection(w http.ResponseWriter, r *http.Request) {
	overview, err := h.store.GetProtection(r.Context())
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "store_error", err.Error())
		return
	}
	w.Header().Set("X-OpenDash-Data-Source", h.dataSource)
	RespondJSON(w, http.StatusOK, overview)
}

func (h *Handlers) Activity(w http.ResponseWriter, r *http.Request) {
	events, err := h.store.ListActivity(r.Context())
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "store_error", err.Error())
		return
	}
	w.Header().Set("X-OpenDash-Data-Source", h.dataSource)
	RespondJSON(w, http.StatusOK, events)
}

// observeApp asks the runtime for the live state of an installed app and
// updates the persisted record so managed app cards reflect Compose state.
func (h *Handlers) observeApp(ctx context.Context, app *models.InstalledApp) (models.InstalledApp, bool) {
	manager, ok := h.store.(store.AppManager)
	if !ok {
		return *app, false
	}
	inst, err := manager.GetAppInstance(ctx, app.ID)
	if err != nil {
		return *app, false
	}
	if err := h.observeAndUpdateInstance(ctx, manager, inst); err != nil {
		return *app, false
	}
	updated := *app
	updated.Status = models.AppStatus(inst.Status)
	updated.Health = inst.Health
	updated.Endpoints = inst.Endpoints
	return updated, true
}

func (h *Handlers) observeAndUpdateInstance(ctx context.Context, manager store.AppManager, inst *models.AppInstance) error {
	obs, err := h.runtime.Observe(ctx, inst.ProjectName, inst.ComposePath)
	if err != nil {
		return err
	}
	status, health := observedToAppInstanceStatus(inst.Status, obs)
	endpoints := updateEndpoints(inst.Endpoints, obs.HostPorts)
	inst.Status = status
	inst.Health = health
	inst.Endpoints = endpoints
	return manager.UpdateAppInstanceStatusAndEndpoints(ctx, inst.ID, status, health, endpoints)
}

func observedToAppInstanceStatus(current models.AppInstanceStatus, obs *runtime.ObservedState) (models.AppInstanceStatus, models.HealthStatus) {
	if len(obs.Services) == 0 {
		return current, models.HealthUnknown
	}
	running := 0
	stopped := 0
	for _, svc := range obs.Services {
		switch svc.State {
		case "running":
			running++
		case "exited", "paused":
			stopped++
		}
	}
	if running == len(obs.Services) {
		return models.AppInstanceRunning, models.HealthHealthy
	}
	if stopped == len(obs.Services) {
		return models.AppInstanceStopped, models.HealthUnknown
	}
	return models.AppInstanceError, models.HealthCritical
}

var portInURL = regexp.MustCompile(`^(https?://[^/:]+|tcp://[^/:]+|udp://[^/:]+|[^/:]+):(\d+)`)

func updateEndpoints(endpoints []models.AppEndpoint, hostPorts map[int]int) []models.AppEndpoint {
	if len(hostPorts) == 0 {
		return endpoints
	}
	out := make([]models.AppEndpoint, len(endpoints))
	copy(out, endpoints)
	for i := range out {
		cp := endpointContainerPort(out[i].URL)
		hp, ok := hostPorts[cp]
		if !ok {
			continue
		}
		out[i].URL = replacePort(out[i].URL, hp)
	}
	return out
}

// endpointContainerPort extracts the numeric port from an endpoint URL. It is
// used to match runtime-published ports back to declared container ports.
func endpointContainerPort(raw string) int {
	// For URLs like http://127.0.0.1:3001/ or 127.0.0.1:3001.
	matches := portInURL.FindStringSubmatch(raw)
	if len(matches) < 3 {
		return 0
	}
	port, _ := strconv.Atoi(matches[2])
	return port
}

func replacePort(raw string, newPort int) string {
	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") || strings.HasPrefix(raw, "tcp://") || strings.HasPrefix(raw, "udp://") {
		return portInURL.ReplaceAllString(raw, fmt.Sprintf("${1}:%d", newPort))
	}
	// Bare host:port form.
	if idx := strings.LastIndex(raw, ":"); idx > 0 {
		return raw[:idx] + fmt.Sprintf(":%d", newPort)
	}
	return raw
}
