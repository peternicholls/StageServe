package commands

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/peternicholls/stageserve/core/config"
	"github.com/peternicholls/stageserve/core/onboarding"
	"github.com/peternicholls/stageserve/core/runtime"
	"github.com/peternicholls/stageserve/core/state"
)

func executeScopedJSON(t *testing.T, command string, projectDir, stackHome string, extra ...string) (onboarding.CommandResult, error) {
	t.Helper()
	root := NewRoot("test")
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(io.Discard)
	args := []string{command, "--json", "--project-dir", projectDir, "--stack-home", stackHome}
	root.SetArgs(append(args, extra...))
	err := root.Execute()
	raw := out.String()
	for _, forbidden := range []string{"\x1b", "private-scope-password", "MYSQL_PASSWORD", "MYSQL_ROOT_PASSWORD"} {
		if strings.Contains(raw, forbidden) {
			t.Fatalf("output contains %q: %s", forbidden, raw)
		}
	}
	decoder := json.NewDecoder(&out)
	var result onboarding.CommandResult
	if decodeErr := decoder.Decode(&result); decodeErr != nil {
		t.Fatalf("decode: %v (command error %v) output %s", decodeErr, err, raw)
	}
	var extraDoc any
	if decodeErr := decoder.Decode(&extraDoc); decodeErr != io.EOF {
		t.Fatalf("expected one JSON document: %v", decodeErr)
	}
	if result.SchemaVersion != 1 {
		t.Fatalf("schema_version=%d", result.SchemaVersion)
	}
	return result, err
}

func TestOnboardingJSONResolvedScope(t *testing.T) {
	t.Setenv("MYSQL_PASSWORD", "private-scope-password")
	t.Setenv("MYSQL_ROOT_PASSWORD", "private-scope-password")
	for _, command := range []string{"init", "doctor", "setup"} {
		t.Run(command, func(t *testing.T) {
			projectDir, stackHome := t.TempDir(), t.TempDir()
			stateDir := filepath.Join(t.TempDir(), "absent-state")
			t.Setenv("STAGESERVE_STATE_DIR", stateDir)
			result, _ := executeScopedJSON(t, command, projectDir, stackHome, "--dry-run")
			canonicalDir, err := filepath.EvalSymlinks(projectDir)
			if err != nil {
				t.Fatal(err)
			}
			if result.ProjectScope == nil || result.ProjectScope.Dir != canonicalDir || result.ProjectScope.Slug == "" || result.ProjectScope.ProjectID != "" {
				t.Fatalf("scope=%+v", result.ProjectScope)
			}
			if command == "init" {
				if _, err := os.Stat(filepath.Join(projectDir, ".env.stageserve")); !os.IsNotExist(err) {
					t.Fatalf("dry-run wrote settings: %v", err)
				}
			}
		})
	}
}

func TestOnboardingJSONConfigErrorsOmitScope(t *testing.T) {
	for _, command := range []string{"init", "doctor", "setup"} {
		t.Run(command, func(t *testing.T) {
			result, err := executeScopedJSON(t, command, filepath.Join(t.TempDir(), "missing"), t.TempDir())
			if err == nil || result.ProjectScope != nil || result.OverallStatus != onboarding.OverallError || result.ExitCode != onboarding.ExitError {
				t.Fatalf("result=%+v error=%v", result, err)
			}
		})
	}
}

