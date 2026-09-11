// Orchestrator wires lifecycle steps over the lower-level interfaces. Up
// follows the documented 11-step flow with rollback at steps 6–9; Down /
// Attach / Detach / Status / Logs delegate to the same interfaces.
package lifecycle

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/peternicholls/stageserve/core/config"
	coreruntime "github.com/peternicholls/stageserve/core/runtime"
	"github.com/peternicholls/stageserve/core/state"
	"github.com/peternicholls/stageserve/infra/gateway"
	"github.com/peternicholls/stageserve/platform/ports"
	stls "github.com/peternicholls/stageserve/platform/tls"
)

// Deps bundles the collaborators the orchestrator needs.
type Deps struct {
	Runtime coreruntime.Manager
	Gateway gateway.GatewayManager
	State   state.StateStore
	Ports   ports.PortAllocator
	TLS     stls.Provider
}

// Orchestrator is the default implementation.
type Orchestrator struct {
	D Deps
}

const sharedGatewayHTTPSFallbackStart = 8443

var sharedGatewayListen = net.Listen

// New returns an orchestrator wired to deps.
func New(d Deps) *Orchestrator { return &Orchestrator{D: d} }

// Up runs the documented 11-step flow.
func (o *Orchestrator) Up(ctx context.Context, cfg config.ProjectConfig) error {
	cfg = resolveSharedGatewayPorts(cfg)
	if err := o.validateRuntimeCapabilities(cfg); err != nil {
		return err
	}
	if err := ValidateRuntimeAssets(cfg); err != nil {
		return err
	}
	if err := lifecycleContextErr(ctx, cfg); err != nil {
		return err
	}

	// Step 1: ensure shared network exists.
	if err := o.ensureSharedNetwork(ctx, cfg); err != nil {
		return Wrap("ensure-shared-network", cfg.Slug, err, "Verify Apple Container is running and its default network is available.")
	}
	// Step 2: allocate ports.
	registry, err := o.D.State.Registry()
	if err != nil {
		return Wrap("registry", cfg.Slug, err, "Inspect the state directory for unreadable JSON files.")
	}
	allocation, err := o.D.Ports.Allocate(ports.Request{
		HostPort:     cfg.Ports.HostPort,
		MySQLPort:    cfg.Ports.MySQLPort,
		PMAPort:      cfg.Ports.PMAPort,
		IsUp:         true,
		OwnSlug:      cfg.Slug,
		ProjectCount: countOtherActive(registry, cfg.Slug),
	}, registry)
	if err != nil {
		return Wrap("allocate-ports", cfg.Slug, fmt.Errorf("%w: %v", ErrPortConflict, err), "Free the conflicting port or pass --mysql-port / --pma-port.")
	}
	cfg.Ports.HostPort = allocation.HostPort
	cfg.Ports.MySQLPort = allocation.MySQLPort
	cfg.Ports.PMAPort = allocation.PMAPort
	cfg.MySQL.Port = allocation.MySQLPort
	cfg.MySQL.PMAPort = allocation.PMAPort

	registryRoutes := routesFromRegistry(registry)
	if err := o.ensureSharedGatewayTLS(cfg, routesWithProject(registryRoutes, cfg)); err != nil {
		return Wrap("tls-cert", cfg.Slug, err, "Install mkcert and trust the local CA with `mkcert -install`, then retry.")
	}

	if err := o.prepareSharedGatewayConfig(cfg, registryRoutes, ""); err != nil {
		return Wrap("gateway-config", "", err, "Inspect the gateway config path under the state directory.")
	}

	// Step 4: ensure shared gateway is running.
	if err := o.ensureSharedGateway(ctx, cfg); err != nil {
		return Wrap("shared-gateway", "", err, sharedRuntimeRemedy(" and inspect logs."))
	}

	// Step 5: write the per-project runtime environment (synthesized from cfg).
	envFile, err := writeEnvFile(cfg)
	if err != nil {
		return Wrap("write-env-file", cfg.Slug, err, "Check the project directory is writable.")
	}

	// Step 6: start the Apple Container project services.
	runtimeOpts := coreruntime.StartOptions{
		ProjectDir:  cfg.Dir,
		Definition:  cfg.StackFile,
		ProjectName: cfg.ComposeProjectName,
		EnvFile:     envFile,
		Profiles:    runtimeProfiles(cfg),
		WaitTimeout: time.Duration(cfg.WaitTimeoutSecs) * time.Second,
	}
	if err := o.D.Runtime.Start(ctx, runtimeOpts); err != nil {
		o.rollbackProject(ctx, cfg)
		return Wrap("runtime-up", cfg.Slug, err, "Check `stage logs` for the failing service.")
	}

	// Step 7: wait for healthchecks.
	if err := o.D.Runtime.WaitHealthy(ctx, cfg.ComposeProjectName, time.Duration(cfg.WaitTimeoutSecs)*time.Second); err != nil {
		o.rollbackProject(ctx, cfg)
		return Wrap("wait-healthy", cfg.Slug, err, "Inspect Apple Container state with `container list --all`, then run `stage logs`.")
	}
	if err := o.runPostUpHook(ctx, cfg); err != nil {
		o.rollbackProject(ctx, cfg)
		return Wrap("post-up-hook", cfg.Slug, err, "Check STAGESERVE_POST_UP_COMMAND and verify it succeeds inside the apache container.")
	}

	// Step 8: regenerate gateway config, reload gateway.
	registry, err = o.D.State.Registry()
	if err != nil {
		o.rollbackProject(ctx, cfg)
		return Wrap("registry", cfg.Slug, err, "Inspect the state directory for unreadable JSON files, then retry `stage up`.")
	}
	currentRoutes := routesFromRegistry(registry)
	nextRoutes := routesWithProject(currentRoutes, cfg)
	if err := o.ensureSharedGatewayTLS(cfg, nextRoutes); err != nil {
		o.rollbackProject(ctx, cfg)
		return Wrap("tls-cert", cfg.Slug, err, "Install mkcert and trust the local CA with `mkcert -install`, then retry.")
	}
	if err := o.prepareSharedGatewayConfig(cfg, nextRoutes, cfg.Slug); err != nil {
		o.rollbackProject(ctx, cfg)
		return Wrap("gateway-config", cfg.Slug, err, "Inspect the gateway config path under the state directory.")
	}
	if err := o.reloadSharedGateway(ctx, cfg); err != nil {
		o.rollbackProject(ctx, cfg)
		return Wrap("gateway-reload", cfg.Slug, err, sharedRuntimeRemedy("."))
	}

	// Step 9: persist state.
	rec := state.Record{
		SchemaVersion:   state.SchemaVersion,
		Project:         cfg,
		AttachmentState: state.StateAttached,
		Runtime:         observedRuntime(ctx, o.D.Runtime, cfg),
	}
	if err := o.D.State.Save(rec); err != nil {
		o.rollbackProject(ctx, cfg)
		return Wrap("save-state", cfg.Slug, err, "Inspect permissions on the state directory.")
	}

	// Step 10/11: log success (caller handles human output).
	return nil
}

