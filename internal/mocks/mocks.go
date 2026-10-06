// Hand-rolled mocks used by the unit-test suite. Each interface in the
// project gets a corresponding mock so `go test -short ./...` runs without a
// running Docker daemon.
package mocks

import (
	"context"
	"errors"
	"io"
	"os"
	"sync"
	"time"

	coreruntime "github.com/peternicholls/stageserve/core/runtime"
	"github.com/peternicholls/stageserve/core/state"
	"github.com/peternicholls/stageserve/infra/compose"
	"github.com/peternicholls/stageserve/infra/docker"
	"github.com/peternicholls/stageserve/infra/gateway"
	"github.com/peternicholls/stageserve/platform/ports"
)

// --- DockerClient ---

type Docker struct {
	mu           sync.Mutex
	Networks     map[string]bool
	NetworkErr   error
	Containers   []docker.Container
	WaitErr      error
	WaitCalled   []string
	ExecCalls    []docker.ExecOptions
	ExecOutput   string
	ExecErr      error
	LogsReader   io.ReadCloser
	LogsErr      error
	ListErr      error
	AvailableErr error
}

func NewDocker() *Docker { return &Docker{Networks: map[string]bool{"default": true}} }

func (m *Docker) Available(ctx context.Context) error { return m.AvailableErr }

func (m *Docker) NetworkExists(ctx context.Context, name string) (bool, error) {
	if m.NetworkErr != nil {
		return false, m.NetworkErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.Networks[name], nil
}

func (m *Docker) CreateNetwork(ctx context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Networks[name] = true
	return nil
}

func (m *Docker) RemoveNetwork(ctx context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.Networks, name)
	return nil
}

func (m *Docker) ListContainersByLabel(ctx context.Context, labels map[string]string) ([]docker.Container, error) {
	if m.ListErr != nil {
		return nil, m.ListErr
	}
	out := make([]docker.Container, 0, len(m.Containers))
	for _, c := range m.Containers {
		match := true
		for k, v := range labels {
			got, ok := c.Labels[k]
			if !ok || (v != "" && got != v) {
				match = false
				break
			}
		}
		if match {
			out = append(out, c)
		}
	}
	return out, nil
}

func (m *Docker) WaitHealthy(ctx context.Context, project string, timeout time.Duration) error {
	m.mu.Lock()
	m.WaitCalled = append(m.WaitCalled, project)
	m.mu.Unlock()
	return m.WaitErr
}

func (m *Docker) ContainerLogs(ctx context.Context, id string, follow bool) (docker.LogStream, error) {
	if m.LogsErr != nil {
		return nil, m.LogsErr
	}
	return m.LogsReader, nil
}

func (m *Docker) Exec(ctx context.Context, opts docker.ExecOptions) (string, error) {
	m.mu.Lock()
	m.ExecCalls = append(m.ExecCalls, opts)
	m.mu.Unlock()
	return m.ExecOutput, m.ExecErr
}

// --- Composer ---

type Composer struct {
	mu           sync.Mutex
	UpCalls      []compose.UpOptions
	DownCalls    []compose.DownOptions
	RestartCalls []compose.RestartOptions
	UpErr        error
	UpErrOnCall  int
	DownErr      error
	RestartErr   error
	LogsErr      error
	ExecErr      error
}

func NewComposer() *Composer { return &Composer{} }

func (m *Composer) Up(ctx context.Context, opts compose.UpOptions) error {
	m.mu.Lock()
	m.UpCalls = append(m.UpCalls, opts)
	callNumber := len(m.UpCalls)
	m.mu.Unlock()
	if m.UpErrOnCall > 0 && callNumber != m.UpErrOnCall {
		return nil
	}
	return m.UpErr
}

func (m *Composer) Down(ctx context.Context, opts compose.DownOptions) error {
	m.mu.Lock()
	m.DownCalls = append(m.DownCalls, opts)
	m.mu.Unlock()
	return m.DownErr
}

func (m *Composer) Restart(ctx context.Context, opts compose.RestartOptions) error {
	m.mu.Lock()
	m.RestartCalls = append(m.RestartCalls, opts)
	m.mu.Unlock()
	return m.RestartErr
}

