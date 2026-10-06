package lifecycle_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/peternicholls/stageserve/core/config"
	"github.com/peternicholls/stageserve/core/lifecycle"
	runtime "github.com/peternicholls/stageserve/core/runtime"
	"github.com/peternicholls/stageserve/core/state"
	"github.com/peternicholls/stageserve/internal/mocks"
	"github.com/peternicholls/stageserve/platform/ports"
)

type ownedLifecycleRuntime struct {
	runtime.Manager
	store           state.LifecycleStore
	resources       map[string]runtime.Resource
	mutations       int
	cancel          context.CancelFunc
	stopError       error
	cleanupCanceled bool
}

func ownedRuntime(store state.LifecycleStore) *ownedLifecycleRuntime {
	return &ownedLifecycleRuntime{Manager: mocks.NewRuntime(mocks.NewDocker(), mocks.NewComposer()), store: store, resources: map[string]runtime.Resource{}}
}
func (m *ownedLifecycleRuntime) Start(ctx context.Context, opts runtime.StartOptions) error {
	if err := opts.Ownership.Validate(); err != nil {
		return err
	}
	ops, err := m.store.PendingOperations(opts.Ownership.ProjectID)
	if err != nil {
		return err
	}
	found := false
	for _, op := range ops {
		if op.ID == opts.Ownership.OperationID {
			found = true
		}
	}
	if !found {
		return errors.New("runtime mutation preceded durable journal")
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	roles := []string{"gateway"}
	if opts.Ownership.Scope == "project" {
		roles = []string{"database", "web"}
	}
	for _, role := range roles {
		kind, id := "container", opts.ProjectName+"-"+role
		if role == "database" {
			kind = "volume"
			id = opts.ProjectName + "-db-data"
		}
		old, exists := m.resources[id]
		if exists && opts.ForceRecreate && kind == "container" {
			old.Deleted = true
			if err := opts.RecordResource(ctx, old); err != nil {
				return err
			}
			delete(m.resources, id)
			exists = false
		}
		if exists {
			if err := opts.RecordResource(ctx, old); err != nil {
				return err
			}
			continue
		}
		r := runtime.Resource{Kind: kind, ID: id, Name: id, InstallationID: opts.Ownership.InstallationID, ProjectID: opts.Ownership.ProjectID, Role: role, CreationOperation: opts.Ownership.OperationID, Planned: true}
		if err := opts.RecordResource(ctx, r); err != nil {
			return fmt.Errorf("record role %s planned %v incarnation %s: %w", role, r.Planned, r.CreationOperation, err)
		}
		// Persisted planned creation is required before the first external mutation.
		pending, err := m.store.PendingOperations(r.ProjectID)
		if err != nil {
			return err
		}
		planned := false
		for _, op := range pending {
			for _, v := range op.Resources {
				if v.ID == id && v.Planned {
					planned = true
				}
			}
		}
		if !planned {
			return errors.New("creation lacked durable resource intent")
		}
		m.mutations++
		r.Planned = false
		m.resources[id] = r
		if err := opts.RecordResource(ctx, r); err != nil {
			return fmt.Errorf("record role %s planned %v incarnation %s: %w", role, r.Planned, r.CreationOperation, err)
		}
	}
	if opts.Ownership.Scope == "project" && m.cancel != nil {
		m.cancel()
		return context.Canceled
	}
	return nil
}
func (m *ownedLifecycleRuntime) Stop(ctx context.Context, opts runtime.StopOptions) error {
	m.cleanupCanceled = ctx.Err() != nil
	if m.stopError != nil {
		return m.stopError
	}
	if err := opts.Ownership.Validate(); err != nil {
		return err
	}
	for _, record := range opts.Ownership.Resources {
		if record.Kind == "volume" && !opts.RemoveVolumes {
			continue
		}
		if record.Deleted {
			continue
		}
		if live, ok := m.resources[record.ID]; ok {
			if live.ProjectID != record.ProjectID || live.InstallationID != record.InstallationID {
				return state.ErrOwnerMismatch
			}
			delete(m.resources, record.ID)
			m.mutations++
			record = live
		}
		record.Planned = false
		record.Deleted = true
		if err := opts.RecordResource(ctx, record); err != nil {
			return err
		}
	}
	return nil
}
func (m *ownedLifecycleRuntime) WaitHealthy(ctx context.Context, _ string, _ time.Duration) error {
	return ctx.Err()
}
func (m *ownedLifecycleRuntime) ListServices(ctx context.Context, name string) ([]runtime.Service, error) {
	out := []runtime.Service{}
	for _, r := range m.resources {
		if r.Kind == "container" && !r.Deleted && r.Name == name+"-web" {
			out = append(out, runtime.Service{ID: r.ID, Name: r.Name, Service: "apache", Status: "running", Labels: map[string]string{"io.stageserve.installation-id": r.InstallationID, "io.stageserve.project-id": r.ProjectID}})
		}
	}
	return out, nil
}
func ownedOrchestrator(store state.LifecycleStore, r *ownedLifecycleRuntime) *lifecycle.Orchestrator {
	return lifecycle.New(lifecycle.Deps{Runtime: r, Gateway: mocks.NewGateway(), State: store, Ports: mocks.NewPorts(ports.Allocation{MySQLPort: 3306, PMAPort: 8081})})
}
func realOwnershipSetup(t *testing.T) (config.ProjectConfig, *state.Store, *ownedLifecycleRuntime) {
	t.Helper()
	cfg := newCfg(t)
	store, err := state.NewStore(cfg.StateDir)
	if err != nil {
		t.Fatal(err)
	}
	r := ownedRuntime(store)
	return cfg, store, r
}

func TestOwnedLifecycleDetachFreshAttachRetainsIdentityAndVolume(t *testing.T) {
	cfg, store, r := realOwnershipSetup(t)
	o := ownedOrchestrator(store, r)
	if err := o.Up(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}
	before, err := store.IdentityForPath(cfg.Dir)
	if err != nil {
		t.Fatal(err)
	}
	volumeID := "stage-" + before.ProjectID + "-db-data"
	volume := r.resources[volumeID]
	if err := o.Detach(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}
	fresh, err := state.NewStore(cfg.StateDir)
	if err != nil {
		t.Fatal(err)
	}
	retained, err := fresh.LoadIdentity(before.ProjectID)
	if err != nil || retained.Registered {
		t.Fatalf("detach: %+v %v", retained, err)
	}
	r.store = fresh
	o = ownedOrchestrator(fresh, r)
	if err := o.Attach(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}
	after, err := fresh.IdentityForPath(cfg.Dir)
	if err != nil || after.ProjectID != before.ProjectID || !after.Registered {
		t.Fatalf("reattach: %+v %v", after, err)
	}
	if r.resources[volumeID].CreationOperation != volume.CreationOperation {
		t.Fatal("reattach recreated retained database volume")
	}
	rec, err := fresh.Load(cfg.Slug)
	if err != nil || rec.ProjectID != before.ProjectID || rec.Project.DatabaseVolume != volumeID {
		t.Fatalf("projection: %+v %v", rec, err)
	}
}
func TestOwnedLifecycleCopiedPathCreatesIndependentIdentity(t *testing.T) {
	cfg, store, r := realOwnershipSetup(t)
	o := ownedOrchestrator(store, r)
	if err := o.Up(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}
	before, _ := store.IdentityForPath(cfg.Dir)
	copy := cfg
	copy.Dir = t.TempDir()
	copy.Slug = "copy"
	copy.Name = "copy"
	copy.Hostname = "copy.test"
	if err := o.Up(context.Background(), copy); err != nil {
		t.Fatal(err)
	}
	after, err := store.IdentityForPath(copy.Dir)
	if err != nil || after.ProjectID == before.ProjectID {
		t.Fatalf("copy adopted identity: %+v %v", after, err)
	}
	if _, ok := r.resources["stage-"+after.ProjectID+"-db-data"]; !ok {
		t.Fatal("copy lacked independent volume")
	}
}
func TestOwnedLifecycleLostLedgerMakesNoMutations(t *testing.T) {
	cfg, store, r := realOwnershipSetup(t)
	o := ownedOrchestrator(store, r)
	if err := o.Up(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}
	id, _ := store.IdentityForPath(cfg.Dir)
	before := r.mutations
	projection := filepath.Join(cfg.StateDir, "projects", cfg.Slug+".json")
	bytes, _ := os.ReadFile(projection)
	if err := os.Remove(filepath.Join(cfg.StateDir, "identities", id.ProjectID+".json")); err != nil {
		t.Fatal(err)
	}
	for _, action := range []func() error{func() error { return o.Up(context.Background(), cfg) }, func() error { return o.Down(context.Background(), cfg, false) }, func() error { return o.Detach(context.Background(), cfg) }} {
		if err := action(); err == nil {
			t.Fatal("lost ledger accepted")
		}
	}
	actual, _ := os.ReadFile(projection)
	if string(actual) != string(bytes) || r.mutations != before {
		t.Fatal("lost ledger changed project state or runtime")
	}
}
func TestOwnedLifecycleCancellationCleansWithFreshContext(t *testing.T) {
	cfg, store, r := realOwnershipSetup(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	r.cancel = cancel
	err := ownedOrchestrator(store, r).Up(ctx, cfg)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel: %v", err)
	}
	if r.cleanupCanceled {
		t.Fatal("cleanup inherited cancellation")
	}
	id, e := store.IdentityForPath(cfg.Dir)
	if e != nil {
		t.Fatal(e)
	}
	ops, e := store.PendingOperations(id.ProjectID)
	if e != nil || len(ops) != 0 {
		t.Fatalf("cleanup pending: %+v %v", ops, e)
	}
	for _, v := range id.Resources {
		if v.Kind == "container" && !v.Deleted {
			t.Fatal("created service remained live")
		}
	}
}
func TestOwnedLifecycleCleanupFailureRemainsDegraded(t *testing.T) {
	cfg, store, r := realOwnershipSetup(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	r.cancel = cancel
	r.stopError = errors.New("cleanup failed")
	err := ownedOrchestrator(store, r).Up(ctx, cfg)
	if err == nil {
		t.Fatal("cleanup failure accepted")
	}
	id, e := store.IdentityForPath(cfg.Dir)
	if e != nil {
		t.Fatal(e)
	}
	ops, e := store.PendingOperations(id.ProjectID)
	if e != nil || len(ops) != 1 || ops[0].Phase != "degraded" || len(ops[0].Leftovers) == 0 {
		t.Fatalf("missing degraded leftovers: %+v %v", ops, e)
	}
	before := r.mutations
	if err := ownedOrchestrator(store, r).Up(context.Background(), cfg); err == nil || r.mutations != before {
		t.Fatal("pending operation permitted another mutation")
	}
}

type failObservedStore struct {
	state.LifecycleStore
	failed bool
}

func (s *failObservedStore) SaveOperation(op state.Operation) error {
	if !s.failed && op.ProjectID != op.InstallationID {
		for _, r := range op.Resources {
			if r.Kind == "container" && !r.Planned && !r.Deleted {
				s.failed = true
				return errors.New("post-creation record failed")
			}
		}
	}
	return s.LifecycleStore.SaveOperation(op)
}
func TestOwnedLifecyclePostCreationRecordFailureCannotLoseResource(t *testing.T) {
	cfg, store, r := realOwnershipSetup(t)
	fault := &failObservedStore{LifecycleStore: store}
	r.store = fault
	err := ownedOrchestrator(fault, r).Up(context.Background(), cfg)
	if err == nil || !fault.failed {
		t.Fatalf("missing recording fault: %v", err)
	}
	id, e := store.IdentityForPath(cfg.Dir)
	if e != nil {
		t.Fatal(e)
	}
	ops, e := store.PendingOperations(id.ProjectID)
	if e != nil {
		t.Fatal(e)
	}
	for _, resource := range r.resources {
		if resource.ProjectID == id.ProjectID && resource.Kind == "container" {
			if len(ops) == 0 {
				t.Fatal("created unrecorded service hidden by terminal journal")
			}
			return
		}
	}
	found := false
	for _, resource := range id.Resources {
		if resource.Kind == "container" && resource.Deleted {
			found = true
		}
	}
	if !found {
		t.Fatal("cleanup lost resource tombstone")
	}
}
