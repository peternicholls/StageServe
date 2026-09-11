// Package runtime defines StageServe's vendor-neutral runtime contract.
package runtime

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"
)

// BackendName identifies a runtime implementation without exposing its API.
type BackendName string

const (
	BackendAppleContainer BackendName = "apple-container"
)

// ParseBackend validates an operator-provided backend name.
func ParseBackend(value string) (BackendName, error) {
	backend := BackendName(strings.ToLower(strings.TrimSpace(value)))
	if backend == "" {
		return BackendAppleContainer, nil
	}
	switch backend {
	case BackendAppleContainer:
		return backend, nil
	default:
		return "", fmt.Errorf("unsupported StageServe runtime %q: use apple-container", value)
	}
}

// Capabilities describes runtime features that can be checked before side effects.
type Capabilities struct {
	MultiService    bool
	Networks        bool
	Volumes         bool
	HealthChecks    bool
	Logs            bool
	Exec            bool
	Restart         bool
	DebugProfile    bool
	SharedGateway   bool
	BareNetworkDNS  bool
	DockerEngineAPI bool
}

// Service is StageServe's observed view of one backend service instance.
type Service struct {
	ID      string
	Name    string
	Status  string
	Service string
	Project string
	Address string
	Labels  map[string]string
}

// StartOptions describes a project or shared-runtime start operation.
type StartOptions struct {
	ProjectDir    string
	Definition    string
	ProjectName   string
	EnvFile       string
	Env           []string
	Profiles      []string
	WaitTimeout   time.Duration
	NoDeps        bool
	ForceRecreate bool
	Services      []string
}

// StopOptions describes a project stop operation.
type StopOptions struct {
	ProjectDir    string
	Definition    string
	ProjectName   string
	EnvFile       string
	Env           []string
	RemoveVolumes bool
}

// RestartOptions describes a service restart operation.
type RestartOptions struct {
	ProjectDir  string
	Definition  string
	ProjectName string
	EnvFile     string
	Env         []string
	Service     string
}

// LogsOptions describes a service or project log stream.
type LogsOptions struct {
	ProjectDir  string
	Definition  string
	ProjectName string
	EnvFile     string
	Env         []string
	Service     string
	Follow      bool
}

// ExecOptions describes a one-shot command in a service.
type ExecOptions struct {
	ProjectDir  string
	Definition  string
	ProjectName string
	EnvFile     string
	Env         []string
	Service     string
	Command     []string
	WorkingDir  string
}

// LogStream is the reader/closer surface used for direct backend log streams.
type LogStream interface {
	io.Reader
	io.Closer
}

// Manager is the runtime boundary used by lifecycle and observability code.
type Manager interface {
	Name() BackendName
	Capabilities() Capabilities
	Available(ctx context.Context) error
	NetworkExists(ctx context.Context, name string) (bool, error)
	CreateNetwork(ctx context.Context, name string) error
	RemoveNetwork(ctx context.Context, name string) error
	Start(ctx context.Context, opts StartOptions) error
	Stop(ctx context.Context, opts StopOptions) error
	Restart(ctx context.Context, opts RestartOptions) error
	Logs(ctx context.Context, opts LogsOptions) error
	Exec(ctx context.Context, opts ExecOptions) (string, error)
	ListServices(ctx context.Context, projectName string) ([]Service, error)
	WaitHealthy(ctx context.Context, projectName string, timeout time.Duration) error
	ServiceLogs(ctx context.Context, serviceID string, follow bool) (LogStream, error)
}
