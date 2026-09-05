package store

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/opendash-project/opendash/internal/models"
)

func TestDashboardPreferencesDefaultsAndRoundTrip(t *testing.T) {
	ctx := context.Background()
	st, err := OpenSQLite(ctx, filepath.Join(t.TempDir(), "test.db"), false)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	got, err := st.GetDashboardPreferences(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got.DefaultView != "grid" || got.TileDensity != "normal" || got.HiddenFields == nil {
		t.Fatalf("unexpected defaults: %+v", got)
	}
	want := &models.DashboardPreferences{DefaultView: "list", TileDensity: "compact", TileSize: "wide", GroupBy: "category", SortBy: "name", TileOrder: []string{"app-1"}, FavoriteAppIDs: []string{"app-1"}, HiddenFields: map[string]bool{"version": true}}
	if err := st.SetDashboardPreferences(ctx, want); err != nil {
		t.Fatal(err)
	}
	got, err = st.GetDashboardPreferences(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got.DefaultView != "list" || got.TileOrder[0] != "app-1" || got.UpdatedAt.IsZero() {
		t.Fatalf("round trip failed: %+v", got)
	}
}
