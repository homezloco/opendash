package docker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/opendash-project/opendash/internal/runtime"
)

type CommandRunner func(ctx context.Context, name string, args ...string) *exec.Cmd

type Runtime struct {
	runner CommandRunner
	root   string
}

func New() *Runtime {
	return NewWithRoot("./data/apps")
}

func NewWithRoot(root string) *Runtime {
	return NewWithRootAndRunner(root, exec.CommandContext)
}

func NewWithRootAndRunner(root string, runner CommandRunner) *Runtime {
	if runner == nil {
		runner = exec.CommandContext
	}
	// Ensure the managed root exists and is not a symlink before we accept it.
	_ = os.MkdirAll(root, 0700)
	return &Runtime{runner: runner, root: root}
}

// NewWithRunner preserves the original test constructor using the default managed root.
func NewWithRunner(runner CommandRunner) *Runtime {
	return NewWithRootAndRunner("./data/apps", runner)
}

func (r *Runtime) Ready(ctx context.Context) (*runtime.Readiness, error) {
	readiness := &runtime.Readiness{Errors: []string{}}

	dockerCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	out, err := run(dockerCtx, r.runner, "docker", "version", "--format", "{{.Server.Version}}")
	if err != nil {
		readiness.Errors = append(readiness.Errors, fmt.Sprintf("docker unavailable: %v", err))
	} else {
		readiness.Docker = true
		readiness.DockerVersion = strings.TrimSpace(string(out))
	}

	composeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	out, err = run(composeCtx, r.runner, "docker", "compose", "version", "--short")
	if err == nil {
		readiness.Compose = true
		readiness.ComposeVersion = strings.TrimSpace(string(out))
	} else {
		out, err = run(composeCtx, r.runner, "docker-compose", "version", "--short")
		if err == nil {
			readiness.Compose = true
			readiness.ComposeVersion = strings.TrimSpace(string(out))
		} else {
			readiness.Errors = append(readiness.Errors, fmt.Sprintf("docker compose unavailable: %v", err))
		}
	}

	return readiness, nil
}

func (r *Runtime) Install(ctx context.Context, appID, projectName, installPath, composePath string, composeData, env []byte) (*runtime.ComposeCommandResult, error) {
	if err := validateAppID(appID); err != nil {
		return nil, err
	}
	if err := validateProjectName(projectName); err != nil {
		return nil, err
	}
	if err := r.validateManagedPath(installPath); err != nil {
		return nil, err
	}
	if !filepath.IsAbs(composePath) {
		composePath = filepath.Join(installPath, composePath)
	}
	if err := r.validateManagedPath(composePath); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(composePath), 0700); err != nil {
		return nil, fmt.Errorf("create compose directory: %w", err)
	}
	if err := r.validateManagedPath(filepath.Dir(composePath)); err != nil {
		return nil, err
	}

	envPath := composePath + ".env"
	if err := writePrivateFile(composePath, composeData); err != nil {
		return nil, fmt.Errorf("write compose file: %w", err)
	}
	if err := writePrivateFile(envPath, env); err != nil {
		return nil, fmt.Errorf("write env file: %w", err)
	}

	installCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	out, err := run(installCtx, r.runner, "docker", "compose", "-f", composePath, "-p", projectName, "--env-file", envPath, "up", "-d", "--quiet-pull")
	if err != nil {
		return nil, fmt.Errorf("docker compose up failed: %w: %s", err, string(out))
	}
	return &runtime.ComposeCommandResult{Output: string(out)}, nil
}

func (r *Runtime) Start(ctx context.Context, projectName, composePath string) (*runtime.ComposeCommandResult, error) {
	return r.composeAction(ctx, projectName, composePath, "start", 2*time.Minute)
}

func (r *Runtime) Stop(ctx context.Context, projectName, composePath string) (*runtime.ComposeCommandResult, error) {
	return r.composeAction(ctx, projectName, composePath, "stop", 2*time.Minute)
}

func (r *Runtime) Restart(ctx context.Context, projectName, composePath string) (*runtime.ComposeCommandResult, error) {
	return r.composeAction(ctx, projectName, composePath, "restart", 2*time.Minute)
}

