package commands

import (
	"errors"
	"testing"

	"github.com/peternicholls/stageserve/core/config"
	"github.com/peternicholls/stageserve/core/runtime"
	"github.com/peternicholls/stageserve/core/state"
)

func TestRecordedReadUsesUUIDNamesAndRefusesPathCollision(t *testing.T) {
	dir, stateDir := t.TempDir(), t.TempDir()
	store, err := state.NewStore(stateDir)
	if err != nil {
		t.Fatal(err)
	}
	id, err := store.CreateIdentity(dir, "demo")
	if err != nil {
		t.Fatal(err)
	}
	applied := config.ProjectConfig{Dir: dir, StateDir: stateDir, Slug: "demo", RuntimeBackend: runtime.BackendAppleContainer, ComposeProjectName: "stage-" + id.ProjectID}
	if err := store.Save(state.Record{Project: applied, ProjectID: id.ProjectID, InstallationID: id.InstallationID, AttachmentState: state.StateAttached}); err != nil {
		t.Fatal(err)
	}
	desired := applied
	desired.ComposeProjectName = "stage-demo"
	got, err := recordedProjectForRead(desired, "")
	if err != nil || got.ComposeProjectName != applied.ComposeProjectName {
		t.Fatalf("runtime config=%+v, err=%v", got, err)
	}
	desired.Dir = t.TempDir()
	if _, err := recordedProjectForRead(desired, ""); !errors.Is(err, state.ErrOwnerMismatch) {
		t.Fatalf("path collision=%v", err)
	}
	got, err = recordedProjectForRead(desired, "demo")
	if err != nil || got.ComposeProjectName != applied.ComposeProjectName {
		t.Fatalf("explicit selector=%+v, err=%v", got, err)
	}
}