func routesWithProject(current []gateway.Route, cfg config.ProjectConfig) []gateway.Route {
	route := gateway.Route{
		Hostname:        cfg.Hostname,
		Slug:            cfg.Slug,
		WebNetworkAlias: cfg.WebNetworkAlias,
	}
	merged := make([]gateway.Route, 0, len(current)+1)
	for _, existing := range current {
		if existing.Slug != route.Slug {
			merged = append(merged, existing)
		}
	}
	merged = append(merged, route)
	return merged
}

func runtimeProfiles(cfg config.ProjectConfig) []string {
	if strings.TrimSpace(cfg.Profile) == "" {
		return nil
	}
	return []string{cfg.Profile}
}

// ValidateRuntimeAssets verifies product-owned Apple Container definitions
// before any runtime side effect begins.
func ValidateRuntimeAssets(cfg config.ProjectConfig) error {
	for _, required := range []struct {
		step  string
		label string
		path  string
	}{
		{step: "runtime-asset-shared", label: "shared Apple Container definition", path: cfg.SharedFile},
		{step: "runtime-asset-project", label: "project Apple Container definition", path: cfg.StackFile},
	} {
		if err := validateRuntimeAsset(required.path); err != nil {
			return Wrap(required.step, cfg.Slug, err, runtimeAssetRemedy(cfg, required.label, required.path))
		}
	}
	return nil
}

func validateRuntimeAsset(path string) error {
	if strings.TrimSpace(path) == "" {
		return errors.New("required runtime asset path is empty")
	}
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("required runtime asset is missing at %s", path)
		}
		return fmt.Errorf("cannot access required runtime asset %s: %w", path, err)
	}
	if info.IsDir() {
		return fmt.Errorf("required runtime asset path is a directory: %s", path)
	}
	return nil
}