func (m *Composer) Logs(ctx context.Context, opts compose.LogsOptions) error { return m.LogsErr }

func (m *Composer) Exec(ctx context.Context, opts compose.ExecOptions) error { return m.ExecErr }

// --- Runtime manager ---

// Runtime adapts the existing narrow test doubles to the vendor-neutral core
// contract. Production code uses the Apple Container manager directly.
type Runtime struct {
	Docker  *Docker
	Compose *Composer
}

func NewRuntime(dockerClient *Docker, composer *Composer) *Runtime {
	return &Runtime{Docker: dockerClient, Compose: composer}
}

func (m *Runtime) Name() coreruntime.BackendName { return coreruntime.BackendAppleContainer }
func (m *Runtime) Capabilities() coreruntime.Capabilities {
	return coreruntime.Capabilities{MultiService: true, Networks: true, Volumes: true, HealthChecks: true, Logs: true, Exec: true, Restart: true, DebugProfile: true, SharedGateway: true}
}
func (m *Runtime) Available(ctx context.Context) error { return m.Docker.Available(ctx) }
func (m *Runtime) NetworkExists(ctx context.Context, name string) (bool, error) {
	return m.Docker.NetworkExists(ctx, name)
}
func (m *Runtime) CreateNetwork(ctx context.Context, name string) error {
	return m.Docker.CreateNetwork(ctx, name)
}
func (m *Runtime) RemoveNetwork(ctx context.Context, name string) error {
	return m.Docker.RemoveNetwork(ctx, name)
}
func (m *Runtime) Start(ctx context.Context, opts coreruntime.StartOptions) error {
	if err := opts.Ownership.Validate(); err != nil {
		return err
	}
	if opts.RecordResource == nil {
		return errors.New("resource recorder is required")
	}
	return m.Compose.Up(ctx, compose.UpOptions{ProjectDir: opts.ProjectDir, ComposeFile: opts.Definition, ProjectName: opts.ProjectName, EnvFile: opts.EnvFile, Env: opts.Env, Profiles: opts.Profiles, WaitTimeout: opts.WaitTimeout, Detach: true, NoDeps: opts.NoDeps, ForceRecreate: opts.ForceRecreate, Services: opts.Services})
}
func (m *Runtime) Stop(ctx context.Context, opts coreruntime.StopOptions) error {
	if err := opts.Ownership.Validate(); err != nil {
		return err
	}
	if opts.RecordResource == nil {
		return errors.New("resource recorder is required")
	}
	return m.Compose.Down(ctx, compose.DownOptions{ProjectDir: opts.ProjectDir, ComposeFile: opts.Definition, ProjectName: opts.ProjectName, EnvFile: opts.EnvFile, Env: opts.Env, RemoveVolumes: opts.RemoveVolumes})
}
func (m *Runtime) Restart(ctx context.Context, opts coreruntime.RestartOptions) error {
	if err := opts.Ownership.Validate(); err != nil {
		return err
	}
	if opts.RecordResource == nil {
		return errors.New("resource recorder is required")
	}
	return m.Compose.Restart(ctx, compose.RestartOptions{ProjectDir: opts.ProjectDir, ComposeFile: opts.Definition, ProjectName: opts.ProjectName, EnvFile: opts.EnvFile, Env: opts.Env, Service: opts.Service})
}
func (m *Runtime) Logs(ctx context.Context, opts coreruntime.LogsOptions) error {
	return m.Compose.Logs(ctx, compose.LogsOptions{ProjectDir: opts.ProjectDir, ComposeFile: opts.Definition, ProjectName: opts.ProjectName, EnvFile: opts.EnvFile, Env: opts.Env, Service: opts.Service, Follow: opts.Follow})
}
func (m *Runtime) Exec(ctx context.Context, opts coreruntime.ExecOptions) (string, error) {
	containers, err := m.Docker.ListContainersByLabel(ctx, map[string]string{"com.docker.compose.project": opts.ProjectName, "com.docker.compose.service": opts.Service})
	if err != nil {
		return "", err
	}
	if len(containers) == 0 {
		return "", errors.New("service not found")
	}
	return m.Docker.Exec(ctx, docker.ExecOptions{ContainerID: containers[0].ID, Cmd: opts.Command, WorkingDir: opts.WorkingDir})
}
func (m *Runtime) ListServices(ctx context.Context, projectName string) ([]coreruntime.Service, error) {
	containers, err := m.Docker.ListContainersByLabel(ctx, map[string]string{"com.docker.compose.project": projectName})
	if err != nil {
		return nil, err
	}
	services := make([]coreruntime.Service, 0, len(containers))
	for _, container := range containers {
		services = append(services, coreruntime.Service{ID: container.ID, Name: container.Name, Status: container.Status, Service: container.Service, Project: container.Project, Labels: container.Labels})
	}
	return services, nil
}
func (m *Runtime) WaitHealthy(ctx context.Context, projectName string, timeout time.Duration) error {
	return m.Docker.WaitHealthy(ctx, projectName, timeout)
}
func (m *Runtime) ServiceLogs(ctx context.Context, serviceID string, follow bool) (coreruntime.LogStream, error) {
	return m.Docker.ContainerLogs(ctx, serviceID, follow)
}

