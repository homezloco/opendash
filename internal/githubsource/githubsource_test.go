package githubsource

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestParseURLValid(t *testing.T) {
	cases := []struct {
		name   string
		input  string
		owner  string
		repo   string
		ref    string
		path   string
		isTree bool
	}{
		{
			name:  "blob",
			input: "https://github.com/opendash-project/catalogs/blob/main/apps/uptime-kuma/app.json",
			owner: "opendash-project", repo: "catalogs", ref: "main",
			path: "/apps/uptime-kuma/app.json", isTree: false,
		},
		{
			name:  "tree",
			input: "https://github.com/opendash-project/catalogs/tree/v1.0/apps/uptime-kuma/app.json",
			owner: "opendash-project", repo: "catalogs", ref: "v1.0",
			path: "/apps/uptime-kuma/app.json", isTree: true,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			u, err := ParseURL(c.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if u.Owner != c.owner || u.Repo != c.repo || u.Ref != c.ref || u.Path != c.path || u.IsTree != c.isTree {
				t.Fatalf("got %+v, want owner=%s repo=%s ref=%s path=%s isTree=%v", u, c.owner, c.repo, c.ref, c.path, c.isTree)
			}
		})
	}
}

func TestParseURLRejections(t *testing.T) {
	cases := []string{
		"http://github.com/owner/repo/blob/main/app.json",
		"https://example.com/owner/repo/blob/main/app.json",
		"https://user:pass@github.com/owner/repo/blob/main/app.json",
		"https://github.com/owner/repo/blob/main/app.json#fragment",
		"https://github.com/owner/repo/blob/main/app.json?foo=bar",
		"https://github.com/owner/repo/raw/main/app.json",
		"https://github.com/owner/repo/blob/main/app.txt",
		"https://github.com/owner/repo/blob/main/../etc/app.json",
		"https://github.com/owner/repo/blob/main/.hidden/app.json",
		"https://github.com/owner/repo/blob/main/app.json/extra",
		"https://github.com/owner/repo/blob/main/",
		"https://github.com/",
		"https://github.com/owner/repo",
	}
	for _, input := range cases {
		t.Run(input, func(t *testing.T) {
			if _, err := ParseURL(input); err == nil {
				t.Fatalf("expected error for %q", input)
			}
		})
	}
}

func minimalManifest(name string) []byte {
	return []byte(fmt.Sprintf(`{
  "id": "%s",
  "name": "%s",
  "version": "1.0.0",
  "description": "test",
  "category": "test",
  "icon": "T",
  "compose": {
    "inline": {
      "services": {
        "app": {
          "image": "docker.io/library/app:1.0.0@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
        }
      }
    }
  }
}`, name, name))
}

func TestResolverResolveAndFetch(t *testing.T) {
	manifestName := "github-source-test"
	manifest := minimalManifest(manifestName)
	commit := "abc123def456789012345678901234567890abcd"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/repos/") && strings.Contains(r.URL.Path, "/commits/") {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"sha":"` + commit + `"}`))
			return
		}
		if strings.HasSuffix(r.URL.Path, "/app.json") {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(manifest)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	r := NewResolver()
	r.APIBase = server.URL
	r.RawBase = server.URL

	u := &URL{Owner: "owner", Repo: "repo", Ref: "main", Path: "/app.json"}
	gotCommit, checksum, m, err := r.Resolve(context.Background(), u)
	if err != nil {
		t.Fatalf("resolve failed: %v", err)
	}
	if gotCommit != commit {
		t.Fatalf("commit = %s, want %s", gotCommit, commit)
	}
	if m.ID != manifestName {
		t.Fatalf("manifest id = %s, want %s", m.ID, manifestName)
	}
	wantChecksum := sha256.Sum256(manifest)
	if checksum != hex.EncodeToString(wantChecksum[:]) {
		t.Fatalf("checksum mismatch")
	}
}

func TestResolverOversized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/repos/") && strings.Contains(r.URL.Path, "/commits/") {
			_, _ = w.Write([]byte(`{"sha":"abc123"}`))
			return
		}
		w.Header().Set("Content-Length", fmt.Sprintf("%d", 10*1024*1024))
		_, _ = w.Write(make([]byte, 10*1024*1024))
	}))
	defer server.Close()

	r := NewResolver()
	r.APIBase = server.URL
	r.RawBase = server.URL
	r.MaxBytes = 1024

	u := &URL{Owner: "owner", Repo: "repo", Ref: "main", Path: "/app.json"}
	_, _, _, err := r.Resolve(context.Background(), u)
	if err == nil || !strings.Contains(err.Error(), "size") {
		t.Fatalf("expected oversized error, got %v", err)
	}
}

func TestResolverRejectLocalhostRedirect(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "http://127.0.0.1:1234/", http.StatusFound)
	}))
	defer server.Close()

	r := NewResolver()
	r.APIBase = server.URL
	r.RawBase = server.URL

	_, _, _, err := r.Resolve(context.Background(), &URL{Owner: "owner", Repo: "repo", Ref: "main", Path: "/app.json"})
	if err == nil || !strings.Contains(err.Error(), "private or loopback") {
		t.Fatalf("expected redirect rejection, got %v", err)
	}
}

func TestRejectUnsafeHostDirect(t *testing.T) {
	if err := rejectUnsafeHost("localhost"); err == nil {
		t.Fatal("expected localhost rejection")
	}
	if err := rejectUnsafeHost("127.0.0.1"); err == nil {
		t.Fatal("expected 127.0.0.1 rejection")
	}
	if err := rejectUnsafeHost("10.0.0.1"); err == nil {
		t.Fatal("expected private rejection")
	}
	if err := rejectUnsafeHost("github.com"); err != nil {
		t.Fatalf("unexpected error for github.com: %v", err)
	}
}
