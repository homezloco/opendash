// External test package: store imports diagnostics, so an in-package test
// importing store would be an import cycle.
package diagnostics_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/opendash-project/opendash/internal/diagnostics"
	"github.com/opendash-project/opendash/internal/runtime/noop"
	"github.com/opendash-project/opendash/internal/store"
)

func TestBuildBundleFromSQLiteStore(t *testing.T) {
	ctx := context.Background()
	// SQLite, not MemoryStore: BundleStore needs ListBackupArchives, which only
	// the SQLite store implements (same pattern as recovery_test.go).
	st, err := store.OpenSQLite(ctx, filepath.Join(t.TempDir(), "test.db"), false)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	rt := noop.New()

	bundle, err := diagnostics.BuildBundle(ctx, st, rt)
	if err != nil {
		t.Fatalf("BuildBundle failed: %v", err)
	}
	if bundle.GeneratedAt.IsZero() {
		t.Fatalf("bundle missing generatedAt")
	}
	if bundle.Version == "" {
		t.Fatalf("bundle missing version")
	}
	if bundle.Readiness == nil {
		t.Fatalf("bundle missing readiness")
	}
	if len(bundle.Findings) == 0 {
		t.Fatalf("bundle missing findings")
	}
	if bundle.Summary == nil {
		t.Fatalf("bundle missing summary")
	}
	for _, f := range bundle.Findings {
		if f.Rule == "" {
			t.Fatalf("finding missing rule: %+v", f)
		}
	}
	if bundle.Version != "dev" {
		t.Fatalf("expected dev version in test, got %q", bundle.Version)
	}
}