// --- GatewayManager ---

type Gateway struct {
	mu        sync.Mutex
	Routes    []gateway.Route
	Probe     string
	Host      string
	WriteErr  error
	LastInput gateway.RenderInput
}

func NewGateway() *Gateway { return &Gateway{Probe: "stageserve-no-route", Host: "localhost"} }

func (m *Gateway) ConfigPath() string { return "/tmp/mock-gateway.conf" }

func (m *Gateway) WriteConfig(input gateway.RenderInput) (string, string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Routes = append([]gateway.Route(nil), input.Routes...)
	m.LastInput = input
	if m.WriteErr != nil {
		return "", "", m.WriteErr
	}
	return m.Probe, m.Host, nil
}

func (m *Gateway) AddRoute(r gateway.Route, current []gateway.Route) (string, string, error) {
	merged := append([]gateway.Route(nil), current...)
	merged = append(merged, r)
	return m.WriteConfig(gateway.RenderInput{Routes: merged, PreferredSlug: r.Slug})
}

func (m *Gateway) RemoveRoute(slug string, current []gateway.Route) (string, string, error) {
	merged := []gateway.Route{}
	for _, r := range current {
		if r.Slug != slug {
			merged = append(merged, r)
		}
	}
	return m.WriteConfig(gateway.RenderInput{Routes: merged})
}

// --- StateStore ---

type State struct {
	mu             sync.Mutex
	Records        map[string]state.Record
	StateDirVal    string
	OwnershipStore *state.Store
	SaveErr        error
	RegistryErr    error
}

func NewState() *State {
	dir, err := os.MkdirTemp("", "stageserve-state-mock-")
	if err != nil {
		panic(err)
	}
	store, err := state.NewStore(dir)
	if err != nil {
		panic(err)
	}
	return &State{Records: map[string]state.Record{}, StateDirVal: dir, OwnershipStore: store}
}

func (m *State) Save(rec state.Record) error {
	if m.SaveErr != nil {
		return m.SaveErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Records[rec.Project.Slug] = rec
	return nil
}

func (m *State) Load(slug string) (state.Record, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	rec, ok := m.Records[slug]
	if !ok {
		return state.Record{}, state.ErrNotFound
	}
	return rec, nil
}

func (m *State) Remove(slug string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.Records, slug)
	return nil
}

