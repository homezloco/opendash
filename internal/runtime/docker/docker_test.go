package docker_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/opendash-project/opendash/internal/runtime/docker"
)

type invocation struct {
	name string
	args []string
}

func TestInstallRejectsSymlinkEscape(t *testing.T) {
	tmp := t.TempDir()
	target := t.TempDir()
	linkRoot := filepath.Join(tmp, "link-root")
	if err := os.Symlink(target, linkRoot); err != nil {
		t.Fatal(err)
	}
	rt := docker.NewWithRoot(linkRoot)
	_, err := rt.Install(context.Background(), "test", "opendash-test", filepath.Join(linkRoot, "apps", "test"), "compose.json", []byte("{}"), []byte(""))
	if err == nil {
		t.Fatal("expected error for symlinked managed root")
	}
}

func TestReadyBothAvailable(t *testing.T) {
	var calls []invocation
	runner := func(ctx context.Context, name string, args ...string) *exec.Cmd {
		calls = append(calls, invocation{name: name, args: args})
		if name == "docker" && len(args) == 3 && args[0] == "version" {
			return exec.Command("echo", "24.0.0")
		}
		return exec.Command("echo", "2.20.0")
	}

	rt := docker.NewWithRunner(runner)
	ready, err := rt.Ready(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !ready.Docker {
		t.Fatal("expected docker ready")
	}
	if !ready.Compose {
		t.Fatal("expected compose ready")
	}
	if ready.DockerVersion != "24.0.0" {
		t.Fatalf("unexpected docker version: %s", ready.DockerVersion)
	}
	if ready.ComposeVersion != "2.20.0" {
		t.Fatalf("unexpected compose version: %s", ready.ComposeVersion)
	}
}

func TestReadyDockerUnavailable(t *testing.T) {
	runner := func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.Command("false")
	}

	rt := docker.NewWithRunner(runner)
	ready, err := rt.Ready(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if ready.Docker {
		t.Fatal("expected docker not ready")
	}
	if ready.Compose {
		t.Fatal("expected compose not ready")
	}
	if len(ready.Errors) == 0 {
		t.Fatal("expected errors")
	}
}

func TestReadyDockerComposeFallback(t *testing.T) {
	var calls int
	runner := func(ctx context.Context, name string, args ...string) *exec.Cmd {
		calls++
		if name == "docker" && len(args) == 3 && args[0] == "version" {
			return exec.Command("echo", "24.0.0")
		}
		if name == "docker" && len(args) == 3 && args[0] == "compose" {
			return exec.Command("false")
		}
		if name == "docker-compose" {
			return exec.Command("echo", "1.29.0")
		}
		return exec.Command("false")
	}

	rt := docker.NewWithRunner(runner)
	ready, err := rt.Ready(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !ready.Compose {
		t.Fatal("expected compose ready via fallback")
	}
	if ready.ComposeVersion != "1.29.0" {
		t.Fatalf("unexpected compose version: %s", ready.ComposeVersion)
	}
}
