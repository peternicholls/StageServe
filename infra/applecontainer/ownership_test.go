package applecontainer

import (
	"context"
	"encoding/json"
	"errors"
	coreruntime "github.com/peternicholls/stageserve/core/runtime"
	"os"
	"path/filepath"
	"testing"
)

func testOwnership() coreruntime.Ownership {
	return coreruntime.Ownership{InstallationID: "11111111-1111-4111-8111-111111111111", ProjectID: "22222222-2222-4222-8222-222222222222", OperationID: "33333333-3333-4333-8333-333333333333", Scope: "project"}
}
func discardResource(context.Context, coreruntime.Resource) error { return nil }
func testDefinition(t *testing.T, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "runtime.json")
	if err := os.WriteFile(p, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	return p
}
func labelsFor(o coreruntime.Ownership, role string) map[string]string {
	return map[string]string{"io.stageserve.installation-id": o.InstallationID, "io.stageserve.project-id": o.ProjectID, "io.stageserve.operation-id": o.OperationID, "io.stageserve.scope": o.Scope, "io.stageserve.role": role, "io.stageserve.service": role, "io.stageserve.project": "demo"}
}

// Synthetic fixtures follow Apple's pinned 1.4.1 serialization, not live qualification.
func serviceFixture(name string, labels map[string]string) []byte {
	b, _ := json.Marshal([]any{map[string]any{"id": name, "configuration": map[string]any{"id": name, "labels": labels}, "status": map[string]any{"state": "running"}}})
	return b
}
func volumeFixture(name string, labels map[string]string) []byte {
	b, _ := json.Marshal([]any{map[string]any{"id": name, "configuration": map[string]any{"name": name, "labels": labels}}})
	return b
}
func TestOwnershipGateMakesNoCalls(t *testing.T) {
	for _, which := range []string{"start", "stop", "restart"} {
		for _, invalid := range []string{"owner", "recorder", "cancelled"} {
			t.Run(which+invalid, func(t *testing.T) {
				r := &fakeRunner{}
				m := NewManager(r)
				o := testOwnership()
				rec := coreruntime.ResourceRecorder(discardResource)
				ctx := context.Background()
				if invalid == "owner" {
					o.ProjectID = "bad"
				}
				if invalid == "recorder" {
					rec = nil
				}
				if invalid == "cancelled" {
					c, cancel := context.WithCancel(ctx)
					cancel()
					ctx = c
				}
				var err error
				switch which {
				case "start":
					err = m.Start(ctx, coreruntime.StartOptions{Ownership: o, RecordResource: rec})
				case "stop":
					err = m.Stop(ctx, coreruntime.StopOptions{Ownership: o, RecordResource: rec})
				case "restart":
					err = m.Restart(ctx, coreruntime.RestartOptions{Ownership: o, RecordResource: rec})
				}
				if err == nil || len(r.calls) != 0 {
					t.Fatalf("err=%v calls=%v", err, r.calls)
				}
			})
		}
	}
}
func TestForeignAndUnrecordedServiceCollisionPreserved(t *testing.T) {
	for _, foreign := range []bool{false, true} {
		o := testOwnership()
		labels := labelsFor(o, "web")
		if foreign {
			labels["io.stageserve.installation-id"] = "foreign"
			o.Resources = []coreruntime.Resource{newResource(o, "container", "demo-web", "web")}
		}
		r := &fakeRunner{outputs: map[string][]byte{"list --all --format json": serviceFixture("demo-web", labels)}}
		err := NewManager(r).Start(context.Background(), coreruntime.StartOptions{Definition: testDefinition(t, `{"services":[{"name":"web","image":"nginx"}]}`), ProjectName: "demo", Ownership: o, RecordResource: discardResource, ForceRecreate: true})
		if err == nil || len(r.calls) != 1 {
			t.Fatalf("err=%v calls=%v", err, r.calls)
		}
	}
}
func TestRecordFailureStopsBeforeNextMutation(t *testing.T) {
	r := &fakeRunner{outputs: map[string][]byte{"volume list --format json": []byte("[]")}}
	sentinel := errors.New("ledger write failed")
	err := NewManager(r).Start(context.Background(), coreruntime.StartOptions{Definition: testDefinition(t, `{"volumes":["custom-data"],"services":[{"name":"web","image":"nginx"}]}`), Ownership: testOwnership(), RecordResource: func(context.Context, coreruntime.Resource) error { return sentinel }})
	if !errors.Is(err, sentinel) || len(r.calls) != 1 {
		t.Fatalf("err=%v calls=%v", err, r.calls)
	}
}
func TestStopDeletesExactCustomVolumeOnly(t *testing.T) {
	for _, remove := range []bool{false, true} {
		o := testOwnership()
		resource := newResource(o, "volume", "custom-data", "volume")
		o.Resources = []coreruntime.Resource{resource}
		r := &fakeRunner{outputs: map[string][]byte{"list --all --format json": []byte("[]"), "volume list --format json": volumeFixture("custom-data", labelsFor(o, "volume"))}}
		var records []coreruntime.Resource
		err := NewManager(r).Stop(context.Background(), coreruntime.StopOptions{Definition: testDefinition(t, `{"volumes":["custom-data"],"services":[{"name":"web","image":"nginx"}]}`), Ownership: o, RemoveVolumes: remove, RecordResource: func(_ context.Context, r coreruntime.Resource) error { records = append(records, r); return nil }})
		if err != nil {
			t.Fatal(err)
		}
		if hasCall(r.calls, []string{"volume", "delete", "custom-data"}) != remove {
			t.Fatalf("calls=%v", r.calls)
		}
		if remove && (len(records) != 1 || !records[0].Deleted) {
			t.Fatalf("records=%v", records)
		}
	}
}
func TestTombstoneCannotAuthorizeReplacement(t *testing.T) {
	o := testOwnership()
	r := newResource(o, "container", "demo-web", "web")
	r.Deleted = true
	o.Resources = []coreruntime.Resource{r}
	if _, err := ownedObserved(o, "container", r.ID, r.Name, r.Role, labelsFor(o, "web")); err == nil {
		t.Fatal("tombstone authorized replacement")
	}
}
func TestNetworksFailClosed(t *testing.T) {
	r := &fakeRunner{}
	m := NewManager(r)
	if m.CreateNetwork(context.Background(), "demo") == nil || m.RemoveNetwork(context.Background(), "demo") == nil || len(r.calls) != 0 {
		t.Fatalf("calls=%v", r.calls)
	}
}

