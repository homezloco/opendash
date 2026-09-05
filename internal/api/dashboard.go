package api

import (
	"net/http"

	"github.com/opendash-project/opendash/internal/models"
)

func (h *Handlers) DashboardPreferences(w http.ResponseWriter, r *http.Request) {
	preferences, err := h.store.GetDashboardPreferences(r.Context())
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "store_error", "failed to load dashboard preferences")
		return
	}
	RespondJSON(w, http.StatusOK, preferences)
}

func (h *Handlers) UpdateDashboardPreferences(w http.ResponseWriter, r *http.Request) {
	var preferences models.DashboardPreferences
	if decodeJSON(w, r, &preferences) != nil {
		return
	}
	if !validDashboardPreferences(&preferences) {
		RespondError(w, http.StatusBadRequest, "invalid_preferences", "invalid dashboard preference value")
		return
	}
	if err := h.store.SetDashboardPreferences(r.Context(), &preferences); err != nil {
		RespondError(w, http.StatusInternalServerError, "store_error", "failed to save dashboard preferences")
		return
	}
	RespondJSON(w, http.StatusOK, &preferences)
}

func validDashboardPreferences(p *models.DashboardPreferences) bool {
	return oneOf(p.DefaultView, "grid", "list") &&
		oneOf(p.TileDensity, "compact", "normal", "spacious") &&
		oneOf(p.TileSize, "small", "medium", "large", "wide") &&
		oneOf(p.GroupBy, "none", "category", "status", "health") &&
		oneOf(p.SortBy, "manual", "name", "status", "category") &&
		validHiddenFields(p.HiddenFields)
}

func oneOf(value string, allowed ...string) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}
	return false
}

func validHiddenFields(fields map[string]bool) bool {
	for field := range fields {
		if !oneOf(field, "version", "status", "health", "endpoints") {
			return false
		}
	}
	return true
}
