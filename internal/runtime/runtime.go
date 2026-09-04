package runtime

import "context"

type Readiness struct {
	Docker         bool     `json:"docker"`
	Compose        bool     `json:"compose"`
	DockerVersion  string   `json:"dockerVersion"`
	ComposeVersion string   `json:"composeVersion"`
	Errors         []string `json:"errors"`
}

type ComposeCommandResult struct {
	Output string `json:"output,omitempty"`
}

type ComposePreview struct {
	ProjectName  string   `json:"projectName"`
	Containers   []string `json:"containers"`
	Images       []string `json:"images"`
	Volumes      []string `json:"volumes"`
	Networks     []string `json:"networks"`
	DataRetained bool     `json:"dataRetained"`
}

// ObservedService describes a single Compose service as seen by the runtime.
type ObservedService struct {
	Name   string `json:"name"`
	State  string `json:"state"`
	Health string `json:"health"`
}

// ObservedState is the live state of a Compose project returned by Observe.
type ObservedState struct {
	Running   bool              `json:"running"`
	Healthy   bool              `json:"healthy"`
	Services  []ObservedService `json:"services"`
	HostPorts map[int]int       `json:"hostPorts"`
}

type Runtime interface {
	Ready(ctx context.Context) (*Readiness, error)
	Install(ctx context.Context, appID, projectName, installPath, composePath string, composeData, env []byte) (*ComposeCommandResult, error)
	Start(ctx context.Context, projectName, composePath string) (*ComposeCommandResult, error)
	Stop(ctx context.Context, projectName, composePath string) (*ComposeCommandResult, error)
	Restart(ctx context.Context, projectName, composePath string) (*ComposeCommandResult, error)
	LogTail(ctx context.Context, projectName, composePath string, lines int) (*ComposeCommandResult, error)
	UninstallPreview(ctx context.Context, projectName, composePath string) (*ComposePreview, error)
	Uninstall(ctx context.Context, projectName, composePath string) (*ComposeCommandResult, error)
	Observe(ctx context.Context, projectName, composePath string) (*ObservedState, error)
}
