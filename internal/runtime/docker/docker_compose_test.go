package docker_test

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/opendash-project/opendash/internal/runtime/docker"
)

func TestInstallRejectsInvalidAppID(t *testing.T) {
	rt := docker.NewWithRoot(t.TempDir())
	_, err := rt.Install(context.Background(), "bad id", "opendash-test", "", "", nil, nil)
	if err == nil {
		t.Fatal("expected error for invalid app id")
	}
}

func TestInstallRejectsBadProjectName(t *testing.T) {
	rt := docker.NewWithRoot(t.TempDir())
	_, err := rt.Install(context.Background(), "test", "myproject", "", "", nil, nil)
	if err == nil {
		t.Fatal("expected error for invalid project name")
	}
}

func TestInstallRejectsPathOutsideRoot(t *testing.T) {
	rt := docker.NewWithRoot(t.TempDir())
	root, _ := filepath.Abs(t.TempDir())
	_, err := rt.Install(context.Background(), "test", "opendash-test", filepath.Join("/tmp", "escape"), filepath.Join(root, "compose.json"), []byte("{}"), []byte(""))
	if err == nil {
		t.Fatal("expected error for path outside root")
	}
}

func TestInstallWritesFilesWithSecurePermissions(t *testing.T) {
	tmp := t.TempDir()
	rt := docker.NewWithRootAndRunner(tmp, func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.Command("echo", "ok")
	})
	root, _ := filepath.Abs(tmp)
	installPath := filepath.Join(root, "apps", "test")
	composePath := filepath.Join(installPath, "compose.json")
	_, err := rt.Install(context.Background(), "test", "opendash-test", installPath, composePath, []byte("{}")[:1], []byte("X=a"))
	if err != nil {
		t.Fatalf("install failed: %v", err)
	}
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

func TestStartStopRestartAndLogs(t *testing.T) {
	tmp := t.TempDir()
	called := 0
	rt := docker.NewWithRootAndRunner(tmp, func(ctx context.Context, name string, args ...string) *exec.Cmd {
		called++
		return exec.Command("echo", "ok")
	})
	root, _ := filepath.Abs(tmp)
	composePath := filepath.Join(root, "compose.json")
	if err := os.WriteFile(composePath, []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, action := range []func(context.Context) error{
		func(ctx context.Context) error { _, err := rt.Start(ctx, "opendash-test", composePath); return err },
		func(ctx context.Context) error { _, err := rt.Stop(ctx, "opendash-test", composePath); return err },
		func(ctx context.Context) error { _, err := rt.Restart(ctx, "opendash-test", composePath); return err },
	} {
		if err := action(context.Background()); err != nil {
			t.Fatalf("action failed: %v", err)
		}
	}
	res, err := rt.LogTail(context.Background(), "opendash-test", composePath, 50)
	if err != nil {
		t.Fatalf("logs failed: %v", err)
	}
	if res.Output != "ok\n" {
		t.Fatalf("unexpected log output: %q", res.Output)
	}
	if called == 0 {
		t.Fatal("expected runner to be invoked")
	}
}

func TestUninstallPreviewParsesConfig(t *testing.T) {
	tmp := t.TempDir()
	rt := docker.NewWithRootAndRunner(tmp, func(ctx context.Context, name string, args ...string) *exec.Cmd {
		payload := `{"services":{"web":{"image":"nginx"}},"volumes":{"data":{}},"networks":{"default":{}}}`
		return exec.Command("echo", payload)
	})
	root, _ := filepath.Abs(tmp)
	composePath := filepath.Join(root, "compose.json")
	if err := os.WriteFile(composePath, []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	preview, err := rt.UninstallPreview(context.Background(), "opendash-test", composePath)
	if err != nil {
		t.Fatalf("preview failed: %v", err)
	}
	if !preview.DataRetained {
		t.Fatal("expected data retained flag")
	}
	if len(preview.Images) != 1 || preview.Images[0] != "nginx" {
		t.Fatalf("unexpected images: %v", preview.Images)
	}
}

func TestObserveParsesComposePs(t *testing.T) {
	tmp := t.TempDir()
	rt := docker.NewWithRootAndRunner(tmp, func(ctx context.Context, name string, args ...string) *exec.Cmd {
		payload := `[{"Name":"opendash-test-app-1","Service":"app","State":"running","Health":"","Publishers":[{"URL":"127.0.0.1","TargetPort":80,"PublishedPort":12345}]}]`
		return exec.Command("echo", payload)
	})
	root, _ := filepath.Abs(tmp)
	composePath := filepath.Join(root, "compose.json")
	if err := os.WriteFile(composePath, []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	state, err := rt.Observe(context.Background(), "opendash-test", composePath)
	if err != nil {
		t.Fatalf("observe failed: %v", err)
	}
	if !state.Running {
		t.Fatal("expected running state")
	}
	if state.HostPorts[80] != 12345 {
		t.Fatalf("unexpected host port: %v", state.HostPorts)
	}
}

func TestCommandTimeout(t *testing.T) {
	tmp := t.TempDir()
	rt := docker.NewWithRootAndRunner(tmp, func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "sleep", "10")
	})
	root, _ := filepath.Abs(tmp)
	composePath := filepath.Join(root, "compose.json")
	if err := os.WriteFile(composePath, []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 0)
	defer cancel()
	_, err := rt.Start(ctx, "opendash-test", composePath)
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline exceeded, got %v", err)
	}
}
