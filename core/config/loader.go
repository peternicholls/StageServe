// Loader implements ConfigLoader. It resolves the configuration precedence
// chain as defined by the active StageServe config contract:
//
//  1. CLI flags (highest)
//  2. .env.stageserve in the project directory
//  3. shell environment
//  4. .env.stageserve in the stack home (canonical stack-owned defaults file)
//  5. application .env DB fallback
//  6. built-in defaults (lowest)
//
// STAGESERVE_POST_UP_COMMAND is a special-case bootstrap setting that is only
// honored when set in the project's .env.stageserve file. It is intentionally
// ignored if set via shell environment, stack-home .env.stageserve, or project .env.
package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/peternicholls/stageserve/core/project"
	"github.com/peternicholls/stageserve/core/runtime"
)

// Loader is the default ConfigLoader implementation.
type Loader struct {
	// Env supplies the shell environment. nil means os.LookupEnv.
	Env func(string) (string, bool)
	// StackHomeOverride forces the stack home (used by tests). When empty the
	// loader uses STACK_HOME or the directory holding the bundled stack catalog.
	StackHomeOverride string
}

// NewLoader returns a Loader with the live os environment.
func NewLoader() *Loader { return &Loader{Env: os.LookupEnv} }

// envOrDefault looks up key in the loader environment.
func (l *Loader) envOrDefault(key, fallback string) string {
	get := l.Env
	if get == nil {
		get = os.LookupEnv
	}
	if v, ok := get(key); ok && v != "" {
		return v
	}
	return fallback
}

// loadEnvFile reads a KEY=VALUE file (POSIX-shell quoting) into a map.
// Mirrors stageserve_load_env_file: blank lines and # comments skipped, optional
// leading "export ", surrounding " or ' stripped.
func loadEnvFile(path string) (map[string]string, error) {
	out := map[string]string{}
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return out, nil
		}
		return nil, err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		eq := strings.IndexByte(line, '=')
		if eq <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:eq])
		value := strings.TrimSpace(line[eq+1:])
		if !validEnvKey(key) {
			continue
		}
		switch {
		case len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"':
			unquoted, unquoteErr := strconv.Unquote(value)
			if unquoteErr != nil {
				value = value[1 : len(value)-1]
			} else {
				value = unquoted
			}
		case len(value) >= 2 && value[0] == '\'' && value[len(value)-1] == '\'':
			value = value[1 : len(value)-1]
		}
		out[key] = value
	}
	return out, sc.Err()
}

func validEnvKey(k string) bool {
	if k == "" {
		return false
	}
	for i, r := range k {
		switch {
		case r == '_':
		case r >= 'A' && r <= 'Z':
		case r >= 'a' && r <= 'z':
		case r >= '0' && r <= '9' && i > 0:
		default:
			return false
		}
	}
	return true
}

func defaultStateDir(stackHome string) string {
	return filepath.Join(stackHome, ".stageserve-state")
}

func loadProjectStageserveEnv(projectDir string) (map[string]string, error) {
	return loadEnvFile(filepath.Join(projectDir, ".env.stageserve"))
}

func loadProjectRuntimeEnv(projectDir string) (map[string]string, error) {
	return loadEnvFile(filepath.Join(projectDir, ".env"))
}

// loadStackEnv reads the canonical stack-owned defaults file. .env.stageserve is
// the only supported file; legacy <stackHome>/.stackenv and <stackHome>/.env
// are intentionally NOT consulted (workspace legacy policy: no compat shims).
func loadStackEnv(stackHome string) (map[string]string, error) {
	return loadEnvFile(filepath.Join(stackHome, ".env.stageserve"))
}

var osStat = os.Stat

// resolveStackHome reproduces stageserve_default_stack_home.
func (l *Loader) resolveStackHome() (string, error) {
	if l.StackHomeOverride != "" {
		return project.AbsDir(l.StackHomeOverride)
	}
	if v := l.envOrDefault("STACK_HOME", ""); v != "" {
		return project.AbsDir(v)
	}
	// Walk up from this binary's location to find the bundled stack catalog.
	exe, err := os.Executable()
	if err == nil {
		dir := filepath.Dir(exe)
		if stackCatalogRootExists(dir) {
			return dir, nil
		}
	}
	if cwd, err := os.Getwd(); err == nil {
		if stackCatalogRootExists(cwd) {
			return cwd, nil
		}
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".stageserve"), nil
}

