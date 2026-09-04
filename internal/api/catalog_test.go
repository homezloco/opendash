package api_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/opendash-project/opendash/internal/api"
	"github.com/opendash-project/opendash/internal/auth"
	"github.com/opendash-project/opendash/internal/models"
	"github.com/opendash-project/opendash/internal/runtime"
	"github.com/opendash-project/opendash/internal/store"
)

type fakeComposeRuntime struct {
	calls []string
}

func (f *fakeComposeRuntime) Ready(ctx context.Context) (*runtime.Readiness, error) {
	return &runtime.Readiness{Docker: true, Compose: true}, nil
}

func (f *fakeComposeRuntime) Install(ctx context.Context, appID, projectName, installPath, composePath string, composeData, env []byte) (*runtime.ComposeCommandResult, error) {
	f.calls = append(f.calls, "install")
	return &runtime.ComposeCommandResult{Output: "installed"}, nil
}

func (f *fakeComposeRuntime) Start(ctx context.Context, projectName, composePath string) (*runtime.ComposeCommandResult, error) {
	f.calls = append(f.calls, "start")
	return &runtime.ComposeCommandResult{}, nil
}

func (f *fakeComposeRuntime) Stop(ctx context.Context, projectName, composePath string) (*runtime.ComposeCommandResult, error) {
	f.calls = append(f.calls, "stop")
	return &runtime.ComposeCommandResult{}, nil
}

func (f *fakeComposeRuntime) Restart(ctx context.Context, projectName, composePath string) (*runtime.ComposeCommandResult, error) {
	f.calls = append(f.calls, "restart")
	return &runtime.ComposeCommandResult{}, nil
}

func (f *fakeComposeRuntime) LogTail(ctx context.Context, projectName, composePath string, lines int) (*runtime.ComposeCommandResult, error) {
	return &runtime.ComposeCommandResult{Output: "logs"}, nil
}

func (f *fakeComposeRuntime) UninstallPreview(ctx context.Context, projectName, composePath string) (*runtime.ComposePreview, error) {
	return &runtime.ComposePreview{ProjectName: projectName, DataRetained: true}, nil
}

func (f *fakeComposeRuntime) Uninstall(ctx context.Context, projectName, composePath string) (*runtime.ComposeCommandResult, error) {
	f.calls = append(f.calls, "uninstall")
	return &runtime.ComposeCommandResult{}, nil
}

func (f *fakeComposeRuntime) Observe(ctx context.Context, projectName, composePath string) (*runtime.ObservedState, error) {
	return &runtime.ObservedState{Running: true, Healthy: true}, nil
}

