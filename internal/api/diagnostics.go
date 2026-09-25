package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/opendash-project/opendash/internal/diagnostics"
)

func (h *Handlers) SupportBundle(w http.ResponseWriter, r *http.Request) {
	st, ok := h.store.(diagnostics.BundleStore)
	if !ok {
		RespondError(w, http.StatusNotImplemented, "bundle_unavailable", "diagnostic support bundle is not available for this data store")
		return
	}
	bundle, err := diagnostics.BuildBundle(r.Context(), st, h.runtime)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "bundle_error", err.Error())
		return
	}
	filename := fmt.Sprintf("opendash-support-bundle-%s.json", bundle.GeneratedAt.UTC().Format(time.RFC3339))
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	RespondJSON(w, http.StatusOK, bundle)
}