func runtimeAssetRemedy(cfg config.ProjectConfig, label, path string) string {
	return fmt.Sprintf("Reinstall StageServe or restore the bundled %s at %s, then run `stage doctor`.", label, path)
}

// Down stops the project, removes project-owned runtime state, and clears any active route.
func (o *Orchestrator) Down(ctx context.Context, cfg config.ProjectConfig, removeVolumes bool) error {
	cfg = resolveSharedGatewayPorts(cfg)
	if err := lifecycleContextErr(ctx, cfg); err != nil {
		return err
	}

	if err := o.stopProject(ctx, cfg, removeVolumes); err != nil {
		return Wrap("runtime-down", cfg.Slug, err, "Inspect the Apple Container error above.")
	}
	rec := state.Record{
		SchemaVersion:   state.SchemaVersion,
		Project:         cfg,
		AttachmentState: state.StateDown,
	}
	if err := o.D.State.Save(rec); err != nil {
		return Wrap("save-state", cfg.Slug, err, "Inspect permissions on the state directory.")
	}
	if err := o.syncSharedGateway(ctx, cfg, ""); err != nil {
		return Wrap("gateway-reload", cfg.Slug, err, sharedRuntimeRemedy("."))
	}
	if err := removeEnvFile(cfg); err != nil {
		return Wrap("remove-env-file", cfg.Slug, err, "Inspect permissions on the generated runtime env file under the state directory.")
	}
	return nil
}

// DownAll stops every recorded project runtime, removes all state records, and
// clears the shared gateway route set plus any generated per-project envfiles.
func (o *Orchestrator) DownAll(ctx context.Context, cfg config.ProjectConfig, removeVolumes bool) error {
	cfg = resolveSharedGatewayPorts(cfg)

	registry, err := o.D.State.Registry()
	if err != nil {
		return Wrap("registry", "", err, "Inspect the state directory for unreadable JSON files.")
	}
	projects := make([]config.ProjectConfig, 0, len(registry))
	var failures []string
	for _, row := range registry {
		rec, err := o.D.State.Load(row.Slug)
		if err != nil {
			failures = append(failures, fmt.Sprintf("%s load-state: %v", row.Slug, err))
			continue
		}
		projects = append(projects, rec.Project)
		if err := o.stopProject(ctx, rec.Project, removeVolumes); err != nil {
			failures = append(failures, fmt.Sprintf("%s runtime-down: %v", row.Slug, err))
			continue
		}
		rec.AttachmentState = state.StateDown
		rec.Runtime = state.RuntimeIdentity{}
		if err := o.D.State.Save(rec); err != nil {
			failures = append(failures, fmt.Sprintf("%s save-state: %v", row.Slug, err))
			continue
		}
	}
	if len(failures) > 0 {
		return Wrap("down-all", "", errors.New(strings.Join(failures, "; ")), "Some projects may already be stopped. Run `stage status --all`, inspect the listed project errors, then retry `stage down --all`.")
	}
	if err := o.syncSharedGateway(ctx, cfg, ""); err != nil {
		return Wrap("gateway-reload", "", err, sharedRuntimeRemedy("."))
	}
	for _, project := range projects {
		if err := removeEnvFile(project); err != nil {
			return Wrap("remove-env-file", project.Slug, err, "Inspect permissions on the generated runtime env file under the state directory.")
		}
	}
	return nil
}