func TestCancellationAfterReuseRecordPreventsStart(t *testing.T) {
	o := testOwnership()
	resource := newResource(o, "container", "demo-web", "web")
	o.Resources = []coreruntime.Resource{resource}
	fixture := serviceFixture(resource.Name, labelsFor(o, "web"))
	fixture = []byte(string(fixture))
	var decoded []map[string]any
	if err := json.Unmarshal(fixture, &decoded); err != nil {
		t.Fatal(err)
	}
	decoded[0]["status"] = map[string]any{"state": "stopped"}
	fixture, _ = json.Marshal(decoded)
	r := &fakeRunner{outputs: map[string][]byte{"list --all --format json": fixture}}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := NewManager(r).Start(ctx, coreruntime.StartOptions{ProjectName: "demo", Definition: testDefinition(t, `{"services":[{"name":"web","image":"nginx"}]}`), Ownership: o, RecordResource: func(context.Context, coreruntime.Resource) error { cancel(); return nil }})
	if !errors.Is(err, context.Canceled) || len(r.calls) != 1 {
		t.Fatalf("err=%v calls=%v", err, r.calls)
	}
}

func TestForeignVolumePreserved(t *testing.T) {
	o := testOwnership()
	o.Resources = []coreruntime.Resource{newResource(o, "volume", "custom-data", "volume")}
	labels := labelsFor(o, "volume")
	labels["io.stageserve.project-id"] = "foreign"
	r := &fakeRunner{outputs: map[string][]byte{"volume list --format json": volumeFixture("custom-data", labels)}}
	err := NewManager(r).Start(context.Background(), coreruntime.StartOptions{Definition: testDefinition(t, `{"volumes":["custom-data"],"services":[{"name":"web","image":"nginx"}]}`), Ownership: o, RecordResource: discardResource})
	if err == nil || len(r.calls) != 1 {
		t.Fatalf("err=%v calls=%v", err, r.calls)
	}
}
func TestUndeclaredNamedVolumeFailsBeforeMutation(t *testing.T) {
	r := &fakeRunner{}
	err := NewManager(r).Start(context.Background(), coreruntime.StartOptions{Definition: testDefinition(t, `{"services":[{"name":"web","image":"nginx","mounts":["unknown:/data"]}]}`), Ownership: testOwnership(), RecordResource: discardResource})
	if err == nil || len(r.calls) != 0 {
		t.Fatalf("err=%v calls=%v", err, r.calls)
	}
}