func (r *Runtime) LogTail(ctx context.Context, projectName, composePath string, lines int) (*runtime.ComposeCommandResult, error) {
	if lines < 1 || lines > 10000 {
		return nil, fmt.Errorf("lines must be between 1 and 10000")
	}
	if err := r.validateManagedCompose(composePath); err != nil {
		return nil, err
	}
	logCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	out, err := run(logCtx, r.runner, "docker", "compose", "-f", composePath, "-p", projectName, "logs", "--no-color", "--tail", fmt.Sprintf("%d", lines))
	if err != nil {
		return nil, fmt.Errorf("docker compose logs failed: %w: %s", err, string(out))
	}
	return &runtime.ComposeCommandResult{Output: string(out)}, nil
}

func (r *Runtime) UninstallPreview(ctx context.Context, projectName, composePath string) (*runtime.ComposePreview, error) {
	if err := r.validateManagedCompose(composePath); err != nil {
		return nil, err
	}
	cfgCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	out, err := run(cfgCtx, r.runner, "docker", "compose", "-f", composePath, "-p", projectName, "config", "--format", "json")
	if err != nil {
		return nil, fmt.Errorf("docker compose config failed: %w: %s", err, string(out))
	}
	return parseComposeConfig(projectName, out)
}

func (r *Runtime) Uninstall(ctx context.Context, projectName, composePath string) (*runtime.ComposeCommandResult, error) {
	if err := r.validateManagedCompose(composePath); err != nil {
		return nil, err
	}
	downCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	// "down" without --volumes leaves named volumes intact, giving the data-retaining
	// uninstall behaviour requested by users.
	out, err := run(downCtx, r.runner, "docker", "compose", "-f", composePath, "-p", projectName, "down")
	if err != nil {
		return nil, fmt.Errorf("docker compose down failed: %w: %s", err, string(out))
	}
	return &runtime.ComposeCommandResult{Output: string(out)}, nil
}

// Observe returns the live state of a Compose project by inspecting running
// containers. It maps container ports to the published host ports so the API can
// present accurate endpoint URLs.
func (r *Runtime) Observe(ctx context.Context, projectName, composePath string) (*runtime.ObservedState, error) {
	if err := r.validateManagedCompose(composePath); err != nil {
		return nil, err
	}
	psCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	out, err := run(psCtx, r.runner, "docker", "compose", "-f", composePath, "-p", projectName, "ps", "--format", "json")
	if err != nil {
		return nil, fmt.Errorf("docker compose ps failed: %w: %s", err, string(out))
	}
	return parseObservedState(out)
}

func (r *Runtime) composeAction(ctx context.Context, projectName, composePath, action string, timeout time.Duration) (*runtime.ComposeCommandResult, error) {
	if err := validateProjectName(projectName); err != nil {
		return nil, err
	}
	if err := r.validateManagedCompose(composePath); err != nil {
		return nil, err
	}
	actCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	out, err := run(actCtx, r.runner, "docker", "compose", "-f", composePath, "-p", projectName, action)
	if err != nil {
		return nil, fmt.Errorf("docker compose %s failed: %w: %s", action, err, string(out))
	}
	return &runtime.ComposeCommandResult{Output: string(out)}, nil
}

func (r *Runtime) validateManagedCompose(composePath string) error {
	if !filepath.IsAbs(composePath) {
		return fmt.Errorf("compose path must be absolute")
	}
	return r.validateManagedPath(composePath)
}

func (r *Runtime) validateManagedPath(p string) error {
	if r.root == "" {
		return errors.New("managed root is not configured")
	}
	absRoot, err := filepath.Abs(filepath.Clean(r.root))
	if err != nil {
		return fmt.Errorf("resolve managed root: %w", err)
	}
	realRoot, err := resolvePath(absRoot)
	if err != nil {
		return fmt.Errorf("resolve managed root: %w", err)
	}
	if realRoot != absRoot {
		return fmt.Errorf("managed root must not be a symbolic link")
	}
	if !filepath.IsAbs(p) {
		p, err = filepath.Abs(p)
		if err != nil {
			return fmt.Errorf("resolve path: %w", err)
		}
	}
	realP, err := resolvePath(p)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}
	rel, err := filepath.Rel(realRoot, realP)
	if err != nil || strings.HasPrefix(rel, "..") || rel == ".." {
		return fmt.Errorf("path %q is outside managed root %q", p, realRoot)
	}
	return nil
}