// Attach updates state + gateway to mark the project routed.
func (o *Orchestrator) Attach(ctx context.Context, cfg config.ProjectConfig) error {
	cfg = resolveSharedGatewayPorts(cfg)
	if err := lifecycleContextErr(ctx, cfg); err != nil {
		return err
	}

	rec, err := o.D.State.Load(cfg.Slug)
	if err != nil {
		if errors.Is(err, state.ErrNotFound) {
			return o.Up(ctx, cfg)
		}
		return Wrap("attach", cfg.Slug, err, "Inspect the recorded state for this project.")
	}
	containers, err := o.D.Runtime.ListServices(ctx, cfg.ComposeProjectName)
	if err != nil {
		return Wrap("attach", cfg.Slug, err, "Inspect Apple Container service availability and the current project containers.")
	}
	if len(containers) == 0 {
		return o.Up(ctx, cfg)
	}
	rec.AttachmentState = state.StateAttached
	if err := o.D.State.Save(rec); err != nil {
		return Wrap("save-state", cfg.Slug, err, "Inspect permissions on the state directory, then retry `stage attach`.")
	}
	registry, err := o.D.State.Registry()
	if err != nil {
		return Wrap("registry", cfg.Slug, err, "Inspect the state directory for unreadable JSON files, then retry `stage attach`.")
	}
	currentRoutes := routesFromRegistry(registry)
	nextRoutes := routesWithProject(currentRoutes, cfg)
	if err := o.ensureSharedGatewayTLS(cfg, nextRoutes); err != nil {
		return Wrap("tls-cert", cfg.Slug, err, "Install mkcert and trust the local CA with `mkcert -install`, then retry.")
	}
	if err := o.prepareSharedGatewayConfig(cfg, nextRoutes, cfg.Slug); err != nil {
		return Wrap("gateway-config", cfg.Slug, err, "Inspect the gateway config path under the state directory.")
	}
	if err := o.reloadSharedGateway(ctx, cfg); err != nil {
		return Wrap("gateway-reload", cfg.Slug, err, sharedRuntimeRemedy("."))
	}
	return nil
}

// Detach stops the project, removes its runtime state, and clears its route.
func (o *Orchestrator) Detach(ctx context.Context, cfg config.ProjectConfig) error {
	cfg = resolveSharedGatewayPorts(cfg)
	if err := lifecycleContextErr(ctx, cfg); err != nil {
		return err
	}

	if err := o.stopProject(ctx, cfg, false); err != nil {
		return Wrap("runtime-down", cfg.Slug, err, "Inspect the Apple Container error above.")
	}
	if err := o.D.State.Remove(cfg.Slug); err != nil {
		return Wrap("remove-state", cfg.Slug, err, "Inspect permissions on the state directory.")
	}
	if err := o.syncSharedGateway(ctx, cfg, ""); err != nil {
		return Wrap("gateway-reload", cfg.Slug, err, sharedRuntimeRemedy("."))
	}
	if err := removeEnvFile(cfg); err != nil {
		return Wrap("remove-env-file", cfg.Slug, err, "Inspect permissions on the generated runtime env file under the state directory.")
	}
	return nil
}

// RestartService restarts one explicit project service without changing
// routing, state records, or project files.
func (o *Orchestrator) RestartService(ctx context.Context, cfg config.ProjectConfig, service string) error {
	service = strings.TrimSpace(service)
	if service == "" {
		return Wrap("restart-service", cfg.Slug, errors.New("service name is required"), "Choose a specific service before restarting it.")
	}
	if err := lifecycleContextErr(ctx, cfg); err != nil {
		return err
	}
	if err := ValidateRuntimeAssets(cfg); err != nil {
		return err
	}
	envFile := envFilePath(cfg)
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		envFile = ""
	} else if err != nil {
		return Wrap("restart-service", cfg.Slug, err, "Inspect permissions on the generated runtime env file under the state directory.")
	}
	if err := o.D.Runtime.Restart(ctx, coreruntime.RestartOptions{
		ProjectDir:  cfg.Dir,
		Definition:  cfg.StackFile,
		ProjectName: cfg.ComposeProjectName,
		EnvFile:     envFile,
		Service:     service,
	}); err != nil {
		return Wrap("restart-service", cfg.Slug, err, "Run `stage status`, then inspect the selected service logs before retrying.")
	}
	return nil
}

// --- helpers ---

func (o *Orchestrator) ensureSharedNetwork(ctx context.Context, cfg config.ProjectConfig) error {
	exists, err := o.D.Runtime.NetworkExists(ctx, cfg.SharedGateway.Network)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return o.D.Runtime.CreateNetwork(ctx, cfg.SharedGateway.Network)
}

func sharedRuntimeRemedy(suffix string) string {
	return fmt.Sprintf(
		"Run `stage doctor` to check Apple Container, or `stage up` to retry%s",
		suffix,
	)
}

func (o *Orchestrator) ensureSharedGateway(ctx context.Context, cfg config.ProjectConfig) error {
	return o.D.Runtime.Start(ctx, coreruntime.StartOptions{
		ProjectDir:  cfg.StackHome,
		Definition:  cfg.SharedFile,
		ProjectName: cfg.SharedGateway.ComposeProjectName,
		Env:         sharedGatewayEnv(cfg),
		WaitTimeout: time.Duration(cfg.WaitTimeoutSecs) * time.Second,
	})
}

