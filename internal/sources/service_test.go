package sources

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/opendash-project/opendash/internal/githubsource"
	"github.com/opendash-project/opendash/internal/models"
	"github.com/opendash-project/opendash/internal/store"
)

func TestUpdatePreviewTrustSemantics(t *testing.T) {
	for _, tc := range []struct {
		name, oldCommit, newCommit              string
		changedContent, wantChanged, wantUpdate bool
	}{
		{"new commit is expected discovery", strings.Repeat("a", 40), strings.Repeat("b", 40), true, false, true},
		{"same commit inconsistent checksum is drift", strings.Repeat("a", 40), strings.Repeat("a", 40), true, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			st, err := store.OpenSQLite(ctx, t.TempDir()+"/test.db", false)
			if err != nil {
				t.Fatal(err)
			}
			defer st.Close()
			oldBody := testManifest("1.0.0")
			newBody := oldBody
			if tc.changedContent {
				newBody = testManifest("2.0.0")
			}
			sum := sha256.Sum256(oldBody)
			src := &models.ManifestSource{ID: "source", SourceURL: "https://github.com/owner/repo/blob/main/app.json", Owner: "owner", Repo: "repo", Ref: "main", Path: "/app.json", CommitSHA: tc.oldCommit, Checksum: hex.EncodeToString(sum[:]), ManifestJSON: string(oldBody), Trust: models.SourceUserTrusted}
			if err := st.CreateManifestSource(ctx, src); err != nil {
				t.Fatal(err)
			}
			inst := &models.AppInstance{ID: "instance", CatalogID: "source", SourceID: "source", Name: "app", Version: "1.0.0", ProjectName: "opendash-app", InstallPath: t.TempDir() + "/apps/app", ComposePath: t.TempDir() + "/compose.yml", Status: models.AppInstanceRunning, Health: models.HealthHealthy}
			if err := st.CreateAppInstance(ctx, inst); err != nil {
				t.Fatal(err)
			}
			var old models.Manifest
			_ = json.Unmarshal(oldBody, &old)
			if err := st.CreateRevision(ctx, &models.Revision{ID: "revision", AppID: inst.ID, Manifest: old, ConfigJSON: "{}", ComposePath: inst.ComposePath, CreatedAt: time.Now()}); err != nil {
				t.Fatal(err)
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.Contains(r.URL.Path, "/commits/") {
					_, _ = w.Write([]byte(`{"sha":"` + tc.newCommit + `"}`))
					return
				}
				_, _ = w.Write(newBody)
			}))
			defer server.Close()
			resolver := githubsource.NewResolver()
			resolver.APIBase, resolver.RawBase = server.URL, server.URL
			preview, err := NewService(st, resolver, t.TempDir()).UpdatePreview(ctx, inst, st)
			if err != nil {
				t.Fatal(err)
			}
			if preview.CanUpdate != tc.wantUpdate {
				t.Fatalf("CanUpdate=%v want %v (%v)", preview.CanUpdate, tc.wantUpdate, preview.BlockedReasons)
			}
			stored, _ := st.GetManifestSource(ctx, src.ID)
			if (stored.Trust == models.SourceChanged) != tc.wantChanged {
				t.Fatalf("trust=%s", stored.Trust)
			}
			if stored.CommitSHA != tc.oldCommit || stored.ManifestJSON != string(oldBody) {
				t.Fatal("preview mutated stored source snapshot")
			}
		})
	}
}

func testManifest(version string) []byte {
	return []byte(`{"id":"app","name":"App","version":"` + version + `","description":"test","category":"test","icon":"A","compose":{"inline":{"services":{"app":{"image":"example/app:` + version + `"}}}}}`)
}