// resolvePath returns the real absolute path by following symlinks in existing
// components. Non-existent leafs are resolved against their real parent path so
// that planned install paths are still accepted.
func resolvePath(p string) (string, error) {
	info, err := os.Lstat(p)
	if err != nil {
		if os.IsNotExist(err) {
			parent := filepath.Dir(p)
			if parent == p {
				return "", fmt.Errorf("path does not exist: %s", p)
			}
			realParent, err := resolvePath(parent)
			if err != nil {
				return "", err
			}
			return filepath.Join(realParent, filepath.Base(p)), nil
		}
		return "", err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return filepath.EvalSymlinks(p)
	}
	return filepath.EvalSymlinks(p)
}

func writePrivateFile(path string, data []byte) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	if _, err := file.Write(data); err != nil {
		file.Close()
		return err
	}
	if err := file.Chmod(0600); err != nil {
		file.Close()
		return err
	}
	return file.Close()
}

func run(ctx context.Context, runner CommandRunner, name string, args ...string) ([]byte, error) {
	cmd := runner(ctx, name, args...)
	return cmd.CombinedOutput()
}

var (
	appIDPattern       = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	projectNamePattern = regexp.MustCompile(`^opendash-[a-zA-Z0-9_-]+$`)
)

func validateAppID(id string) error {
	if !appIDPattern.MatchString(id) {
		return fmt.Errorf("invalid app id %q", id)
	}
	return nil
}

func validateProjectName(name string) error {
	if !projectNamePattern.MatchString(name) {
		return fmt.Errorf("project name %q must start with opendash- and contain only alphanumeric, underscore or hyphen characters", name)
	}
	return nil
}

func parseComposeConfig(projectName string, data []byte) (*runtime.ComposePreview, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse compose config: %w", err)
	}
	preview := &runtime.ComposePreview{
		ProjectName:  projectName,
		DataRetained: true,
	}
	if svcRaw, ok := raw["services"]; ok {
		var services map[string]composeService
		if err := json.Unmarshal(svcRaw, &services); err == nil {
			for name, svc := range services {
				preview.Containers = append(preview.Containers, projectName+"-"+name)
				if svc.Image != "" {
					preview.Images = append(preview.Images, svc.Image)
				}
			}
		}
	}
	if volRaw, ok := raw["volumes"]; ok {
		var volumes map[string]any
		if err := json.Unmarshal(volRaw, &volumes); err == nil {
			for name := range volumes {
				preview.Volumes = append(preview.Volumes, name)
			}
		}
	}
	if netRaw, ok := raw["networks"]; ok {
		var networks map[string]any
		if err := json.Unmarshal(netRaw, &networks); err == nil {
			for name := range networks {
				preview.Networks = append(preview.Networks, name)
			}
		}
	}
	return preview, nil
}

type composeService struct {
	Image string `json:"image"`
}

type composePsItem struct {
	Name       string      `json:"Name"`
	Service    string      `json:"Service"`
	State      string      `json:"State"`
	Health     string      `json:"Health"`
	Publishers []publisher `json:"Publishers"`
}

type publisher struct {
	URL           string `json:"URL"`
	TargetPort    int    `json:"TargetPort"`
	PublishedPort int    `json:"PublishedPort"`
}

func parseObservedState(data []byte) (*runtime.ObservedState, error) {
	var items []composePsItem
	if err := json.Unmarshal(data, &items); err != nil {
		// Some older Compose versions return a single object; try that as a fallback.
		var single composePsItem
		if err2 := json.Unmarshal(data, &single); err2 != nil {
			return nil, fmt.Errorf("parse compose ps output: %w", err)
		}
		items = []composePsItem{single}
	}

	state := &runtime.ObservedState{
		HostPorts: make(map[int]int),
	}
	allRunning := len(items) > 0
	allHealthy := len(items) > 0
	for _, item := range items {
		state.Services = append(state.Services, runtime.ObservedService{
			Name:   item.Service,
			State:  item.State,
			Health: item.Health,
		})
		if item.State != "running" {
			allRunning = false
		}
		if item.Health != "" && item.Health != "healthy" {
			allHealthy = false
		}
		for _, pub := range item.Publishers {
			if pub.PublishedPort > 0 {
				state.HostPorts[pub.TargetPort] = pub.PublishedPort
			}
		}
	}
	state.Running = allRunning
	state.Healthy = allRunning && allHealthy
	return state, nil
}