// Load implements ConfigLoader.
func (l *Loader) Load(projectDir string, flags CLIFlags) (ProjectConfig, error) {
	cfg := ProjectConfig{}

	// 1. Project dir resolution (CLI wins, otherwise cwd).
	pd := flags.ProjectDir
	if pd == "" {
		pd = projectDir
	}
	if pd == "" {
		var err error
		pd, err = os.Getwd()
		if err != nil {
			return cfg, fmt.Errorf("resolve project dir: %w", err)
		}
	}
	pdAbs, err := project.AbsDir(pd)
	if err != nil {
		return cfg, fmt.Errorf("resolve project dir: %w", err)
	}
	cfg.Dir = pdAbs

	// 2. Stack home + state dir.
	stackHome, err := l.resolveStackHome()
	if err != nil {
		return cfg, fmt.Errorf("resolve stack home: %w", err)
	}
	cfg.StackHome = stackHome
	stateDir := l.envOrDefault("STAGESERVE_STATE_DIR", defaultStateDir(stackHome))
	cfg.StateDir = project.AbsPathFromBase(stackHome, stateDir)

	// 3. Build the precedence-merged map. Lower precedence first; higher
	// precedence overwrites by key. Order: defaults -> application DB fallback ->
	// stack .env.stageserve -> shell env -> project .env.stageserve ->
	// CLI flags.
	merged := defaults()
	origins := map[string]ConfigSource{}
	for _, key := range trackedEnvKeys {
		origins[key] = SourceDefault
	}
	origins["STAGESERVE_POST_UP_COMMAND"] = SourceDefault
	// Application DB values are the lowest-priority explicit source.
	if values, err := loadProjectRuntimeEnv(pdAbs); err != nil {
		return cfg, fmt.Errorf("read project .env: %w", err)
	} else {
		for appKey, key := range map[string]string{"DB_DATABASE": "MYSQL_DATABASE", "DB_USERNAME": "MYSQL_USER", "DB_PASSWORD": "MYSQL_PASSWORD"} {
			if value, present := values[appKey]; present {
				merged[key] = value
				origins[key] = SourceApplication
			}
		}
	}
	merge := func(values map[string]string, source ConfigSource) {
		for key, value := range values {
			if _, known := origins[key]; !known {
				continue
			}
			if key == "STACK_HOME" || key == "STAGESERVE_STATE_DIR" {
				continue
			}
			if key == "STAGESERVE_POST_UP_COMMAND" && source != SourceProject {
				continue
			}
			merged[key] = value
			origins[key] = source
		}
	}

	// .env.stageserve in the stack home applies just above built-in defaults.
	// STAGESERVE_POST_UP_COMMAND is excluded from this merge: bootstrap is a
	// project-scoped declaration sourced only from project .env.stageserve.
	if values, err := loadStackEnv(stackHome); err != nil {
		return cfg, fmt.Errorf("read stack .env.stageserve: %w", err)
	} else {
		merge(values, SourceStack)
	}
	for _, key := range trackedEnvKeys {
		if value, present := l.lookupEnv(key); present {
			merged[key] = value
			origins[key] = SourceShell
		}
	}
	if values, err := loadProjectStageserveEnv(pdAbs); err != nil {
		return cfg, fmt.Errorf("read project .env.stageserve: %w", err)
	} else {
		merge(values, SourceProject)
	}
	// CLI flags
	if flags.SiteName != "" {
		merged["SITE_NAME"] = flags.SiteName
		origins["SITE_NAME"] = SourceFlag
	}
	if flags.SiteHostname != "" {
		merged["SITE_HOSTNAME"] = flags.SiteHostname
		origins["SITE_HOSTNAME"] = SourceFlag
	}
	if flags.SiteSuffix != "" {
		merged["SITE_SUFFIX"] = flags.SiteSuffix
		origins["SITE_SUFFIX"] = SourceFlag
	}
	if flags.DocRoot != "" {
		merged["DOCROOT"] = flags.DocRoot
		origins["DOCROOT"] = SourceFlag
	}
	if flags.PHPVersion != "" {
		merged["PHP_VERSION"] = flags.PHPVersion
		origins["PHP_VERSION"] = SourceFlag
	}
	if flags.MySQLDatabase != "" {
		merged["MYSQL_DATABASE"] = flags.MySQLDatabase
		origins["MYSQL_DATABASE"] = SourceFlag
	}
	if flags.MySQLUser != "" {
		merged["MYSQL_USER"] = flags.MySQLUser
		origins["MYSQL_USER"] = SourceFlag
	}
	if flags.MySQLPort != "" {
		merged["MYSQL_PORT"] = flags.MySQLPort
		origins["MYSQL_PORT"] = SourceFlag
	}
	if flags.PMAPort != "" {
		merged["PMA_PORT"] = flags.PMAPort
		origins["PMA_PORT"] = SourceFlag
	}
	if flags.HostPort != "" {
		merged["HOST_PORT"] = flags.HostPort
		origins["HOST_PORT"] = SourceFlag
	}

	// 4. Materialise ProjectConfig from the merged map.
	stackKind := normalizeStackKind(merged["STAGESERVE_STACK"])
	if stackKind == "" {
		stackKind = "20i"
	}
	stackDef, ok := lookupStackDefinition(stackKind)
	if !ok {
		return cfg, fmt.Errorf("unsupported STAGESERVE_STACK %q: only 20i is implemented today", stackKind)
	}
	cfg.StackKind = stackDef.Kind
	if selected := strings.TrimSpace(merged["STAGESERVE_RUNTIME"]); selected != "" && selected != string(runtime.BackendAppleContainer) {
		return cfg, fmt.Errorf("unsupported STAGESERVE_RUNTIME: only apple-container is supported")
	}
	cfg.RuntimeBackend = runtime.BackendAppleContainer
	cfg.Origins = origins
	cfg.Stack = stackDef
	cfg.StackFile = stackDef.projectFilePath(stackHome)
	cfg.SharedFile = stackDef.sharedFilePath(stackHome)
	cfg.Name = strOr(merged["SITE_NAME"], filepath.Base(pdAbs))
	cfg.Slug = project.Slugify(cfg.Name)
	cfg.SiteSuffix = strOr(merged["SITE_SUFFIX"], "test")
	cfg.Hostname, cfg.SiteSuffix = project.ResolveHostname(cfg.Slug, merged["SITE_HOSTNAME"], cfg.SiteSuffix)
	if !project.HostnameValid(cfg.Hostname) {
		return cfg, fmt.Errorf("invalid site hostname %q: use a dotted hostname such as %s.%s", cfg.Hostname, cfg.Slug, cfg.SiteSuffix)
	}

	cfg.ComposeProjectName = strOr(merged["COMPOSE_PROJECT_NAME"], "stage-"+cfg.Slug)
	cfg.WebNetworkAlias = strOr(merged["WEB_NETWORK_ALIAS"], cfg.ComposeProjectName+"-nginx."+cfg.SiteSuffix)
	cfg.ContainerSiteRoot = "/home/sites/" + cfg.Slug
	cfg.RuntimeNetwork = cfg.ComposeProjectName + "-runtime"
	cfg.DatabaseVolume = cfg.ComposeProjectName + "-db-data"

	docroot, rel, err := project.ResolveDocRoot(pdAbs, merged["DOCROOT"], merged["CODE_DIR"])
	if err != nil {
		return cfg, err
	}
	cfg.DocRoot = docroot
	cfg.DocRootRelative = rel
	if rel == "" {
		cfg.ContainerDocRoot = cfg.ContainerSiteRoot
	} else {
		cfg.ContainerDocRoot = cfg.ContainerSiteRoot + "/" + rel
	}

	cfg.PHPVersion = strOr(merged["PHP_VERSION"], "8.5")

	// MySQL defaults key off the slug.
	cfg.MySQL.Version = strOr(merged["MYSQL_VERSION"], "10.6")
	cfg.MySQL.RootPassword = merged["MYSQL_ROOT_PASSWORD"]
	cfg.MySQL.Database = merged["MYSQL_DATABASE"]
	if origins["MYSQL_DATABASE"] == SourceDefault {
		cfg.MySQL.Database = cfg.Slug
	}
	cfg.MySQL.User = merged["MYSQL_USER"]
	if origins["MYSQL_USER"] == SourceDefault {
		cfg.MySQL.User = cfg.Slug
	}
	cfg.MySQL.Password = merged["MYSQL_PASSWORD"]

	// Shared gateway settings are runtime-owned, not env-configurable.
	cfg.SharedGateway.Network = "default"
	cfg.SharedGateway.HTTPPort = 80
	cfg.SharedGateway.HTTPSPort = 443
	cfg.SharedGateway.ComposeProjectName = "stage-shared"
	cfg.SharedGateway.ConfigFile = filepath.Join(cfg.StateDir, "shared", "gateway.conf")

	// Local DNS.
	cfg.LocalDNS.Provider = strOr(merged["LOCAL_DNS_PROVIDER"], "dnsmasq")
	cfg.LocalDNS.IP = strOr(merged["LOCAL_DNS_IP"], "127.0.0.1")
	cfg.LocalDNS.Port = atoiOr(merged["LOCAL_DNS_PORT"], 53535)
	cfg.LocalDNS.Suffix = strOr(merged["LOCAL_DNS_SUFFIX"], cfg.SiteSuffix)

	cfg.Ports.HostPort = atoiOr(merged["HOST_PORT"], 0)
	cfg.Ports.MySQLPort = atoiOr(merged["MYSQL_PORT"], 0)
	cfg.Ports.PMAPort = atoiOr(merged["PMA_PORT"], 0)
	cfg.MySQL.Port = cfg.Ports.MySQLPort
	cfg.MySQL.PMAPort = cfg.Ports.PMAPort

	// Wait timeout: CLI > env > default (FR-009).
	switch {
	case flags.WaitTimeoutSecs > 0:
		cfg.WaitTimeoutSecs = flags.WaitTimeoutSecs
		origins["STAGESERVE_WAIT_TIMEOUT"] = SourceFlag
	case merged["STAGESERVE_WAIT_TIMEOUT"] != "":
		cfg.WaitTimeoutSecs = atoiOr(merged["STAGESERVE_WAIT_TIMEOUT"], 120)
	default:
		cfg.WaitTimeoutSecs = 120
	}
	cfg.PostUpCommand = merged["STAGESERVE_POST_UP_COMMAND"]

	return cfg, nil
}

