package commands

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/peternicholls/stageserve/core/onboarding"
)

// Freeze script consent semantics before the guided editor evolves.
func TestInitScriptConsentContract(t *testing.T) {
	for _, tc := range []struct {
		name                    string
		existing, force, dryRun bool
		message                 string
	}{
		{"create", false, false, false, "created"},
		{"preserve", true, false, false, "already exists"},
		{"overwrite", true, true, false, "overwritten"},
		{"preview create", false, false, true, "would be created"},
		{"preview overwrite", true, true, true, "would be overwritten"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, ".env.stageserve")
			const original = "SITE_NAME=original\n"
			if tc.existing {
				if err := os.WriteFile(path, []byte(original), 0600); err != nil {
					t.Fatal(err)
				}
			}
			root := NewRoot("test")
			var stdout, stderr bytes.Buffer
			root.SetOut(&stdout)
			root.SetErr(&stderr)
			args := []string{"init", "--project-dir", dir, "--json", "--cli"}
			if tc.force {
				args = append(args, "--force")
			}
			if tc.dryRun {
				args = append(args, "--dry-run")
			}
			root.SetArgs(args)
			if err := root.Execute(); err != nil {
				t.Fatal(err)
			}
			var result onboarding.CommandResult
			if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
				t.Fatalf("single JSON result: %v: %s", err, stdout.String())
			}
			if result.ExitCode != 0 || result.OverallStatus != "ready" || len(result.Steps) != 1 || result.Steps[0].ID != "init.env_file" || result.Steps[0].Status != "ready" || !strings.Contains(result.Steps[0].Message, tc.message) {
				t.Fatalf("result contract: %+v", result)
			}
			assertMachineOutputHasNoTUIHints(t, stdout.String())
			body, err := os.ReadFile(path)
			switch {
			case !tc.existing && tc.dryRun:
				if !os.IsNotExist(err) {
					t.Fatalf("preview created settings: %v", err)
				}
			case tc.existing && (!tc.force || tc.dryRun):
				if err != nil || string(body) != original {
					t.Fatalf("existing settings changed: %q, %v", body, err)
				}
			default:
				if err != nil || !bytes.Contains(body, []byte("STAGESERVE_STACK=20i")) || string(body) == original {
					t.Fatalf("settings not written: %q, %v", body, err)
				}
			}
		})
	}
}

func TestInitInvalidWebFolderNeverWrites(t *testing.T) {
	for _, dryRun := range []bool{false, true} {
		dir := t.TempDir()
		path := filepath.Join(dir, ".env.stageserve")
		const original = "SITE_NAME=original\n"
		if err := os.WriteFile(path, []byte(original), 0600); err != nil {
			t.Fatal(err)
		}
		root := NewRoot("test")
		var out bytes.Buffer
		root.SetOut(&out)
		root.SetErr(&out)
		args := []string{"init", "--project-dir", dir, "--json", "--force", "--docroot", "../outside"}
		if dryRun {
			args = append(args, "--dry-run")
		}
		root.SetArgs(args)
		if err := root.Execute(); err == nil {
			t.Fatal("invalid web folder accepted")
		}
		body, err := os.ReadFile(path)
		if err != nil || string(body) != original {
			t.Fatalf("invalid settings changed file: %q, %v", body, err)
		}
	}
}
