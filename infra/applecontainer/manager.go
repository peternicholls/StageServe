package applecontainer

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	coreruntime "github.com/peternicholls/stageserve/core/runtime"
)

// Manager implements StageServe runtime operations with Apple's container CLI.
type Manager struct {
	Runner Runner
	Stdout io.Writer

	mu        sync.Mutex
	manifests map[string]resolvedManifest
}

type resolvedManifest struct {
	Definition string
	Env        map[string]string
	Manifest   manifest
	Profiles   []string
}

func NewManager(runner Runner) *Manager {
	if runner == nil {
		runner = CommandRunner{}
	}
	return &Manager{Runner: runner, Stdout: os.Stdout, manifests: map[string]resolvedManifest{}}
}

func (m *Manager) Name() coreruntime.BackendName { return coreruntime.BackendAppleContainer }

func (m *Manager) Capabilities() coreruntime.Capabilities {
	return coreruntime.Capabilities{
		MultiService: true, Networks: true, Volumes: true, HealthChecks: true,
		Logs: true, Exec: true, Restart: true, DebugProfile: true,
		SharedGateway: true, BareNetworkDNS: false, DockerEngineAPI: false,
	}
}

func (m *Manager) Available(ctx context.Context) error {
	_, err := m.Runner.Run(ctx, "system", "status", "--format", "json")
	return err
}

func (m *Manager) NetworkExists(ctx context.Context, name string) (bool, error) {
	out, err := m.Runner.Run(ctx, "network", "list", "--quiet")
	if err != nil {
		return false, err
	}
	return containsLine(out, name), nil
}

func (m *Manager) CreateNetwork(ctx context.Context, name string) error {
	_, err := m.Runner.Run(ctx, "network", "create", "--label", "io.stageserve.managed=true", name)
	return err
}

func (m *Manager) RemoveNetwork(ctx context.Context, name string) error {
	exists, err := m.NetworkExists(ctx, name)
	if err != nil || !exists {
		return err
	}
	_, err = m.Runner.Run(ctx, "network", "delete", name)
	return err
}

func (m *Manager) Start(ctx context.Context, opts coreruntime.StartOptions) error {
	definition, err := loadManifest(opts.Definition)
	if err != nil {
		return err
	}
	env, err := environment(opts.EnvFile, opts.Env)
	if err != nil {
		return err
	}
	m.mu.Lock()
	m.manifests[opts.ProjectName] = resolvedManifest{Definition: opts.Definition, Env: env, Manifest: definition, Profiles: append([]string(nil), opts.Profiles...)}
	m.mu.Unlock()

	for _, volume := range definition.Volumes {
		if err := m.ensureVolume(ctx, expand(volume, env)); err != nil {
			return err
		}
	}
	selected := selectedServices(opts.Services)
	started := []string{}
	for _, service := range definition.Services {
		if !profileEnabled(service, opts.Profiles) {
			continue
		}
		if len(selected) > 0 && !selected[service.Name] {
			continue
		}
		name := serviceName(opts.ProjectName, service.Name)
		if opts.ForceRecreate {
			_, _ = m.Runner.Run(ctx, "delete", "--force", name)
		}
		exists, running, err := m.serviceState(ctx, name)
		if err != nil {
			return err
		}
		if exists && running {
			continue
		}
		if exists {
			if _, err := m.Runner.Run(ctx, "start", name); err != nil {
				return err
			}
		} else {
			if err := m.buildService(ctx, opts, service, env); err != nil {
				m.rollback(ctx, started)
				return err
			}
			if _, err := m.Runner.Run(ctx, m.runArgs(opts, service, env)...); err != nil {
				m.rollback(ctx, started)
				return err
			}
		}
		started = append(started, name)
		if service.Health != nil {
			if err := m.waitService(ctx, name, service.Health, opts.WaitTimeout); err != nil {
				m.rollback(ctx, started)
				return err
			}
		}
	}
	return nil
}