func (o *Orchestrator) reloadSharedGateway(ctx context.Context, cfg config.ProjectConfig) error {
	return o.D.Runtime.Start(ctx, coreruntime.StartOptions{
		ProjectDir:    cfg.StackHome,
		Definition:    cfg.SharedFile,
		ProjectName:   cfg.SharedGateway.ComposeProjectName,
		Env:           sharedGatewayEnv(cfg),
		ForceRecreate: true,
		Services:      []string{"gateway"},
	})
}

func (o *Orchestrator) stopProject(ctx context.Context, cfg config.ProjectConfig, removeVolumes bool) error {
	envFile := envFilePath(cfg)
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		envFile = ""
	} else if err != nil {
		return err
	}
	return o.D.Runtime.Stop(ctx, coreruntime.StopOptions{
		ProjectDir:    cfg.Dir,
		Definition:    cfg.StackFile,
		ProjectName:   cfg.ComposeProjectName,
		EnvFile:       envFile,
		RemoveVolumes: removeVolumes,
	})
}

func lifecycleContextErr(ctx context.Context, cfg config.ProjectConfig) error {
	if err := ctx.Err(); err != nil {
		return Wrap("cancelled", cfg.Slug, err, "Run `stage status` to check this project before retrying.")
	}
	return nil
}

func (o *Orchestrator) syncSharedGateway(ctx context.Context, cfg config.ProjectConfig, preferredSlug string) error {
	registry, err := o.D.State.Registry()
	if err != nil {
		return err
	}
	if err := o.prepareSharedGatewayConfig(cfg, routesFromRegistry(registry), preferredSlug); err != nil {
		return err
	}
	return o.reloadSharedGateway(ctx, cfg)
}

func (o *Orchestrator) prepareSharedGatewayConfig(cfg config.ProjectConfig, routes []gateway.Route, preferredSlug string) error {
	_, _, err := o.D.Gateway.WriteConfig(gateway.RenderInput{
		Routes:        routes,
		PreferredSlug: preferredSlug,
		TLSEnabled:    cfg.SiteSuffix == "dev",
		HTTPSPort:     cfg.SharedGateway.HTTPSPort,
	})
	return err
}

func (o *Orchestrator) runPostUpHook(ctx context.Context, cfg config.ProjectConfig) error {
	if strings.TrimSpace(cfg.PostUpCommand) == "" {
		return nil
	}
	_, err := o.D.Runtime.Exec(ctx, coreruntime.ExecOptions{
		ProjectDir: cfg.Dir, Definition: cfg.StackFile, ProjectName: cfg.ComposeProjectName,
		Service: "apache", Command: []string{"sh", "-lc", cfg.PostUpCommand},
		WorkingDir: cfg.ContainerSiteRoot,
	})
	return err
}

func (o *Orchestrator) rollbackProject(ctx context.Context, cfg config.ProjectConfig) {
	envFile := envFilePath(cfg)
	_ = o.D.Runtime.Stop(ctx, coreruntime.StopOptions{
		ProjectDir:  cfg.Dir,
		Definition:  cfg.StackFile,
		ProjectName: cfg.ComposeProjectName,
		EnvFile:     envFile,
	})
	o.rollbackGatewayRoute(ctx, cfg)
	_ = removeEnvFile(cfg)
}

func (o *Orchestrator) rollbackGatewayRoute(ctx context.Context, cfg config.ProjectConfig) {
	registry, err := o.D.State.Registry()
	if err != nil {
		return
	}
	if err := o.prepareSharedGatewayConfig(cfg, routesFromRegistry(registry), ""); err != nil {
		return
	}
	_ = o.reloadSharedGateway(ctx, cfg)
}

func envFilePath(cfg config.ProjectConfig) string {
	return filepath.Join(cfg.StateDir, "envfiles", cfg.Slug+".env")
}

