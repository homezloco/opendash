package noop

import (
	"context"
	"errors"

	"github.com/opendash-project/opendash/internal/runtime"
)

type Runtime struct{}

func New() *Runtime {
	return &Runtime{}
}

func (r *Runtime) Ready(ctx context.Context) (*runtime.Readiness, error) {
	return &runtime.Readiness{Docker: false, Compose: false, Errors: []string{"runtime disabled"}}, nil
}

func (r *Runtime) Install(ctx context.Context, appID, projectName, installPath, composePath string, composeData, env []byte) (*runtime.ComposeCommandResult, error) {
	return nil, errors.New("docker runtime is disabled")
}

func (r *Runtime) Start(ctx context.Context, projectName, composePath string) (*runtime.ComposeCommandResult, error) {
	return nil, errors.New("docker runtime is disabled")
}

func (r *Runtime) Stop(ctx context.Context, projectName, composePath string) (*runtime.ComposeCommandResult, error) {
	return nil, errors.New("docker runtime is disabled")
}

func (r *Runtime) Restart(ctx context.Context, projectName, composePath string) (*runtime.ComposeCommandResult, error) {
	return nil, errors.New("docker runtime is disabled")
}

func (r *Runtime) LogTail(ctx context.Context, projectName, composePath string, lines int) (*runtime.ComposeCommandResult, error) {
	return nil, errors.New("docker runtime is disabled")
}

func (r *Runtime) UninstallPreview(ctx context.Context, projectName, composePath string) (*runtime.ComposePreview, error) {
	return nil, errors.New("docker runtime is disabled")
}

func (r *Runtime) Uninstall(ctx context.Context, projectName, composePath string) (*runtime.ComposeCommandResult, error) {
	return nil, errors.New("docker runtime is disabled")
}

func (r *Runtime) Observe(ctx context.Context, projectName, composePath string) (*runtime.ObservedState, error) {
	return &runtime.ObservedState{}, nil
}
