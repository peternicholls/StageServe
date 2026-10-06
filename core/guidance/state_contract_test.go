package guidance

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/peternicholls/stageserve/core/config"
)

func TestCollectDoesNotAdoptIncompatibleRecordedState(t *testing.T) {
	for _, body := range []string{
		`{"project":{"Slug":"demo"},"attachment_state":"attached"}`,
		`{"schema_version":999,"project":{"Slug":"demo","RuntimeBackend":"apple-container"},"attachment_state":"attached"}`,
		`{"schema_version":2,"project":{"Slug":"demo","RuntimeBackend":"docker"},"attachment_state":"attached"}`,
	} {
		dir, stateDir := t.TempDir(), t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, ".env.stageserve"), nil, 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.Mkdir(filepath.Join(stateDir, "projects"), 0700); err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(stateDir, "projects", "demo.json")
		if err := os.WriteFile(path, []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
		got := Collect(context.Background(), config.ProjectConfig{Dir: dir, StateDir: stateDir, Slug: "demo"}, CollectOptions{})
		if got.ProjectState != nil || len(got.Warnings) == 0 || !strings.Contains(got.Warnings[0], "could not be read") {
			t.Fatalf("incompatible state adopted: %+v", got)
		}
		if Plan(got).Situation == SituationProjectRunning {
			t.Fatal("incompatible record reported as running")
		}
		data, err := os.ReadFile(path)
		if err != nil || string(data) != body {
			t.Fatalf("state changed: %q, %v", data, err)
		}
	}
}

func TestCollectMissingStateDoesNotCreateStateDirectory(t *testing.T) {
	dir := t.TempDir()
	stateDir := filepath.Join(t.TempDir(), "absent")
	Collect(context.Background(), config.ProjectConfig{Dir: dir, StateDir: stateDir, Slug: "demo"}, CollectOptions{})
	if _, err := os.Stat(stateDir); !os.IsNotExist(err) {
		t.Fatalf("context collection created state: %v", err)
	}
}