func (m *Manager) Stop(ctx context.Context, opts coreruntime.StopOptions) error {
	services, err := m.ListServices(ctx, opts.ProjectName)
	if err != nil {
		return err
	}
	for i := len(services) - 1; i >= 0; i-- {
		_, _ = m.Runner.Run(ctx, "stop", services[i].ID)
		if _, err := m.Runner.Run(ctx, "delete", services[i].ID); err != nil {
			return err
		}
	}
	if opts.RemoveVolumes {
		volume := opts.ProjectName + "-db-data"
		if exists, err := m.volumeExists(ctx, volume); err != nil {
			return err
		} else if exists {
			if _, err := m.Runner.Run(ctx, "volume", "delete", volume); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *Manager) Restart(ctx context.Context, opts coreruntime.RestartOptions) error {
	name := serviceName(opts.ProjectName, opts.Service)
	if _, err := m.Runner.Run(ctx, "stop", name); err != nil {
		return err
	}
	_, err := m.Runner.Run(ctx, "start", name)
	return err
}

func (m *Manager) Logs(ctx context.Context, opts coreruntime.LogsOptions) error {
	stream, err := m.ServiceLogs(ctx, serviceName(opts.ProjectName, opts.Service), opts.Follow)
	if err != nil {
		return err
	}
	defer stream.Close()
	_, err = io.Copy(m.Stdout, stream)
	return err
}

func (m *Manager) Exec(ctx context.Context, opts coreruntime.ExecOptions) (string, error) {
	args := []string{"exec"}
	if opts.WorkingDir != "" {
		args = append(args, "--workdir", opts.WorkingDir)
	}
	args = append(args, serviceName(opts.ProjectName, opts.Service))
	args = append(args, opts.Command...)
	out, err := m.Runner.Run(ctx, args...)
	return strings.TrimSpace(string(out)), err
}

func (m *Manager) ListServices(ctx context.Context, projectName string) ([]coreruntime.Service, error) {
	out, err := m.Runner.Run(ctx, "list", "--all", "--format", "json")
	if err != nil {
		return nil, err
	}
	services, err := parseServices(out, projectName)
	if err != nil {
		return nil, err
	}
	sort.SliceStable(services, func(i, j int) bool { return services[i].Service < services[j].Service })
	return services, nil
}

func (m *Manager) WaitHealthy(ctx context.Context, projectName string, timeout time.Duration) error {
	m.mu.Lock()
	resolved, ok := m.manifests[projectName]
	m.mu.Unlock()
	if !ok {
		return fmt.Errorf("no Apple Container runtime definition loaded for project %s", projectName)
	}
	for _, service := range resolved.Manifest.Services {
		if !profileEnabled(service, resolved.Profiles) {
			continue
		}
		if service.Health == nil {
			continue
		}
		if err := m.waitService(ctx, serviceName(projectName, service.Name), service.Health, timeout); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) ServiceLogs(ctx context.Context, serviceID string, follow bool) (coreruntime.LogStream, error) {
	args := []string{"logs"}
	if follow {
		args = append(args, "--follow")
	}
	args = append(args, serviceID)
	return m.Runner.Stream(ctx, args...)
}

func (m *Manager) runArgs(opts coreruntime.StartOptions, service serviceSpec, env map[string]string) []string {
	args := []string{"run", "--detach", "--name", serviceName(opts.ProjectName, service.Name),
		"--label", "io.stageserve.project=" + opts.ProjectName,
		"--label", "io.stageserve.service=" + service.Name,
		"--network", "default"}
	if service.Workdir != "" {
		args = append(args, "--workdir", expand(service.Workdir, env))
	}
	keys := make([]string, 0, len(service.Environment))
	for key := range service.Environment {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		args = append(args, "--env", key+"="+expand(service.Environment[key], env))
	}
	for _, mount := range service.Mounts {
		args = append(args, "--volume", expand(mount, env))
	}
	for _, port := range service.Ports {
		args = append(args, "--publish", expand(port, env))
	}
	image := expand(service.Image, env)
	if service.Build != nil {
		image = buildTag(opts.ProjectName, service.Name)
	}
	args = append(args, image)
	args = append(args, service.Command...)
	return args
}

func (m *Manager) buildService(ctx context.Context, opts coreruntime.StartOptions, service serviceSpec, env map[string]string) error {
	if service.Build == nil {
		return nil
	}
	base := filepath.Dir(opts.Definition)
	args := []string{"build", "--tag", buildTag(opts.ProjectName, service.Name)}
	keys := make([]string, 0, len(service.Build.Args))
	for key := range service.Build.Args {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		args = append(args, "--build-arg", key+"="+expand(service.Build.Args[key], env))
	}
	if service.Build.File != "" {
		args = append(args, "--file", manifestPath(base, expand(service.Build.File, env)))
	}
	args = append(args, manifestPath(base, expand(service.Build.Context, env)))
	_, err := m.Runner.Run(ctx, args...)
	return err
}

func (m *Manager) waitService(ctx context.Context, name string, health *healthSpec, timeout time.Duration) error {
	if timeout <= 0 {
		timeout = 120 * time.Second
	}
	retries := health.Retries
	if retries <= 0 {
		retries = int(timeout/healthInterval(health)) + 1
	}
	var lastErr error
	for attempt := 0; attempt < retries; attempt++ {
		_, lastErr = m.Runner.Run(ctx, append([]string{"exec", name}, health.Command...)...)
		if lastErr == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(healthInterval(health)):
		}
	}
	return fmt.Errorf("service %s did not become healthy: %w", name, lastErr)
}

func (m *Manager) serviceState(ctx context.Context, name string) (bool, bool, error) {
	services, err := m.ListServices(ctx, "")
	if err != nil {
		return false, false, err
	}
	for _, service := range services {
		if service.ID == name || service.Name == name {
			return true, strings.EqualFold(service.Status, "running"), nil
		}
	}
	return false, false, nil
}

func (m *Manager) ensureVolume(ctx context.Context, name string) error {
	exists, err := m.volumeExists(ctx, name)
	if err != nil || exists {
		return err
	}
	_, err = m.Runner.Run(ctx, "volume", "create", "--label", "io.stageserve.managed=true", name)
	return err
}

func (m *Manager) volumeExists(ctx context.Context, name string) (bool, error) {
	out, err := m.Runner.Run(ctx, "volume", "list", "--quiet")
	if err != nil {
		return false, err
	}
	return containsLine(out, name), nil
}

func (m *Manager) rollback(ctx context.Context, names []string) {
	for i := len(names) - 1; i >= 0; i-- {
		_, _ = m.Runner.Run(ctx, "stop", names[i])
		_, _ = m.Runner.Run(ctx, "delete", names[i])
	}
}

func parseServices(data []byte, projectName string) ([]coreruntime.Service, error) {
	var items []map[string]any
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, fmt.Errorf("parse container list JSON: %w", err)
	}
	prefix := projectName + "-"
	services := []coreruntime.Service{}
	for _, item := range items {
		id := stringField(item, "id", "ID", "name", "Name")
		if projectName != "" && !strings.HasPrefix(id, prefix) {
			continue
		}
		service := strings.TrimPrefix(id, prefix)
		services = append(services, coreruntime.Service{
			ID: id, Name: id, Project: projectName, Service: service,
			Status:  stringField(item, "state", "State", "status", "Status"),
			Address: stringField(item, "ip", "IP", "address", "Address"),
		})
	}
	return services, nil
}

func stringField(item map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := item[key].(string); ok {
			return value
		}
	}
	return ""
}

func serviceName(project, service string) string { return project + "-" + service }
func buildTag(project, service string) string {
	return "stageserve/" + project + "-" + service + ":latest"
}
func selectedServices(values []string) map[string]bool {
	selected := map[string]bool{}
	for _, value := range values {
		selected[value] = true
	}
	return selected
}
func containsLine(data []byte, value string) bool {
	for _, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) == value {
			return true
		}
	}
	return false
}

var _ coreruntime.Manager = (*Manager)(nil)
