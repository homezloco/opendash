package recovery

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/opendash-project/opendash/internal/models"
	"github.com/opendash-project/opendash/internal/store"
)

func TestInventoryExportAndImportPreview(t *testing.T) {
	ctx := context.Background()
	st, err := store.OpenSQLite(ctx, filepath.Join(t.TempDir(), "test.db"), false)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	now := time.Now().UTC()
	app := &models.AppInstance{ID: "app-1", CatalogID: "demo", Name: "Demo", Version: "1.0.0", ProjectName: "demo", InstallPath: filepath.Join(t.TempDir(), "app"), ComposePath: filepath.Join(t.TempDir(), "compose.yml"), Status: models.AppInstanceStopped, Health: models.HealthUnknown, CreatedAt: now, UpdatedAt: now}
	if err = st.CreateAppInstance(ctx, app); err != nil {
		t.Fatal(err)
	}
	svc := NewService(st, st, st)
	inv, err := svc.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if inv.Version != 1 || len(inv.Apps) != 1 {
		t.Fatalf("unexpected inventory: %#v", inv)
	}
	payload, _ := json.Marshal(inv)
	preview, err := svc.Preview(ctx, payload)
	if err != nil {
		t.Fatal(err)
	}
	if !preview.Valid || len(preview.Conflicts) != 1 {
		t.Fatalf("unexpected preview: %#v", preview)
	}
}
