//go:build docker_live

package docker_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/opendash-project/opendash/internal/catalog"
	"github.com/opendash-project/opendash/internal/models"
	"github.com/opendash-project/opendash/internal/runtime/docker"
)

func TestLiveNginxLifecycle(t *testing.T) {
	if os.Getenv("OPENDASH_DOCKER_LIVE") != "1" {
		t.Skip("Set OPENDASH_DOCKER_LIVE=1 to run live Docker tests")
	}
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker not in PATH")
	}
	if err := exec.Command("docker", "version").Run(); err != nil {
		t.Skipf("cannot connect to docker daemon: %v", err)
	}

	appsRoot := t.TempDir()
	rt := docker.NewWithRoot(appsRoot)

	m := &models.Manifest{
		ID:          "test-hello",
		Name:        "Hello Nginx",
		Version:     "1.0.0",
		Description: "Minimal nginx fixture for live Docker verification.",
		Category:    "Testing",
		Icon:        "HN",
		Compose: models.ManifestCompose{
			Inline: &models.ManifestInlineCompose{
				Services: map[string]models.ManifestService{
					"app": {
						Image:   "docker.io/library/nginx:stable@sha256:a3dd43c8fc7a7bab8bca52ec19b97d3e9e230c2f66947c9210f28486e04c1405",
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
			{Key: "NGINX_HOST", DefaultValue: "localhost", Required: false},
		},
	}

	plan, err := catalog.BuildInstallPlan(m, appsRoot, nil, nil)
	if err != nil {
		t.Fatalf("build install plan: %v", err)
	}
	hostPorts, err := catalog.HostPorts(m, nil)
	if err != nil {
		t.Fatalf("allocate host ports: %v", err)
	}
	config := map[string]string{"NGINX_HOST": "localhost"}
	composePath, composeData, envData, err := catalog.RenderComposeFiles(m, plan, hostPorts, config, plan.ProjectPath)
	if err != nil {
		t.Fatalf("render compose files: %v", err)
	}

	projectName := plan.ProjectName
	installPath := plan.ProjectPath
	appID := m.ID

	t.Logf("install path: %s", installPath)
	t.Logf("compose file: %s", composePath)

	// Install (pulls and starts the container)
	_, err = rt.Install(context.Background(), appID, projectName, installPath, composePath, composeData, envData)
	if err != nil {
		t.Fatalf("install failed: %v", err)
	}
	defer func() {
		// Best-effort cleanup of only resources created in this test.
		_, _ = rt.Uninstall(context.Background(), projectName, composePath)
		_ = exec.Command("docker", "volume", "rm", "-f", projectName+"-Data").Run()
		_ = exec.Command("docker", "rmi", "-f", "docker.io/library/nginx:stable@sha256:09cc2702709e6388d979d8030e3ab4eb1ceb699b2dced26d7543e872a822e823").Run()
	}()

	waitCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	hostPort := hostPorts[80]
	url := fmt.Sprintf("http://127.0.0.1:%d/", hostPort)
	if err := waitForURL(waitCtx, url); err != nil {
		t.Fatalf("nginx did not become reachable: %v", err)
	}

	// Logs
	res, err := rt.LogTail(context.Background(), projectName, composePath, 50)
	if err != nil {
		t.Fatalf("logs failed: %v", err)
	}
	if !strings.Contains(res.Output, "nginx") {
		t.Fatalf("logs did not contain nginx output: %s", res.Output)
	}

	// Stop / Start / Restart
	if _, err := rt.Stop(context.Background(), projectName, composePath); err != nil {
		t.Fatalf("stop failed: %v", err)
	}
	if _, err := rt.Start(context.Background(), projectName, composePath); err != nil {
		t.Fatalf("start failed: %v", err)
	}
	if err := waitForURL(context.Background(), url); err != nil {
		t.Fatalf("nginx not reachable after restart: %v", err)
	}
	if _, err := rt.Restart(context.Background(), projectName, composePath); err != nil {
		t.Fatalf("restart failed: %v", err)
	}
	if err := waitForURL(context.Background(), url); err != nil {
		t.Fatalf("nginx not reachable after restart: %v", err)
	}

	// Uninstall preview
	preview, err := rt.UninstallPreview(context.Background(), projectName, composePath)
	if err != nil {
		t.Fatalf("uninstall preview failed: %v", err)
	}
	if !preview.DataRetained {
		t.Fatal("expected data retained flag")
	}
	found := false
	for _, img := range preview.Images {
		if strings.Contains(img, "nginx") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected nginx image in preview, got %v", preview.Images)
	}

	// Uninstall apply
	_, err = rt.Uninstall(context.Background(), projectName, composePath)
	if err != nil {
		t.Fatalf("uninstall failed: %v", err)
	}
	_ = err

	// Verify project directory and files exist with 0600 permissions.
	for _, path := range []string{composePath, composePath + ".env"} {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("expected file %s: %v", path, err)
		}
		if info.Mode().Perm() != 0600 {
			t.Fatalf("expected mode 0600 for %s, got %04o", path, info.Mode().Perm())
		}
	}
}

func waitForURL(ctx context.Context, url string) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		resp, err := http.Get(url)
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			return nil
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(500 * time.Millisecond)
	}
}
