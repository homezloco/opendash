package catalog_test

import (
	"testing"

	"github.com/opendash-project/opendash/internal/catalog"
	"github.com/opendash-project/opendash/internal/models"
)

func manifest() *models.Manifest {
	return &models.Manifest{
		ID:          "test-app",
		Name:        "Test App",
		Version:     "1.0.0",
		Description: "A test app.",
		Category:    "Testing",
		Icon:        "TA",
		Compose: models.ManifestCompose{
			Inline: &models.ManifestInlineCompose{
				Services: map[string]models.ManifestService{
					"app": {
						Image:   "docker.io/library/nginx:1.27@sha256:deadbeef",
						Restart: "unless-stopped",
					},
				},
			},
			MainService: "app",
		},
		Endpoints: []models.ManifestEndpoint{
			{Label: "Web", Port: 80, Path: "/", Kind: models.EndpointWeb},
		},
		Storage: []models.ManifestStorage{
			{Name: "Data", Path: "/data"},
		},
		Config: []models.ManifestConfig{
			{Key: "TZ", DefaultValue: "UTC", Required: false},
		},
	}
}

func TestValidateManifestOK(t *testing.T) {
	if err := catalog.ValidateManifest(manifest()); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
}

func TestValidateManifestMissingCompose(t *testing.T) {
	m := manifest()
	m.Compose = models.ManifestCompose{}
	if err := catalog.ValidateManifest(m); err == nil {
		t.Fatal("expected error for missing compose")
	}
}

func TestParsePortMapping(t *testing.T) {
	cases := []struct {
		in       string
		wantHost int
		wantCont int
	}{
		{"8080", 0, 8080},
		{"8080:80", 8080, 80},
		{"127.0.0.1:8080:80", 8080, 80},
	}
	for _, c := range cases {
		hp, cp, err := catalog.ParsePortMapping(c.in)
		if err != nil {
			t.Fatalf("parse %q: %v", c.in, err)
		}
		if hp != c.wantHost || cp != c.wantCont {
			t.Fatalf("%q: got %d:%d, want %d:%d", c.in, hp, cp, c.wantHost, c.wantCont)
		}
	}
}

func TestBuildInstallPlan(t *testing.T) {
	m := manifest()
	plan, err := catalog.BuildInstallPlan(m, "/tmp/apps", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plan.AppID != "test-app" {
		t.Fatalf("unexpected app id: %s", plan.AppID)
	}
	if len(plan.Images) != 1 {
		t.Fatalf("expected 1 image, got %d", len(plan.Images))
	}
	if len(plan.Ports) != 1 || plan.Ports[0].ContainerPort != 80 {
		t.Fatalf("unexpected ports: %+v", plan.Ports)
	}
	if len(plan.Volumes) != 1 || plan.Volumes[0].MountPath != "/data" {
		t.Fatalf("unexpected volumes: %+v", plan.Volumes)
	}
	if plan.ProjectPath != "/tmp/apps/apps/test-app" {
		t.Fatalf("unexpected project path: %s", plan.ProjectPath)
	}
}

func TestBuildInstallPlanConflictExisting(t *testing.T) {
	m := manifest()
	existing := []models.AppInstance{
		{CatalogID: "test-app", ProjectName: "opendash-test-app", InstallPath: "/tmp/apps/apps/test-app"},
	}
	plan, err := catalog.BuildInstallPlan(m, "/tmp/apps", nil, existing)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plan.Conflicts) == 0 {
		t.Fatal("expected conflicts for existing app")
	}
}

func TestHostPortsAllocates(t *testing.T) {
	m := manifest()
	ports, err := catalog.HostPorts(m, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ports[80] == 0 {
		t.Fatal("expected allocated host port")
	}
}

func TestRenderInlineCompose(t *testing.T) {
	m := manifest()
	plan, err := catalog.BuildInstallPlan(m, "/tmp/apps", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ports := map[int]int{80: 12345}
	config := map[string]string{"TZ": "UTC"}
	composePath, compose, env, err := catalog.RenderComposeFiles(m, plan, ports, config, plan.ProjectPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if composePath == "" || len(compose) == 0 || len(env) == 0 {
		t.Fatal("expected rendered compose and env files")
	}
	if len(env) == 0 {
		t.Fatal("expected env file")
	}
}