func (m *State) Registry() ([]state.RegistryRow, error) {
	if m.RegistryErr != nil {
		return nil, m.RegistryErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	rows := []state.RegistryRow{}
	for _, rec := range m.Records {
		rows = append(rows, state.RegistryRow{
			ProjectID: rec.ProjectID, InstallationID: rec.InstallationID,
			Slug:            rec.Project.Slug,
			AttachmentState: rec.AttachmentState,
			Name:            rec.Project.Name,
			Hostname:        rec.Project.Hostname,
			ComposeProject:  rec.Project.ComposeProjectName,
			RuntimeNetwork:  rec.Project.RuntimeNetwork,
			DatabaseVolume:  rec.Project.DatabaseVolume,
			WebNetworkAlias: rec.Project.WebNetworkAlias,
			MySQLPort:       rec.Project.MySQL.Port,
			PMAPort:         rec.Project.MySQL.PMAPort,
		})
	}
	return rows, nil
}

func (m *State) StateFileForSelector(selector string) (state.Record, string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for slug, rec := range m.Records {
		p := rec.Project
		if selector == slug || selector == p.Name || selector == p.Hostname || selector == p.Dir {
			return rec, "", nil
		}
	}
	return state.Record{}, "", state.ErrNotFound
}

func (m *State) StateDir() string { return m.StateDirVal }

// --- PortAllocator ---

type Ports struct {
	Out ports.Allocation
	Err error
}

func NewPorts(a ports.Allocation) *Ports { return &Ports{Out: a} }

func (m *Ports) Allocate(req ports.Request, registry []state.RegistryRow) (ports.Allocation, error) {
	if m.Err != nil {
		return ports.Allocation{}, m.Err
	}
	out := m.Out
	if out.MySQLPort == 0 && req.MySQLPort != 0 {
		out.MySQLPort = req.MySQLPort
	}
	if out.PMAPort == 0 && req.PMAPort != 0 {
		out.PMAPort = req.PMAPort
	}
	return out, nil
}

// errStub keeps the linter happy if we ever extend the package without using errors.
var _ = errors.New

// Seed records explicit ownership for lifecycle fixtures; Save never adopts IDs.
func (m *State) Seed(rec state.Record) error {
	id, err := m.CreateIdentity(rec.Project.Dir, rec.Project.Slug)
	if err != nil {
		return err
	}
	rec.ProjectID = id.ProjectID
	rec.InstallationID = id.InstallationID
	rec.SchemaVersion = state.SchemaVersion
	rec.Project.ComposeProjectName = "stage-" + id.ProjectID
	rec.Project.DatabaseVolume = rec.Project.ComposeProjectName + "-db-data"
	if err := m.OwnershipStore.Save(rec); err != nil {
		return err
	}
	return m.Save(rec)
}
func (m *State) CreateIdentity(path, slug string) (state.Identity, error) {
	return m.OwnershipStore.CreateIdentity(path, slug)
}
func (m *State) LoadIdentity(id string) (state.Identity, error) {
	return m.OwnershipStore.LoadIdentity(id)
}
func (m *State) IdentityForPath(path string) (state.Identity, error) {
	return m.OwnershipStore.IdentityForPath(path)
}
func (m *State) SaveIdentity(id state.Identity, rev uint64) error {
	return m.OwnershipStore.SaveIdentity(id, rev)
}
func (m *State) InstallationIdentity(path string) (state.Identity, error) {
	return m.OwnershipStore.InstallationIdentity(path)
}
func (m *State) BeginOperation(id string, rev uint64, kind, hash string) (state.Operation, error) {
	return m.OwnershipStore.BeginOperation(id, rev, kind, hash)
}
func (m *State) SaveOperation(op state.Operation) error { return m.OwnershipStore.SaveOperation(op) }
func (m *State) PendingOperations(id string) ([]state.Operation, error) {
	return m.OwnershipStore.PendingOperations(id)
}
func (m *State) RecoverOperation(id string) error { return m.OwnershipStore.RecoverOperation(id) }
func (m *State) CommitInstallationOperation(op state.Operation, id state.Identity, rev uint64) error {
	return m.OwnershipStore.CommitInstallationOperation(op, id, rev)
}
func (m *State) CommitOperation(op state.Operation, id state.Identity, rec *state.Record, rev uint64) error {
	if m.SaveErr != nil {
		return m.SaveErr
	}
	if err := m.OwnershipStore.CommitOperation(op, id, rec, rev); err != nil {
		return err
	}
	if rec == nil {
		return m.Remove(id.Slug)
	}
	copy := *rec
	copy.SchemaVersion = state.SchemaVersion
	return m.Save(copy)
}

var _ state.LifecycleStore = (*State)(nil)
