package api_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/opendash-project/opendash/internal/api"
	"github.com/opendash-project/opendash/internal/models"
	"github.com/opendash-project/opendash/internal/runtime"
	"github.com/opendash-project/opendash/internal/store"
)

type mockRuntime struct {
	ready *runtime.Readiness
}

func (m *mockRuntime) Ready(ctx context.Context) (*runtime.Readiness, error) {
	return m.ready, nil
}

func (m *mockRuntime) Install(ctx context.Context, appID, projectName, installPath, composePath string, composeData, env []byte) (*runtime.ComposeCommandResult, error) {
	return nil, errors.New("mock install not implemented")
}

func (m *mockRuntime) Start(ctx context.Context, projectName, composePath string) (*runtime.ComposeCommandResult, error) {
	return nil, errors.New("mock start not implemented")
}

func (m *mockRuntime) Stop(ctx context.Context, projectName, composePath string) (*runtime.ComposeCommandResult, error) {
	return nil, errors.New("mock stop not implemented")
}

func (m *mockRuntime) Restart(ctx context.Context, projectName, composePath string) (*runtime.ComposeCommandResult, error) {
	return nil, errors.New("mock restart not implemented")
}

func (m *mockRuntime) LogTail(ctx context.Context, projectName, composePath string, lines int) (*runtime.ComposeCommandResult, error) {
	return &runtime.ComposeCommandResult{Output: "mock log output"}, nil
}

func (m *mockRuntime) UninstallPreview(ctx context.Context, projectName, composePath string) (*runtime.ComposePreview, error) {
	return &runtime.ComposePreview{ProjectName: projectName, DataRetained: true}, nil
}

func (m *mockRuntime) Uninstall(ctx context.Context, projectName, composePath string) (*runtime.ComposeCommandResult, error) {
	return &runtime.ComposeCommandResult{}, nil
}

func (m *mockRuntime) Observe(ctx context.Context, projectName, composePath string) (*runtime.ObservedState, error) {
	return &runtime.ObservedState{}, nil
}

func setup(t *testing.T) (*api.Handlers, http.Handler) {
	t.Helper()
	rt := &mockRuntime{ready: &runtime.Readiness{Docker: true, Compose: true, DockerVersion: "24.0", ComposeVersion: "2.20"}}
	h := api.NewHandlers(store.NewMemoryStore(), rt)
	return h, api.NewRouter(h, 1024*1024)
}

func TestHealth(t *testing.T) {
	_, router := setup(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
	}
	var body map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["status"] != "ok" {
		t.Fatalf("expected status ok, got %v", body)
	}
}

func TestReadiness(t *testing.T) {
	_, router := setup(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/readiness", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	var body runtime.Readiness
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if !body.Docker || !body.Compose {
		t.Fatalf("expected docker and compose ready, got %+v", body)
	}
}

func TestSummary(t *testing.T) {
	_, router := setup(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/summary", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	var body models.SystemSummary
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.AppsTotal != 10 {
		t.Fatalf("expected 10 total apps, got %d", body.AppsTotal)
	}
	if rr.Header().Get("X-OpenDash-Data-Source") != "demo" {
		t.Fatalf("expected demo data header")
	}
}

func TestAppsAndDetail(t *testing.T) {
	_, router := setup(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/apps", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	var apps []models.InstalledApp
	if err := json.Unmarshal(rr.Body.Bytes(), &apps); err != nil {
		t.Fatal(err)
	}
	if len(apps) == 0 {
		t.Fatal("expected apps")
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/apps/immich", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	var app models.InstalledApp
	if err := json.Unmarshal(rr.Body.Bytes(), &app); err != nil {
		t.Fatal(err)
	}
	if app.ID != "immich" {
		t.Fatalf("expected immich, got %s", app.ID)
	}
}

func TestAppDetailNotFound(t *testing.T) {
	_, router := setup(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/apps/unknown", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rr.Code)
	}
}

func TestCatalog(t *testing.T) {
	_, router := setup(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/catalog", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	var catalog []models.CatalogApp
	if err := json.Unmarshal(rr.Body.Bytes(), &catalog); err != nil {
		t.Fatal(err)
	}
	if len(catalog) == 0 {
		t.Fatal("expected catalog apps")
	}
}

func TestProtection(t *testing.T) {
	_, router := setup(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/protection", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	var overview models.ProtectionOverview
	if err := json.Unmarshal(rr.Body.Bytes(), &overview); err != nil {
		t.Fatal(err)
	}
	if overview.Status != models.ProtectionArmed {
		t.Fatalf("expected armed, got %s", overview.Status)
	}
}

func TestActivity(t *testing.T) {
	_, router := setup(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/activity", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	var events []models.ActivityEvent
	if err := json.Unmarshal(rr.Body.Bytes(), &events); err != nil {
		t.Fatal(err)
	}
	if len(events) == 0 {
		t.Fatal("expected events")
	}
}

func TestNotFound(t *testing.T) {
	_, router := setup(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/missing", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rr.Code)
	}
}