func removeEnvFile(cfg config.ProjectConfig) error {
	err := os.Remove(envFilePath(cfg))
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func writeEnvFile(cfg config.ProjectConfig) (string, error) {
	path := envFilePath(cfg)
	if err := mkdirAll(filepath.Dir(path)); err != nil {
		return "", err
	}
	body := strings.Join([]string{
		"COMPOSE_PROJECT_NAME=" + cfg.ComposeProjectName,
		"PROJECT_NAME=" + cfg.Name,
		"PROJECT_SLUG=" + cfg.Slug,
		"PROJECT_ROOT=" + cfg.Dir,
		"PROJECT_HOSTNAME=" + cfg.Hostname,
		"PROJECT_DOCROOT=" + cfg.DocRoot,
		"PROJECT_DIR=" + cfg.Dir,
		"DOCROOT=" + cfg.DocRoot,
		"CONTAINER_SITE_ROOT=" + cfg.ContainerSiteRoot,
		"CONTAINER_DOCROOT=" + cfg.ContainerDocRoot,
		"DB_HOST=mariadb",
		"APPLE_MARIADB_HOST=" + cfg.ComposeProjectName + "-mariadb." + cfg.SiteSuffix,
		"DB_PORT=3306",
		"DB_DATABASE=" + cfg.MySQL.Database,
		"DB_USERNAME=" + cfg.MySQL.User,
		"DB_PASSWORD=" + cfg.MySQL.Password,
		"PHP_VERSION=" + cfg.PHPVersion,
		"MYSQL_VERSION=" + cfg.MySQL.Version,
		"MYSQL_DATABASE=" + cfg.MySQL.Database,
		"MYSQL_USER=" + cfg.MySQL.User,
		"MYSQL_PASSWORD=" + cfg.MySQL.Password,
		"MYSQL_ROOT_PASSWORD=" + cfg.MySQL.RootPassword,
		"MYSQL_PORT=" + intStr(cfg.MySQL.Port),
		"PMA_PORT=" + intStr(cfg.MySQL.PMAPort),
		"WEB_NETWORK_ALIAS=" + cfg.WebNetworkAlias,
		"PROJECT_RUNTIME_NETWORK=" + cfg.RuntimeNetwork,
		"PROJECT_DATABASE_VOLUME=" + cfg.DatabaseVolume,
		"SHARED_GATEWAY_NETWORK=" + cfg.SharedGateway.Network,
		"PROJECT_NGINX_TEMPLATE=" + filepath.Join(cfg.StackHome, "docker", "nginx.conf.tmpl"),
	}, "\n") + "\n"
	if err := writeFile(path, []byte(body)); err != nil {
		return "", err
	}
	return path, nil
}

func sharedGatewayEnv(cfg config.ProjectConfig) []string {
	env := []string{
		"SHARED_GATEWAY_NETWORK=" + cfg.SharedGateway.Network,
		"SHARED_GATEWAY_HTTP_PORT=" + intStr(cfg.SharedGateway.HTTPPort),
		"SHARED_GATEWAY_HTTPS_PORT=" + intStr(cfg.SharedGateway.HTTPSPort),
		"SHARED_GATEWAY_CONFIG_FILE=" + cfg.SharedGateway.ConfigFile,
		"SHARED_GATEWAY_CERTS_DIR=/dev/null",
	}
	if sharedGatewayTLSEnabled(cfg) {
		env[len(env)-1] = "SHARED_GATEWAY_CERTS_DIR=" + sharedGatewayCertsDir(cfg)
	}
	return env
}

func (o *Orchestrator) ensureSharedGatewayTLS(cfg config.ProjectConfig, routes []gateway.Route) error {
	if !sharedGatewayTLSEnabled(cfg) {
		return nil
	}
	provider := o.D.TLS
	if provider == nil {
		provider = stls.NewMkcert()
	}
	certsDir := sharedGatewayCertsDir(cfg)
	_, err := provider.Ensure(
		filepath.Join(certsDir, "tls.pem"),
		filepath.Join(certsDir, "tls-key.pem"),
		gatewayTLSHosts(cfg, routes),
	)
	return err
}

func sharedGatewayTLSEnabled(cfg config.ProjectConfig) bool {
	return cfg.SiteSuffix == "dev"
}

func sharedGatewayCertsDir(cfg config.ProjectConfig) string {
	return filepath.Join(cfg.StateDir, "shared", "certs")
}

func gatewayTLSHosts(cfg config.ProjectConfig, routes []gateway.Route) []string {
	seen := map[string]bool{}
	if cfg.Hostname != "" {
		seen[cfg.Hostname] = true
	}
	for _, route := range routes {
		if route.Hostname != "" {
			seen[route.Hostname] = true
		}
	}
	hosts := make([]string, 0, len(seen))
	for host := range seen {
		hosts = append(hosts, host)
	}
	sort.Strings(hosts)
	return hosts
}

func resolveSharedGatewayPorts(cfg config.ProjectConfig) config.ProjectConfig {
	if cfg.SharedGateway.HTTPSPort == 0 {
		cfg.SharedGateway.HTTPSPort = 443
	}
	if cfg.SiteSuffix == "dev" && cfg.SharedGateway.HTTPSPort == 443 {
		cfg.SharedGateway.HTTPSPort = sharedGatewayHTTPSFallbackStart
		return cfg
	}
	if cfg.SharedGateway.HTTPSPort != 443 || !sharedGatewayPortInUse(443) {
		return cfg
	}
	if fallback, ok := firstAvailableSharedGatewayPort(sharedGatewayHTTPSFallbackStart); ok {
		cfg.SharedGateway.HTTPSPort = fallback
	}
	return cfg
}

func sharedGatewayPortInUse(port int) bool {
	ln, err := sharedGatewayListen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
	if err != nil {
		return true
	}
	_ = ln.Close()
	return false
}

func firstAvailableSharedGatewayPort(start int) (int, bool) {
	for port := start; port < 65535; port++ {
		if !sharedGatewayPortInUse(port) {
			return port, true
		}
	}
	return 0, false
}

func intStr(n int) string {
	if n == 0 {
		return ""
	}
	return fmt.Sprintf("%d", n)
}

func countOtherActive(rows []state.RegistryRow, ownSlug string) int {
	n := 0
	for _, r := range rows {
		if r.Slug == ownSlug {
			continue
		}
		if r.AttachmentState == state.StateAttached {
			n++
		}
	}
	return n
}

func routesFromRegistry(rows []state.RegistryRow) []gateway.Route {
	out := make([]gateway.Route, 0, len(rows))
	for _, r := range rows {
		if r.AttachmentState != state.StateAttached {
			continue
		}
		out = append(out, gateway.Route{
			Hostname:        r.Hostname,
			Slug:            r.Slug,
			WebNetworkAlias: r.WebNetworkAlias,
		})
	}
	return out
}

func observedRuntime(ctx context.Context, manager coreruntime.Manager, cfg config.ProjectConfig) state.RuntimeIdentity {
	containers, err := manager.ListServices(ctx, cfg.ComposeProjectName)
	if err != nil {
		return state.RuntimeIdentity{Backend: manager.Name()}
	}
	rt := state.RuntimeIdentity{Backend: manager.Name()}
	names := make([]string, 0, len(containers))
	for _, c := range containers {
		names = append(names, c.Service+"="+c.Status)
		ident := state.ContainerIdentity{ID: c.ID, Name: c.Name, Status: c.Status}
		switch c.Service {
		case "nginx":
			rt.Nginx = ident
		case "apache":
			rt.Apache = ident
		case "mariadb":
			rt.MariaDB = ident
		case "phpmyadmin":
			rt.PhpMyAdmin = ident
		}
	}
	rt.SummaryLine = strings.Join(names, " ")
	return rt
}

func (o *Orchestrator) validateRuntimeCapabilities(cfg config.ProjectConfig) error {
	if o.D.Runtime == nil {
		return Wrap("runtime-capability", cfg.Slug, errors.New("runtime manager is not configured"), "Reinstall StageServe or select a configured runtime backend.")
	}
	selected, err := coreruntime.ParseBackend(string(cfg.RuntimeBackend))
	if err != nil {
		return Wrap("runtime-capability", cfg.Slug, err, "Use the Apple Container runtime.")
	}
	if o.D.Runtime.Name() != selected {
		return Wrap("runtime-capability", cfg.Slug, fmt.Errorf("selected runtime %q is not available through configured backend %q", cfg.RuntimeBackend, o.D.Runtime.Name()), "Select the configured runtime or run `stage doctor` for backend readiness.")
	}
	capabilities := o.D.Runtime.Capabilities()
	if !capabilities.MultiService || !capabilities.SharedGateway {
		return Wrap("runtime-capability", cfg.Slug, fmt.Errorf("Apple Container does not support the required multi-service shared-gateway profile"), "Update Apple Container and run `stage doctor` before retrying.")
	}
	if strings.TrimSpace(cfg.Profile) != "" && !capabilities.DebugProfile {
		return Wrap("runtime-capability", cfg.Slug, fmt.Errorf("runtime %q does not support profile %q", cfg.RuntimeBackend, cfg.Profile), "Use STAGESERVE_RUNTIME=docker or remove the unsupported profile.")
	}
	return nil
}
