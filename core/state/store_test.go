// State store round-trip tests. Each subtest uses an isolated state directory.
package state

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/peternicholls/stageserve/core/config"
	"github.com/peternicholls/stageserve/core/runtime"
)

func TestStore_SaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStore(dir)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	rec := Record{
		Project: config.ProjectConfig{
			Slug:               "demo",
			Name:               "demo",
			Dir:                "/tmp/demo",
			Hostname:           "demo.test",
			ComposeProjectName: "stage-demo",
			MySQL:              config.MySQL{Port: 33060, PMAPort: 8082},
		},
		AttachmentState: StateAttached,
	}
	if err := store.Save(rec); err != nil {
		t.Fatalf("save: %v", err)
	}
	got, err := store.Load("demo")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got.Project.Hostname != "demo.test" || got.AttachmentState != StateAttached {
		t.Errorf("round-trip mismatch: %+v", got)
	}
	if got.SchemaVersion != SchemaVersion {
		t.Errorf("schema version not stamped: %d", got.SchemaVersion)
	}
	if got.Project.RuntimeBackend != runtime.BackendAppleContainer || got.Runtime.Backend != runtime.BackendAppleContainer {
		t.Errorf("runtime backend defaults not stamped: project=%q observed=%q", got.Project.RuntimeBackend, got.Runtime.Backend)
	}
}

func TestStore_LoadRecordDefaultsToAppleContainer(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	legacy := `{"schema_version":1,"project":{"Slug":"legacy","Name":"legacy"},"attachment_state":"down","runtime":{}}`
	if err := os.WriteFile(filepath.Join(dir, "projects", "legacy.json"), []byte(legacy), 0o644); err != nil {
		t.Fatal(err)
	}
	rec, err := store.Load("legacy")
	if err != nil {
		t.Fatalf("load legacy record: %v", err)
	}
	if rec.Project.RuntimeBackend != runtime.BackendAppleContainer || rec.Runtime.Backend != runtime.BackendAppleContainer {
		t.Fatalf("backend project=%q observed=%q want apple-container", rec.Project.RuntimeBackend, rec.Runtime.Backend)
	}
}

func TestStore_LoadUnknownBackendFailsClosed(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	unknown := `{"schema_version":2,"project":{"Slug":"future","RuntimeBackend":"unknown"},"attachment_state":"down","runtime":{}}`
	if err := os.WriteFile(filepath.Join(dir, "projects", "future.json"), []byte(unknown), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Load("future"); err == nil {
		t.Fatal("unknown runtime backend loaded successfully")
	}
}

func TestStore_RegistryProjection(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStore(dir)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	for _, slug := range []string{"alpha", "beta"} {
		rec := Record{
			Project: config.ProjectConfig{
				Slug: slug, Name: slug, Hostname: slug + ".test",
				ComposeProjectName: "stage-" + slug,
			},
			AttachmentState: StateAttached,
		}
		if err := store.Save(rec); err != nil {
			t.Fatalf("save %s: %v", slug, err)
		}
	}
	rows, err := store.Registry()
	if err != nil {
		t.Fatalf("registry: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("registry rows=%d want 2", len(rows))
	}
	if rows[0].Slug != "alpha" || rows[1].Slug != "beta" {
		t.Errorf("registry not slug-sorted: %+v", rows)
	}
}

func TestStore_Remove(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewStore(dir)
	rec := Record{Project: config.ProjectConfig{Slug: "gone"}}
	_ = store.Save(rec)
	if err := store.Remove("gone"); err != nil {
		t.Fatalf("remove: %v", err)
	}
	if _, err := store.Load("gone"); err == nil {
		t.Errorf("load after remove should fail")
	}
	if err := store.Remove("never-existed"); err != nil {
		t.Errorf("remove of missing slug should be idempotent: %v", err)
	}
}
