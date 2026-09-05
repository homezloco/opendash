package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/opendash-project/opendash/internal/auth"
	"github.com/opendash-project/opendash/internal/githubsource"
	"github.com/opendash-project/opendash/internal/runtime/noop"
	"github.com/opendash-project/opendash/internal/store"
)

func TestEndToEndAdminCatalogAndSourcePreview(t *testing.T) {
	ctx := context.Background()
	tmp := t.TempDir()

	s, err := store.OpenSQLite(ctx, filepath.Join(tmp, "opendash.db"), false)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	catalogDir := filepath.Join(tmp, "catalog")
	if err := os.MkdirAll(filepath.Join(catalogDir, "apps", "test"), 0755); err != nil {
		t.Fatal(err)
	}
	index := `{"version":"0.1.0","generatedAt":"2024-01-01T00:00:00Z","schemaUrl":"","apps":[{"id":"test","name":"Test App","version":"1.0.0","description":"test","category":"test","icon":"T","installed":false,"manifestPath":"apps/test/app.json"}],"categories":[]}`
	manifest := `{"id":"test","name":"Test App","version":"1.0.0","description":"test","category":"test","icon":"T","compose":{"inline":{"services":{"app":{"image":"example/app:1.0.0@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}}}},"endpoints":[],"storage":[],"secrets":[],"config":[],"permissions":[]}`
	if err := os.WriteFile(filepath.Join(catalogDir, "index.json"), []byte(index), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(catalogDir, "apps", "test", "app.json"), []byte(manifest), 0644); err != nil {
		t.Fatal(err)
	}

	h := NewHandlers(s, noop.New())
	h.ConfigureAuth(auth.New(s.DB(), time.Hour), true, false, time.Hour)
	h.ConfigureCatalog(catalogDir, filepath.Join(tmp, "apps"))
	h.ConfigureBackups(filepath.Join(tmp, "backups"))

	commit := "abc123def456789012345678901234567890abcd"
	ghManifest := []byte(`{"id":"gh-test","name":"GH Test","version":"1.0.0","description":"test","category":"test","icon":"G","compose":{"inline":{"services":{"app":{"image":"example/gh:1.0.0@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}}}},"endpoints":[],"storage":[],"secrets":[],"config":[],"permissions":[]}`)
	mockGH := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/commits/") {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"sha":"` + commit + `"}`))
			return
		}
		if strings.HasSuffix(r.URL.Path, "/app.json") {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(ghManifest)
			return
		}
		http.NotFound(w, r)
	}))
	defer mockGH.Close()

	resolver := githubsource.NewResolver()
	resolver.APIBase = mockGH.URL
	resolver.RawBase = mockGH.URL
	h.SetSourceResolver(resolver)

	server := httptest.NewServer(NewRouter(h, 1<<20))
	defer server.Close()

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	client := server.Client()
	client.Jar = jar

	resp := assertStatus(t, client, http.MethodPost, server.URL+"/api/v1/bootstrap", `{"username":"admin","password":"correct horse battery staple"}`, "", http.StatusCreated)
	var session auth.Session
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if session.Username != "admin" || session.CSRFToken == "" {
		t.Fatalf("unexpected session: %+v", session)
	}

	resp = assertStatus(t, client, http.MethodGet, server.URL+"/api/v1/catalog", "", "", http.StatusOK)
	var catalog []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&catalog); err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if len(catalog) != 1 {
		t.Fatalf("expected 1 catalog app, got %d", len(catalog))
	}
	if catalog[0]["id"] != "test" {
		t.Fatalf("expected catalog app id 'test', got %v", catalog[0]["id"])
	}

	previewURL := "https://github.com/owner/repo/blob/main/app.json"
	resp = assertStatus(t, client, http.MethodPost, server.URL+"/api/v1/sources/preview", `{"url":"`+previewURL+`"}`, session.CSRFToken, http.StatusOK)
	var preview map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&preview); err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if preview["commitSha"] != commit {
		t.Fatalf("expected commit %s, got %v", commit, preview["commitSha"])
	}
	if preview["owner"] != "owner" || preview["repo"] != "repo" {
		t.Fatalf("unexpected owner/repo: %+v", preview)
	}
}