func setupCatalogTest(t *testing.T) (*api.Handlers, http.Handler, string, *fakeComposeRuntime) {
	t.Helper()
	dir := t.TempDir()
	catalogRoot := filepath.Join(dir, "catalog")
	appsRoot := filepath.Join(dir, "apps")
	if err := os.MkdirAll(filepath.Join(catalogRoot, "apps", "test-app"), 0755); err != nil {
		t.Fatal(err)
	}
	manifest := []byte(`{
  "id": "test-app",
  "name": "Test App",
  "version": "1.0.0",
  "description": "Test.",
  "category": "Testing",
  "icon": "TA",
  "compose": {
    "inline": {
      "services": {
        "app": {
          "image": "docker.io/library/nginx:1.27@sha256:a3dd43c8fc7a7bab8bca52ec19b97d3e9e230c2f66947c9210f28486e04c1405"
        }
      }
    },
    "mainService": "app"
  },
  "endpoints": [{"label": "Web", "port": 80, "path": "/", "kind": "web"}],
  "storage": [{"name": "Data", "path": "/data"}]
}`)
	if err := os.WriteFile(filepath.Join(catalogRoot, "apps", "test-app", "app.json"), manifest, 0644); err != nil {
		t.Fatal(err)
	}
	index := []byte(`{"version":"1.0.0","generatedAt":"2026-09-04T12:00:00Z","apps":[{"id":"test-app","name":"Test App","version":"1.0.0","category":"Testing","icon":"TA","manifestPath":"apps/test-app/app.json"}],"categories":[]}`)
	if err := os.WriteFile(filepath.Join(catalogRoot, "index.json"), index, 0644); err != nil {
		t.Fatal(err)
	}
	s, err := store.OpenSQLite(context.Background(), filepath.Join(dir, "opendash.db"), false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	fake := &fakeComposeRuntime{}
	h := api.NewHandlers(s, fake)
	h.SetLifecycle(context.Background())
	h.ConfigureCatalog(catalogRoot, appsRoot)
	h.ConfigureAuth(auth.New(s.DB(), time.Hour), true, false, time.Hour)
	return h, api.NewRouter(h, 1<<20), dir, fake
}

func TestCatalogAppDetail(t *testing.T) {
	_, router, _, _ := setupCatalogTest(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/catalog/test-app", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestInstallPlanRequiresAuth(t *testing.T) {
	_, router, _, _ := setupCatalogTest(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/catalog/test-app/install-plan", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestInstallPlanAndInstall(t *testing.T) {
	_, router, dir, fake := setupCatalogTest(t)
	srv := httptest.NewServer(router)
	defer srv.Close()
	client := srv.Client()
	jar, _ := cookiejar.New(nil)
	client.Jar = jar

	// bootstrap
	resp, err := client.Post(srv.URL+"/api/v1/bootstrap", "application/json", jsonReader(`{"username":"admin","password":"correct horse battery staple"}`))
	if err != nil {
		t.Fatal(err)
	}
	var session map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	csrf := session["csrfToken"].(string)

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/api/v1/catalog/test-app/install-plan", nil)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("install-plan status: %d", resp.StatusCode)
	}
	var plan models.InstallPlan
	if err := json.NewDecoder(resp.Body).Decode(&plan); err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if plan.ProjectPath != filepath.Join(dir, "apps", "apps", "test-app") {
		t.Fatalf("unexpected project path: %s", plan.ProjectPath)
	}
	if len(plan.Volumes) != 1 || plan.Volumes[0].HostPath != "opendash-test-app-data" {
		t.Fatalf("expected named volume, got %+v", plan.Volumes)
	}

	installReq, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/v1/catalog/test-app/install", jsonReader(`{"config":{}}`))
	installReq.Header.Set("Content-Type", "application/json")
	installReq.Header.Set("X-CSRF-Token", csrf)
	resp, err = client.Do(installReq)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusAccepted {
		body, _ := readBody(resp)
		t.Fatalf("install status: %d %s", resp.StatusCode, body)
	}
	var op models.Operation
	if err := json.NewDecoder(resp.Body).Decode(&op); err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	// Poll operation until complete.
	for i := 0; i < 50; i++ {
		time.Sleep(20 * time.Millisecond)
		req, _ := http.NewRequest(http.MethodGet, srv.URL+"/api/v1/operations/"+op.ID, nil)
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		var current models.Operation
		if err := json.NewDecoder(resp.Body).Decode(&current); err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if current.Status == models.OperationCompleted || current.Status == models.OperationFailed {
			if current.Status != models.OperationCompleted {
				t.Fatalf("operation failed: %s", current.Error)
			}
			break
		}
	}

	if len(fake.calls) == 0 || fake.calls[0] != "install" {
		t.Fatalf("expected install call, got %v", fake.calls)
	}

	// Find the installed app by catalog id so we use the generated app instance id.
	req, _ = http.NewRequest(http.MethodGet, srv.URL+"/api/v1/apps", nil)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	var apps []models.InstalledApp
	if err := json.NewDecoder(resp.Body).Decode(&apps); err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	var appID string
	for _, app := range apps {
		if app.Name == "Test App" {
			appID = app.ID
			break
		}
	}
	if appID == "" {
		t.Fatal("installed app not found in app list")
	}

	// Start/stop lifecycle
	req, _ = http.NewRequest(http.MethodPost, srv.URL+"/api/v1/apps/"+appID+"/start", nil)
	req.Header.Set("X-CSRF-Token", csrf)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	req, _ = http.NewRequest(http.MethodGet, srv.URL+"/api/v1/apps/"+appID+"/logs?lines=50", nil)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("logs status: %d", resp.StatusCode)
	}
	resp.Body.Close()

	req, _ = http.NewRequest(http.MethodGet, srv.URL+"/api/v1/apps/"+appID+"/uninstall-plan", nil)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("uninstall-plan status: %d", resp.StatusCode)
	}
	resp.Body.Close()

	req, _ = http.NewRequest(http.MethodPost, srv.URL+"/api/v1/apps/"+appID+"/uninstall", nil)
	req.Header.Set("X-CSRF-Token", csrf)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
}

func TestInstallIdempotentConflict(t *testing.T) {
	_, router, _, _ := setupCatalogTest(t)
	srv := httptest.NewServer(router)
	defer srv.Close()
	client := srv.Client()
	jar, _ := cookiejar.New(nil)
	client.Jar = jar

	resp, _ := client.Post(srv.URL+"/api/v1/bootstrap", "application/json", jsonReader(`{"username":"admin","password":"correct horse battery staple"}`))
	var session map[string]any
	json.NewDecoder(resp.Body).Decode(&session)
	resp.Body.Close()
	csrf := session["csrfToken"].(string)

	for i := 0; i < 2; i++ {
		req, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/v1/catalog/test-app/install", jsonReader(`{"config":{}}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-CSRF-Token", csrf)
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		status := resp.StatusCode
		body, _ := readBody(resp)
		if i == 0 && status != http.StatusAccepted {
			t.Fatalf("first install should be accepted, got %d: %s", status, body)
		}
		if i == 1 && status != http.StatusConflict {
			t.Fatalf("second install should conflict, got %d", status)
		}
	}
}

func jsonReader(s string) *strings.Reader {
	return strings.NewReader(s)
}

func readBody(resp *http.Response) (string, error) {
	defer resp.Body.Close()
	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", err
	}
	return fmt.Sprintf("%v", body), nil
}
