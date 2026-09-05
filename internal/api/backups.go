package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"

	"github.com/opendash-project/opendash/internal/backup"
	"github.com/opendash-project/opendash/internal/store"
)

func (h *Handlers) CreateBackup(w http.ResponseWriter, r *http.Request) {
	if h.backupService == nil {
		RespondError(w, 501, "backup_unavailable", "backup service unavailable")
		return
	}
	var req backup.CreateRequest
	if r.Body != nil {
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil && !errors.Is(err, io.EOF) {
			RespondError(w, 400, "invalid_request", "invalid backup request")
			return
		}
	}
	v, err := h.backupService.Create(r.Context(), r.PathValue("id"), req)
	if err != nil {
		RespondError(w, 400, "backup_failed", err.Error())
		return
	}
	RespondJSON(w, 201, v)
}
func (h *Handlers) ListBackups(w http.ResponseWriter, r *http.Request) {
	bm, ok := h.store.(store.BackupManager)
	if !ok {
		RespondError(w, 501, "backup_unavailable", "backup service unavailable")
		return
	}
	v, err := bm.ListBackupArchives(r.Context(), r.PathValue("id"))
	if err != nil {
		RespondError(w, 500, "store_error", err.Error())
		return
	}
	RespondJSON(w, 200, v)
}
func (h *Handlers) RestorePreview(w http.ResponseWriter, r *http.Request) {
	if h.backupService == nil {
		RespondError(w, 501, "backup_unavailable", "backup service unavailable")
		return
	}
	v, err := h.backupService.Preview(r.Context(), r.PathValue("id"))
	if err != nil {
		RespondError(w, 400, "preview_failed", err.Error())
		return
	}
	RespondJSON(w, 200, v)
}
func (h *Handlers) VerifyBackup(w http.ResponseWriter, r *http.Request) {
	if h.backupService == nil {
		RespondError(w, 501, "backup_unavailable", "backup service unavailable")
		return
	}
	v, err := h.backupService.Verify(r.Context(), r.PathValue("id"))
	if err != nil {
		RespondError(w, 501, "verification_unavailable", err.Error())
		return
	}
	RespondJSON(w, 200, v)
}
func (h *Handlers) RestoreBackup(w http.ResponseWriter, r *http.Request) {
	RespondError(w, 501, "restore_disabled", "destructive restore is disabled; use preview and isolated verification")
}
func (h *Handlers) RecoveryExport(w http.ResponseWriter, r *http.Request) {
	if h.recoveryService == nil {
		RespondError(w, 501, "recovery_unavailable", "recovery service unavailable")
		return
	}
	v, err := h.recoveryService.Export(r.Context())
	if err != nil {
		RespondError(w, 500, "export_failed", err.Error())
		return
	}
	w.Header().Set("Content-Disposition", "attachment; filename=opendash-recovery.json")
	RespondJSON(w, 200, v)
}
func (h *Handlers) RecoveryImportPreview(w http.ResponseWriter, r *http.Request) {
	if h.recoveryService == nil {
		RespondError(w, 501, "recovery_unavailable", "recovery service unavailable")
		return
	}
	payload, err := io.ReadAll(io.LimitReader(r.Body, 8<<20+1))
	if err != nil || len(payload) > 8<<20 {
		RespondError(w, 400, "invalid_inventory", "inventory is too large")
		return
	}
	v, err := h.recoveryService.Preview(r.Context(), payload)
	if err != nil {
		RespondError(w, 400, "invalid_inventory", err.Error())
		return
	}
	RespondJSON(w, 200, v)
}

var _ = os.ErrNotExist
