package backup

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/opendash-project/opendash/internal/models"
)

func TestPlanValidatesAndSortsVolumes(t *testing.T) {
	got, err := Plan(models.Manifest{Backup: &models.ManifestBackup{Volumes: []string{"z-data", "app_data"}}})
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"app_data", "z-data"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
	if _, err := Plan(models.Manifest{Backup: &models.ManifestBackup{Volumes: []string{"../../host"}}}); err == nil {
		t.Fatal("expected unsafe volume rejection")
	}
}

func TestArchiveCreationAndMetadataPreviewReader(t *testing.T) {
	stage := t.TempDir()
	metadata := `{"version":1,"app":{"id":"app"},"manifest":{"id":"demo","name":"Demo","version":"1.0.0","description":"d","category":"c","icon":"d","compose":{}},"volumes":["demo_data"],"images":["demo:1"]}`
	if err := os.WriteFile(filepath.Join(stage, "metadata.json"), []byte(metadata), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stage, "volume-demo_data.tar"), []byte("mock volume archive"), 0600); err != nil {
		t.Fatal(err)
	}
	archive := filepath.Join(t.TempDir(), "backup.tar.gz")
	if err := pack(stage, archive); err != nil {
		t.Fatal(err)
	}
	meta, err := readMetadata(archive)
	if err != nil {
		t.Fatal(err)
	}
	if meta.Version != 1 || len(meta.Volumes) != 1 || meta.Images[0] != "demo:1" {
		t.Fatalf("unexpected metadata: %#v", meta)
	}
	if sum, size, err := fileHash(archive); err != nil || sum == "" || size == 0 {
		t.Fatalf("invalid archive hash: %q %d %v", sum, size, err)
	}
}