func TestRecreateRecordsTombstoneBeforeNewIncarnation(t *testing.T) {
	owner := testOwnership()
	prior := owner
	prior.OperationID = "44444444-4444-4444-8444-444444444444"
	original := newResource(prior, "container", "demo-web", "web")
	owner.Resources = []coreruntime.Resource{original}
	runner := &fakeRunner{simulate: true, outputs: map[string][]byte{"list --all --format json": serviceFixture(original.Name, labelsFor(prior, "web"))}}
	var records []coreruntime.Resource
	err := NewManager(runner).Start(context.Background(), coreruntime.StartOptions{Definition: testDefinition(t, `{"services":[{"name":"web","image":"nginx"}]}`), ProjectName: "demo", Ownership: owner, ForceRecreate: true, RecordResource: func(_ context.Context, r coreruntime.Resource) error {
		records = append(records, r)
		if len(records) == 1 && !hasCall(runner.calls, []string{"delete", "--force", "demo-web"}) {
			t.Fatal("tombstone before deletion")
		}
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 3 || !records[0].Deleted || records[0].CreationOperation != prior.OperationID || !records[1].Planned || records[2].Deleted || records[2].Planned || records[2].CreationOperation != owner.OperationID {
		t.Fatalf("records=%v", records)
	}
}
func TestReusePreservesCreationOperation(t *testing.T) {
	owner := testOwnership()
	prior := owner
	prior.OperationID = "44444444-4444-4444-8444-444444444444"
	resource := newResource(prior, "volume", "custom-data", "volume")
	owner.Resources = []coreruntime.Resource{resource}
	runner := &fakeRunner{simulate: true, outputs: map[string][]byte{"volume list --format json": volumeFixture(resource.Name, labelsFor(prior, "volume")), "list --all --format json": []byte("[]")}}
	var records []coreruntime.Resource
	err := NewManager(runner).Start(context.Background(), coreruntime.StartOptions{Definition: testDefinition(t, `{"volumes":["custom-data"],"services":[{"name":"web","image":"nginx"}]}`), ProjectName: "demo", Ownership: owner, RecordResource: func(_ context.Context, r coreruntime.Resource) error { records = append(records, r); return nil }})
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 3 || records[0].CreationOperation != prior.OperationID {
		t.Fatalf("records=%v", records)
	}
}

func TestCreationRequiresObservedOwnedIncarnation(t *testing.T) {
	for _, kind := range []string{"container", "volume"} {
		for _, mode := range []string{"absent", "foreign", "write-failure", "success"} {
			t.Run(kind+mode, func(t *testing.T) {
				o := testOwnership()
				name := "demo-web"
				role := "web"
				key := "list --all --format json"
				body := `{"services":[{"name":"web","image":"nginx"}]}`
				labels := labelsFor(o, role)
				var after []byte
				if kind == "volume" {
					name = "custom-data"
					role = "volume"
					key = "volume list --format json"
					body = `{"volumes":["custom-data"],"services":[{"name":"web","image":"nginx"}]}`
					labels = labelsFor(o, role)
				}
				if mode == "foreign" {
					labels["io.stageserve.installation-id"] = "foreign"
				}
				if kind == "volume" {
					after = volumeFixture(name, labels)
				} else {
					after = serviceFixture(name, labels)
				}
				if mode == "absent" {
					after = []byte("[]")
				}
				r := &fakeRunner{outputs: map[string][]byte{}, sequence: map[string][][]byte{key: {[]byte("[]"), after}}}
				var records []coreruntime.Resource
				sentinel := errors.New("observed write failed")
				selected := []string(nil)
				if kind == "volume" {
					selected = []string{"none"}
				}
				err := NewManager(r).Start(context.Background(), coreruntime.StartOptions{ProjectName: "demo", Definition: testDefinition(t, body), Ownership: o, Services: selected, RecordResource: func(_ context.Context, v coreruntime.Resource) error {
					records = append(records, v)
					if mode == "write-failure" && !v.Planned {
						return sentinel
					}
					return nil
				}})
				if mode == "success" {
					if err != nil || len(records) != 2 || records[1].Planned {
						t.Fatalf("err=%v records=%v", err, records)
					}
				} else {
					if err == nil {
						t.Fatal("unverified creation accepted")
					}
					if mode == "write-failure" && !errors.Is(err, sentinel) {
						t.Fatal(err)
					}
				}
				if len(records) == 0 || !records[0].Planned {
					t.Fatalf("missing write ahead plan: %v", records)
				}
				mutations := 0
				for _, call := range r.calls {
					if call[0] == "run" || (len(call) > 1 && call[0] == "volume" && call[1] == "create") {
						mutations++
					}
				}
				if mutations != 1 {
					t.Fatalf("calls=%v", r.calls)
				}
			})
		}
	}
}
func TestMissingRecordedIncarnationPreventsCreation(t *testing.T) {
	for _, kind := range []string{"container", "volume"} {
		owner := testOwnership()
		name := "demo-web"
		role := "web"
		body := `{"services":[{"name":"web","image":"nginx"}]}`
		if kind == "volume" {
			name = "custom-data"
			role = "volume"
			body = `{"volumes":["custom-data"],"services":[{"name":"web","image":"nginx"}]}`
		}
		owner.Resources = []coreruntime.Resource{newResource(owner, kind, name, role)}
		runner := &fakeRunner{outputs: map[string][]byte{"volume list --format json": []byte("[]"), "list --all --format json": []byte("[]")}}
		err := NewManager(runner).Start(context.Background(), coreruntime.StartOptions{ProjectName: "demo", Definition: testDefinition(t, body), Ownership: owner, RecordResource: discardResource})
		if err == nil || len(runner.calls) != 1 {
			t.Fatalf("err=%v calls=%v", err, runner.calls)
		}
	}
}
func TestInvalidLedgerBlocksAllCalls(t *testing.T) {
	owner := testOwnership()
	owner.Resources = []coreruntime.Resource{newResource(owner, "container", "demo-web", "web")}
	owner.Resources[0].CreationOperation = "invalid"
	runner := &fakeRunner{}
	err := NewManager(runner).Start(context.Background(), coreruntime.StartOptions{Ownership: owner, RecordResource: discardResource})
	if err == nil || len(runner.calls) != 0 {
		t.Fatalf("err=%v calls=%v", err, runner.calls)
	}
}