func TestOnboardingJSONRegisteredAndDamagedScope(t *testing.T) {
	for _, scenario := range []string{"registered", "identity-only", "lost-ledger", "lost-installation", "corrupt-ledger", "corrupt-journal"} {
		t.Run(scenario, func(t *testing.T) {
			projectDir, stackHome, stateDir := t.TempDir(), t.TempDir(), t.TempDir()
			t.Setenv("STAGESERVE_STATE_DIR", stateDir)
			cfg, err := loadConfig(&SharedFlags{ProjectDir: projectDir, StackHome: stackHome})
			if err != nil {
				t.Fatal(err)
			}
			store, err := state.NewStore(cfg.StateDir)
			if err != nil {
				t.Fatal(err)
			}
			identity, err := store.CreateIdentity(cfg.Dir, cfg.Slug)
			if err != nil {
				t.Fatal(err)
			}
			if scenario != "identity-only" {
				rec := state.Record{ProjectID: identity.ProjectID, InstallationID: identity.InstallationID, Project: config.ProjectConfig{Dir: cfg.Dir, Slug: cfg.Slug, RuntimeBackend: runtime.BackendAppleContainer}, AttachmentState: state.StateDown}
				if err := store.Save(rec); err != nil {
					t.Fatal(err)
				}
			}
			switch scenario {
			case "corrupt-journal":
				op, err := store.BeginOperation(identity.ProjectID, identity.Revision, "up", "config-hash")
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(stateDir, "operations", op.ID+".json"), []byte(`{"private-scope-password":`), 0600); err != nil {
					t.Fatal(err)
				}
			case "lost-ledger":
				if err := os.Remove(filepath.Join(stateDir, "identities", identity.ProjectID+".json")); err != nil {
					t.Fatal(err)
				}
			case "lost-installation":
				if err := os.Remove(filepath.Join(stateDir, "installation.json")); err != nil {
					t.Fatal(err)
				}
			case "corrupt-ledger":
				if err := os.WriteFile(filepath.Join(stateDir, "identities", identity.ProjectID+".json"), []byte(`{"private-scope-password":`), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if scenario != "registered" && scenario != "identity-only" {
				scope, scopeErr := onboardingProjectScope(cfg)
				if scopeErr == nil || scope.ProjectID != "" {
					t.Fatalf("invalid ownership escaped helper: scope=%+v error=%v", scope, scopeErr)
				}
			}
			for _, command := range []string{"init", "doctor", "setup"} {
				result, commandErr := executeScopedJSON(t, command, projectDir, stackHome, "--dry-run")
				if scenario == "registered" || scenario == "identity-only" {
					if (command == "init" && commandErr != nil) || result.ProjectScope == nil || result.ProjectScope.ProjectID != identity.ProjectID {
						t.Fatalf("%s scope=%+v error=%v", command, result.ProjectScope, commandErr)
					}
				} else {
					hasScopeError := false
					for _, step := range result.Steps {
						if step.Code == "project-state-invalid" {
							hasScopeError = true
						}
					}
					if commandErr == nil || result.OverallStatus != onboarding.OverallError || result.ProjectScope == nil || result.ProjectScope.ProjectID != "" || !hasScopeError {
						t.Fatalf("%s result=%+v error=%v", command, result, commandErr)
					}
					if _, err := os.Stat(filepath.Join(projectDir, ".env.stageserve")); !os.IsNotExist(err) {
						t.Fatalf("damaged state wrote settings: %v", err)
					}
				}
			}
		})
	}
}

func TestOnboardingJSONInitWritesWithResolvedScope(t *testing.T) {
	projectDir, stackHome := t.TempDir(), t.TempDir()
	t.Setenv("STAGESERVE_STATE_DIR", filepath.Join(t.TempDir(), "absent"))
	result, err := executeScopedJSON(t, "init", projectDir, stackHome)
	if err != nil || result.ProjectScope == nil || result.OverallStatus != onboarding.OverallReady {
		t.Fatalf("result=%+v error=%v", result, err)
	}
	if _, err := os.Stat(filepath.Join(projectDir, ".env.stageserve")); err != nil {
		t.Fatal(err)
	}
}

func TestOnboardingProjectScopeAcceptsCanonicalAlias(t *testing.T) {
	projectDir, stateDir := t.TempDir(), t.TempDir()
	canonicalDir, err := filepath.EvalSymlinks(projectDir)
	if err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(t.TempDir(), "alias")
	if err := os.Symlink(canonicalDir, alias); err != nil {
		t.Fatal(err)
	}
	store, err := state.NewStore(stateDir)
	if err != nil {
		t.Fatal(err)
	}
	identity, err := store.CreateIdentity(canonicalDir, "demo")
	if err != nil {
		t.Fatal(err)
	}
	rec := state.Record{ProjectID: identity.ProjectID, InstallationID: identity.InstallationID, Project: config.ProjectConfig{Dir: alias, Slug: "demo", RuntimeBackend: runtime.BackendAppleContainer}, AttachmentState: state.StateDown}
	if err := store.Save(rec); err != nil {
		t.Fatal(err)
	}
	for _, dir := range []string{alias, canonicalDir} {
		scope, err := onboardingProjectScope(config.ProjectConfig{Dir: dir, Slug: "demo", StateDir: stateDir})
		if err != nil || scope.ProjectID != identity.ProjectID {
			t.Fatalf("scope=%+v error=%v", scope, err)
		}
	}
}

func TestOnboardingProjectScopeRejectsInstallationOnlyDamage(t *testing.T) {
	for _, raw := range []string{`{"schema_version":`, `{"schema_version":999,"id":"00000000-0000-4000-8000-000000000001"}`} {
		t.Run(raw, func(t *testing.T) {
			stateDir := t.TempDir()
			if err := os.WriteFile(filepath.Join(stateDir, "installation.json"), []byte(raw), 0600); err != nil {
				t.Fatal(err)
			}
			scope, err := onboardingProjectScope(config.ProjectConfig{Dir: t.TempDir(), Slug: "demo", StateDir: stateDir})
			if err == nil || scope.ProjectID != "" {
				t.Fatalf("scope=%+v error=%v", scope, err)
			}
		})
	}
}