// trackedEnvKeys is the closed set of shell variables ConfigLoader honours.
// STAGESERVE_POST_UP_COMMAND is intentionally absent: bootstrap is sourced
// only from project .env.stageserve (FR-016).
var trackedEnvKeys = []string{
	"STAGESERVE_STACK", "STAGESERVE_RUNTIME",
	"SITE_NAME", "SITE_HOSTNAME", "SITE_SUFFIX", "DOCROOT", "CODE_DIR",
	"PHP_VERSION", "MYSQL_VERSION", "MYSQL_ROOT_PASSWORD",
	"MYSQL_DATABASE", "MYSQL_USER", "MYSQL_PASSWORD", "MYSQL_PORT", "PMA_PORT",
	"HOST_PORT", "COMPOSE_PROJECT_NAME", "WEB_NETWORK_ALIAS",
	"LOCAL_DNS_PROVIDER", "LOCAL_DNS_IP", "LOCAL_DNS_PORT", "LOCAL_DNS_SUFFIX",
	"STACK_HOME", "STAGESERVE_STATE_DIR", "STAGESERVE_WAIT_TIMEOUT",
}

func (l *Loader) lookupEnv(k string) (string, bool) {
	get := l.Env
	if get == nil {
		get = os.LookupEnv
	}
	return get(k)
}

func defaults() map[string]string {
	return map[string]string{
		"STAGESERVE_STACK":    "20i",
		"MYSQL_VERSION":       "10.6",
		"MYSQL_ROOT_PASSWORD": "root",
		"MYSQL_DATABASE":      "devdb",
		"MYSQL_USER":          "devuser",
		"MYSQL_PASSWORD":      "devpass",
		"PHP_VERSION":         "8.5",
		"LOCAL_DNS_PROVIDER":  "dnsmasq",
		"LOCAL_DNS_IP":        "127.0.0.1",
		"LOCAL_DNS_PORT":      "53535",
		"SITE_SUFFIX":         "test",
	}
}

func normalizeStackKind(v string) string {
	return strings.ToLower(strings.TrimSpace(v))
}

func strOr(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}

func atoiOr(v string, fallback int) int {
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
