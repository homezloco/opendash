package api

import (
	"net/http"
	"os"
	"path/filepath"
)

type Middleware func(http.Handler) http.Handler

func NewRouter(h *Handlers, bodyLimit int64) http.Handler {
	root := http.NewServeMux()
	root.HandleFunc("GET /api/v1/health", h.Health)
	root.HandleFunc("GET /api/v1/bootstrap", h.BootstrapStatus)
	root.HandleFunc("POST /api/v1/bootstrap", h.Bootstrap)
	root.HandleFunc("POST /api/v1/auth/login", h.Login)
	root.HandleFunc("GET /api/v1/auth/session", h.Session)
	root.Handle("POST /api/v1/auth/logout", h.requireAuth(http.HandlerFunc(h.Logout)))

	protected := http.NewServeMux()
	protected.HandleFunc("GET /api/v1/readiness", h.Readiness)
	protected.HandleFunc("GET /api/v1/summary", h.Summary)
	protected.HandleFunc("GET /api/v1/attention", h.Attention)
	protected.HandleFunc("GET /api/v1/apps", h.Apps)
	protected.HandleFunc("GET /api/v1/apps/{id}", h.AppDetail)
	protected.HandleFunc("POST /api/v1/apps/{id}/start", h.AppStart)
	protected.HandleFunc("POST /api/v1/apps/{id}/stop", h.AppStop)
	protected.HandleFunc("POST /api/v1/apps/{id}/restart", h.AppRestart)
	protected.HandleFunc("GET /api/v1/apps/{id}/logs", h.AppLogs)
	protected.HandleFunc("GET /api/v1/apps/{id}/uninstall-plan", h.AppUninstallPreview)
	protected.HandleFunc("POST /api/v1/apps/{id}/uninstall", h.AppUninstall)
	protected.HandleFunc("GET /api/v1/catalog", h.Catalog)
	protected.HandleFunc("GET /api/v1/catalog/{id}", h.CatalogAppDetail)
	protected.HandleFunc("GET /api/v1/catalog/{id}/install-plan", h.InstallPlan)
	protected.HandleFunc("POST /api/v1/catalog/{id}/install", h.InstallApp)
	protected.HandleFunc("GET /api/v1/operations/{id}", h.AppOperation)
	protected.HandleFunc("GET /api/v1/protection", h.Protection)
	protected.HandleFunc("GET /api/v1/activity", h.Activity)
	protected.HandleFunc("POST /api/v1/sources/preview", h.SourcePreview)
	protected.HandleFunc("POST /api/v1/sources", h.AddSource)
	protected.HandleFunc("GET /api/v1/sources", h.ListSources)
	protected.HandleFunc("DELETE /api/v1/sources/{id}", h.RemoveSource)
	protected.HandleFunc("GET /api/v1/apps/{id}/update-plan", h.AppUpdatePlan)
	protected.HandleFunc("POST /api/v1/apps/{id}/update", h.AppApplyUpdate)
	protected.HandleFunc("POST /api/v1/apps/{id}/backups", h.CreateBackup)
	protected.HandleFunc("GET /api/v1/apps/{id}/backups", h.ListBackups)
	protected.HandleFunc("POST /api/v1/backups/{id}/restore-preview", h.RestorePreview)
	protected.HandleFunc("POST /api/v1/backups/{id}/verify", h.VerifyBackup)
	protected.HandleFunc("POST /api/v1/backups/{id}/restore", h.RestoreBackup)
	protected.HandleFunc("GET /api/v1/recovery/export", h.RecoveryExport)
	protected.HandleFunc("POST /api/v1/recovery/import-preview", h.RecoveryImportPreview)
	root.Handle("/api/v1/", h.requireAuth(protected))
	root.Handle("/", serveStaticOrSPA(h.staticRoot))
	return chain(root, RequestID, RequestLogger, Recovery, SecurityHeaders, BodyLimit(bodyLimit))
}

func serveStaticOrSPA(root string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if root == "" {
			RespondError(w, http.StatusNotFound, "not_found", "resource not found")
			return
		}
		if _, err := os.Stat(root); os.IsNotExist(err) {
			RespondError(w, http.StatusNotFound, "not_found", "resource not found")
			return
		}

		// Try to serve an exact file first, then fall back to index.html for
		// client-side routing. Path traversal is blocked by filepath.Clean.
		cleanPath := filepath.Join(root, filepath.Clean("/"+r.URL.Path))
		if info, err := os.Stat(cleanPath); err == nil && !info.IsDir() {
			http.ServeFile(w, r, cleanPath)
			return
		}

		index := filepath.Join(root, "index.html")
		if info, err := os.Stat(index); err == nil && !info.IsDir() {
			http.ServeFile(w, r, index)
			return
		}

		RespondError(w, http.StatusNotFound, "not_found", "resource not found")
	})
}

func chain(h http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}
