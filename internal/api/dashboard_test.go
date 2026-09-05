package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/opendash-project/opendash/internal/store"
)

func TestDashboardPreferencesHandlers(t *testing.T) {
	h := NewHandlers(store.NewMemoryStore(), nil)
	get := httptest.NewRecorder()
	h.DashboardPreferences(get, httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/preferences", nil))
	if get.Code != http.StatusOK || !strings.Contains(get.Body.String(), `"defaultView":"grid"`) {
		t.Fatalf("unexpected GET: %d %s", get.Code, get.Body.String())
	}

	bad := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/dashboard/preferences", strings.NewReader(`{"defaultView":"cards","tileDensity":"normal","tileSize":"medium","groupBy":"none","sortBy":"manual","tileOrder":[],"favoriteAppIds":[],"hiddenFields":{}}`))
	req.Header.Set("Content-Type", "application/json")
	h.UpdateDashboardPreferences(bad, req)
	if bad.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid enum rejection, got %d", bad.Code)
	}

	unknown := httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, "/api/v1/dashboard/preferences", strings.NewReader(`{"defaultView":"grid","tileDensity":"normal","tileSize":"medium","groupBy":"none","sortBy":"manual","tileOrder":[],"favoriteAppIds":[],"hiddenFields":{},"extra":true}`))
	req.Header.Set("Content-Type", "application/json")
	h.UpdateDashboardPreferences(unknown, req)
	if unknown.Code != http.StatusBadRequest {
		t.Fatalf("expected unknown field rejection, got %d", unknown.Code)
	}
}
