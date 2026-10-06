package state

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
)

func TestIdentityRetainedAndCopy(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStore(dir)
	p := t.TempDir()
	v, e := s.CreateIdentity(p, "one")
	if e != nil {
		t.Fatal(e)
	}
	alias := filepath.Join(t.TempDir(), "alias")
	if e = os.Symlink(p, alias); e != nil {
		t.Fatal(e)
	}
	same, e := s.CreateIdentity(alias, "one")
	if e != nil || same.ProjectID != v.ProjectID {
		t.Fatalf("alias: %+v %v", same, e)
	}
	v.Registered = false
	if e = s.SaveIdentity(v, v.Revision); e != nil {
		t.Fatal(e)
	}
	s, _ = NewStore(dir)
	got, e := s.IdentityForPath(p)
	if e != nil || got.ProjectID != v.ProjectID || got.Registered {
		t.Fatalf("retained: %+v %v", got, e)
	}
	if e = s.SaveIdentity(v, v.Revision); !errors.Is(e, ErrRevision) {
		t.Fatalf("stale: %v", e)
	}
	copy, e := s.CreateIdentity(t.TempDir(), "two")
	if e != nil || copy.ProjectID == v.ProjectID || len(copy.Resources) != 0 {
		t.Fatalf("copy: %+v %v", copy, e)
	}
	if _, e = s.CreateIdentity(t.TempDir(), "one"); !errors.Is(e, ErrCollision) {
		t.Fatalf("slug collision: %v", e)
	}
}
func TestJournalReloadAndOwner(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStore(dir)
	v, e := s.CreateIdentity(t.TempDir(), "demo")
	if e != nil {
		t.Fatal(e)
	}
	op, e := s.BeginOperation(v.ProjectID, v.Revision, "start", "hash")
	if e != nil {
		t.Fatal(e)
	}
	r := ResourceRecord{Kind: "volume", ID: "exact-id", Name: "volume", InstallationID: v.InstallationID, ProjectID: v.ProjectID, Role: "db", CreationOperation: op.ID}
	op.Resources = []ResourceRecord{r}
	op.Phase = "apply"
	if e = s.SaveOperation(op); e != nil {
		t.Fatal(e)
	}
	s, _ = NewStore(dir)
	ops, e := s.PendingOperations(v.ProjectID)
	if e != nil || len(ops) != 1 || ops[0].Resources[0].ID != "exact-id" {
		t.Fatalf("reload: %+v %v", ops, e)
	}
	op.Resources[0].InstallationID = "alien"
	if e = s.SaveOperation(op); !errors.Is(e, ErrOwnerMismatch) {
		t.Fatalf("owner: %v", e)
	}
	v.Resources = []ResourceRecord{r}
	if e = s.SaveIdentity(v, v.Revision); e != nil {
		t.Fatal(e)
	}
	got, e := s.LoadIdentity(v.ProjectID)
	if e != nil || len(got.Resources) != 1 {
		t.Fatalf("ledger: %+v %v", got, e)
	}
	if _, e = s.BeginOperation(v.ProjectID, got.Revision, "stop", ""); !errors.Is(e, ErrCollision) {
		t.Fatalf("pending collision: %v", e)
	}
	op.Resources[0] = r
	op.Phase = "committed"
	if e = s.SaveOperation(op); e != nil {
		t.Fatal(e)
	}
	ops, e = s.PendingOperations(v.ProjectID)
	if e != nil || len(ops) != 0 {
		t.Fatalf("completed: %+v %v", ops, e)
	}
}
func TestCorruptAndLostIdentityFailClosed(t *testing.T) {
	for _, which := range []string{"identity", "installation", "journal"} {
		t.Run(which, func(t *testing.T) {
			dir := t.TempDir()
			s, _ := NewStore(dir)
			v, _ := s.CreateIdentity(t.TempDir(), "demo")
			op, _ := s.BeginOperation(v.ProjectID, v.Revision, "start", "")
			path := filepath.Join(dir, "identities", v.ProjectID+".json")
			if which == "installation" {
				path = filepath.Join(dir, "installation.json")
			}
			if which == "journal" {
				path = filepath.Join(dir, "operations", op.ID+".json")
			}
			if e := os.WriteFile(path, []byte("{"), 0600); e != nil {
				t.Fatal(e)
			}
			if _, e := s.PendingOperations(v.ProjectID); e == nil {
				t.Fatal("corrupt state accepted")
			}
		})
	}
	dir := t.TempDir()
	s, _ := NewStore(dir)
	v, _ := s.CreateIdentity(t.TempDir(), "demo")
	os.Remove(filepath.Join(dir, "installation.json"))
	if _, e := s.LoadIdentity(v.ProjectID); e == nil {
		t.Fatal("missing installation adopted")
	}
	if _, e := s.CreateIdentity(t.TempDir(), "new"); e == nil {
		t.Fatal("new installation adopted old ledgers")
	}
}
func TestSchemaAndRuntimeFailClosed(t *testing.T) {
	for _, data := range []string{`{"project":{"Slug":"bad"}}`, `{"schema_version":1,"project":{"Slug":"bad"}}`, `{"schema_version":999,"project":{"Slug":"bad","RuntimeBackend":"apple-container"}}`, `{"schema_version":2,"project":{"Slug":"bad","RuntimeBackend":"docker"}}`, `{"schema_version":2,"project":{"Slug":"bad","RuntimeBackend":"apple-container"},"runtime":{"Backend":"unknown"}}`} {
		s, _ := NewStore(t.TempDir())
		path := s.projectFile("bad")
		os.WriteFile(path, []byte(data), 0600)
		if _, e := s.Load("bad"); e == nil {
			t.Fatal("invalid record accepted")
		}
		if _, e := s.Registry(); e == nil {
			t.Fatal("registry hid invalid state")
		}
		if e := s.Remove("bad"); e == nil {
			t.Fatal("invalid state removed")
		}
		got, _ := os.ReadFile(path)
		if string(got) != data {
			t.Fatal("invalid state changed")
		}
	}
}
func TestStoreRejectsTraversal(t *testing.T) {
	s, _ := NewStore(t.TempDir())
	for _, slug := range []string{"../escape", "/absolute", "..", ""} {
		if _, e := s.Load(slug); e == nil {
			t.Fatal("traversal load")
		}
		if e := s.Remove(slug); e == nil {
			t.Fatal("traversal remove")
		}
	}
}

func TestIdentityConcurrentProcesses(t *testing.T) {
	if os.Getenv("STAGESERVE_IDENTITY_CHILD") == "1" {
		s, err := NewStore(os.Getenv("STAGESERVE_IDENTITY_STORE"))
		if err != nil {
			t.Fatal(err)
		}
		v, err := s.CreateIdentity(os.Getenv("STAGESERVE_IDENTITY_PROJECT"), "shared")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(os.Getenv("STAGESERVE_IDENTITY_RESULT"), []byte(v.InstallationID+":"+v.ProjectID), 0600); err != nil {
			t.Fatal(err)
		}
		return
	}
	dir, project := t.TempDir(), t.TempDir()
	var commands []*exec.Cmd
	var results []string
	var outputs []*bytes.Buffer
	for i := 0; i < 6; i++ {
		result := filepath.Join(t.TempDir(), "result")
		cmd := exec.Command(os.Args[0], "-test.run=^TestIdentityConcurrentProcesses$")
		cmd.Env = append(os.Environ(), "STAGESERVE_IDENTITY_CHILD=1", "STAGESERVE_IDENTITY_STORE="+dir, "STAGESERVE_IDENTITY_PROJECT="+project, "STAGESERVE_IDENTITY_RESULT="+result)
		output := new(bytes.Buffer)
		cmd.Stdout, cmd.Stderr = output, output
		if err := cmd.Start(); err != nil {
			t.Fatal(err)
		}
		commands, results, outputs = append(commands, cmd), append(results, result), append(outputs, output)
	}
	var want string
	for i, cmd := range commands {
		if err := cmd.Wait(); err != nil {
			t.Fatalf("process: %v %s", err, outputs[i])
		}
		got, err := os.ReadFile(results[i])
		if err != nil {
			t.Fatal(err)
		}
		if i == 0 {
			want = string(got)
		} else if string(got) != want {
			t.Fatalf("split identity: %s != %s", got, want)
		}
	}
	s, _ := NewStore(dir)
	identities, err := s.identities()
	if err != nil || len(identities) != 1 {
		t.Fatalf("identities: %+v %v", identities, err)
	}
}

func TestIdentityConcurrentRevisionAndJournal(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStore(dir)
	v, err := s.CreateIdentity(t.TempDir(), "shared")
	if err != nil {
		t.Fatal(err)
	}
	run := func(fn func(*Store) error) {
		t.Helper()
		results := make(chan error, 12)
		var wg sync.WaitGroup
		for i := 0; i < cap(results); i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				fresh, e := NewStore(dir)
				if e == nil {
					e = fn(fresh)
				}
				results <- e
			}()
		}
		wg.Wait()
		close(results)
		successes := 0
		for e := range results {
			if e == nil {
				successes++
			} else if !errors.Is(e, ErrRevision) && !errors.Is(e, ErrCollision) {
				t.Fatal(e)
			}
		}
		if successes != 1 {
			t.Fatalf("successful writers: %d", successes)
		}
	}
	run(func(fresh *Store) error { return fresh.SaveIdentity(v, v.Revision) })
	v, err = s.LoadIdentity(v.ProjectID)
	if err != nil || v.Revision != 2 {
		t.Fatalf("revision: %+v %v", v, err)
	}
	run(func(fresh *Store) error {
		_, e := fresh.BeginOperation(v.ProjectID, v.Revision, "start", "hash")
		return e
	})
}

func TestJournalRejectsPersistedTampering(t *testing.T) {
	for _, field := range []string{"phase", "kind", "zero_revision", "future_revision"} {
		t.Run(field, func(t *testing.T) {
			dir := t.TempDir()
			s, _ := NewStore(dir)
			v, _ := s.CreateIdentity(t.TempDir(), "demo")
			op, _ := s.BeginOperation(v.ProjectID, v.Revision, "start", "hash")
			switch field {
			case "phase":
				op.Phase = "unknown"
			case "kind":
				op.Kind = ""
			case "zero_revision":
				op.BaseRevision = 0
			case "future_revision":
				op.BaseRevision = v.Revision + 1
			}
			if err := durableJSON(filepath.Join(dir, "operations", op.ID+".json"), op); err != nil {
				t.Fatal(err)
			}
			fresh, _ := NewStore(dir)
			if _, err := fresh.PendingOperations(v.ProjectID); !errors.Is(err, ErrInvalidIdentity) {
				t.Fatalf("tampering accepted: %v", err)
			}
		})
	}
}

func TestJournalMonotonicAndImmutable(t *testing.T) {
	s, _ := NewStore(t.TempDir())
	v, _ := s.CreateIdentity(t.TempDir(), "demo")
	op, _ := s.BeginOperation(v.ProjectID, v.Revision, "start", "hash")
	op.Phase = "apply"
	op.PriorConfiguration = json.RawMessage(`{"before":true}`)
	op.Resources = []ResourceRecord{{Kind: "volume", ID: "id", Name: "name", InstallationID: v.InstallationID, ProjectID: v.ProjectID, Role: "db", CreationOperation: op.ID}}
	if err := s.SaveOperation(op); err != nil {
		t.Fatal(err)
	}
	changed := op
	changed.Phase = "prepare"
	if err := s.SaveOperation(changed); !errors.Is(err, ErrRevision) {
		t.Fatalf("backward phase: %v", err)
	}
	changed = op
	changed.PriorConfiguration = json.RawMessage(`{"before":false}`)
	if err := s.SaveOperation(changed); !errors.Is(err, ErrOwnerMismatch) {
		t.Fatalf("prior configuration: %v", err)
	}
	changed = op
	changed.Resources = append([]ResourceRecord(nil), op.Resources...)
	changed.Resources[0].CreationOperation, _ = newUUID()
	if err := s.SaveOperation(changed); !errors.Is(err, ErrOwnerMismatch) {
		t.Fatalf("resource reassigned: %v", err)
	}
	for _, phase := range []string{"rollback", "degraded", "rollback", "rolled_back"} {
		op.Phase = phase
		if err := s.SaveOperation(op); err != nil {
			t.Fatalf("%s: %v", phase, err)
		}
	}
	if err := s.SaveOperation(op); !errors.Is(err, ErrRevision) {
		t.Fatalf("terminal rewrite: %v", err)
	}
}

func TestIdentityRetainsResourceProvenance(t *testing.T) {
	s, _ := NewStore(t.TempDir())
	v, _ := s.CreateIdentity(t.TempDir(), "demo")
	op, _ := s.BeginOperation(v.ProjectID, v.Revision, "start", "hash")
	v.Resources = []ResourceRecord{{Kind: "volume", ID: "id", Name: "name", InstallationID: v.InstallationID, ProjectID: v.ProjectID, Role: "db", CreationOperation: op.ID}}
	if err := s.SaveIdentity(v, v.Revision); err != nil {
		t.Fatal(err)
	}
	v, _ = s.LoadIdentity(v.ProjectID)
	removed := v
	removed.Resources = nil
	if err := s.SaveIdentity(removed, removed.Revision); !errors.Is(err, ErrOwnerMismatch) {
		t.Fatalf("forgot resource: %v", err)
	}
	v.Resources[0].Deleted = true
	if err := s.SaveIdentity(v, v.Revision); err != nil {
		t.Fatal(err)
	}
	v, _ = s.LoadIdentity(v.ProjectID)
	v.Resources[0].Deleted = false
	if err := s.SaveIdentity(v, v.Revision); !errors.Is(err, ErrOwnerMismatch) {
		t.Fatalf("resurrected resource: %v", err)
	}
}

func TestJournalCannotCompleteWithUnresolvedLeftovers(t *testing.T) {
	for _, terminal := range []string{"committed", "rolled_back"} {
		t.Run(terminal, func(t *testing.T) {
			dir := t.TempDir()
			s, _ := NewStore(dir)
			v, _ := s.CreateIdentity(t.TempDir(), "demo")
			op, _ := s.BeginOperation(v.ProjectID, v.Revision, "start", "hash")
			op.Leftovers = []ResourceRecord{{Kind: "volume", ID: "leftover", Name: "name", InstallationID: v.InstallationID, ProjectID: v.ProjectID, Role: "db", CreationOperation: op.ID}}
			if terminal == "rolled_back" {
				op.Phase = "degraded"
			} else {
				op.Phase = "route"
			}
			if err := s.SaveOperation(op); err != nil {
				t.Fatal(err)
			}
			op.Phase = terminal
			if err := s.SaveOperation(op); !errors.Is(err, ErrInvalidIdentity) {
				t.Fatalf("unresolved completion: %v", err)
			}
			fresh, _ := NewStore(dir)
			if _, err := fresh.BeginOperation(v.ProjectID, v.Revision, "stop", ""); !errors.Is(err, ErrCollision) {
				t.Fatalf("unresolved journal allowed mutation: %v", err)
			}
			// Direct persisted tampering must fail closed even in a fresh process.
			if err := durableJSON(filepath.Join(dir, "operations", op.ID+".json"), op); err != nil {
				t.Fatal(err)
			}
			if _, err := fresh.PendingOperations(v.ProjectID); !errors.Is(err, ErrInvalidIdentity) {
				t.Fatalf("persisted unresolved completion: %v", err)
			}
			op.Leftovers[0].Deleted = true
			// Restore the durable preceding phase to complete through the public API.
			prior := op
			if terminal == "rolled_back" {
				prior.Phase = "degraded"
			} else {
				prior.Phase = "route"
			}
			if err := durableJSON(filepath.Join(dir, "operations", op.ID+".json"), prior); err != nil {
				t.Fatal(err)
			}
			if err := fresh.SaveOperation(op); err != nil {
				t.Fatalf("resolved completion: %v", err)
			}
			if pending, err := fresh.PendingOperations(v.ProjectID); err != nil || len(pending) != 0 {
				t.Fatalf("resolved pending: %+v %v", pending, err)
			}
		})
	}
}

func TestIdentityLookupValidatesEmptyInstallation(t *testing.T) {
	for _, scenario := range []string{"corrupt", "newer-schema", "invalid-uuid", "orphan-operation", "orphan-identity"} {
		t.Run(scenario, func(t *testing.T) {
			dir := t.TempDir()
			s, err := NewStore(dir)
			if err != nil {
				t.Fatal(err)
			}
			expected := ErrOwnerMismatch
			switch scenario {
			case "corrupt":
				if err := os.WriteFile(filepath.Join(dir, "installation.json"), []byte("broken"), 0600); err != nil {
					t.Fatal(err)
				}
				expected = nil
			case "newer-schema":
				if err := durableJSON(filepath.Join(dir, "installation.json"), map[string]any{"schema_version": 2, "id": "00000000-0000-0000-0000-000000000000"}); err != nil {
					t.Fatal(err)
				}
				expected = ErrSchema
			case "invalid-uuid":
				if err := durableJSON(filepath.Join(dir, "installation.json"), map[string]any{"schema_version": 1, "id": "invalid"}); err != nil {
					t.Fatal(err)
				}
				expected = ErrInvalidIdentity
			case "orphan-operation", "orphan-identity":
				subdir := "operations"
				if scenario == "orphan-identity" {
					subdir = "identities"
				}
				if err := os.MkdirAll(filepath.Join(dir, subdir), 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, subdir, "orphan.json"), []byte("{}"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			path := t.TempDir()
			for name, call := range map[string]func() error{"lookup": func() error { _, err := s.IdentityForPath(path); return err }, "create": func() error { _, err := s.CreateIdentity(path, "demo"); return err }} {
				err := call()
				if err == nil || errors.Is(err, ErrNotFound) || (expected != nil && !errors.Is(err, expected)) {
					t.Fatalf("%s %s: %v", name, scenario, err)
				}
			}
		})
	}
}

func TestResourceIncarnationRetention(t *testing.T) {
	for _, scenario := range []string{"recreate", "two-live", "rewrite-provenance", "resurrection", "drop-tombstone"} {
		t.Run(scenario, func(t *testing.T) {
			s, err := NewStore(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			id, err := s.CreateIdentity(t.TempDir(), "demo")
			if err != nil {
				t.Fatal(err)
			}
			creation, err := newUUID()
			if err != nil {
				t.Fatal(err)
			}
			old := ResourceRecord{Kind: "volume", ID: "named", Name: "named", InstallationID: id.InstallationID, ProjectID: id.ProjectID, Role: "database", CreationOperation: creation, Deleted: scenario != "two-live" && scenario != "rewrite-provenance"}
			id.Resources = []ResourceRecord{old}
			if err := s.SaveIdentity(id, id.Revision); err != nil {
				t.Fatal(err)
			}
			id, err = s.LoadIdentity(id.ProjectID)
			if err != nil {
				t.Fatal(err)
			}
			next := old
			next.CreationOperation, err = newUUID()
			if err != nil {
				t.Fatal(err)
			}
			next.Deleted = false
			switch scenario {
			case "rewrite-provenance":
				id.Resources = []ResourceRecord{next}
			case "resurrection":
				id.Resources[0].Deleted = false
			case "drop-tombstone":
				id.Resources = []ResourceRecord{next}
			default:
				id.Resources = append(id.Resources, next)
			}
			err = s.SaveIdentity(id, id.Revision)
			if scenario == "recreate" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil {
				t.Fatal("invalid incarnation accepted")
			}
		})
	}
}

func TestPlannedResourcesRequireExplicitResolution(t *testing.T) {
	for _, resolution := range []string{"observed", "deleted"} {
		t.Run(resolution, func(t *testing.T) {
			s, err := NewStore(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			id, err := s.CreateIdentity(t.TempDir(), "demo")
			if err != nil {
				t.Fatal(err)
			}
			op, err := s.BeginOperation(id.ProjectID, id.Revision, "up", "")
			if err != nil {
				t.Fatal(err)
			}
			resource := ResourceRecord{Kind: "volume", ID: "intended", Name: "intended", InstallationID: id.InstallationID, ProjectID: id.ProjectID, Role: "database", CreationOperation: op.ID, Planned: true}
			op.Resources = []ResourceRecord{resource}
			op.Phase = "apply"
			if err := s.SaveOperation(op); err != nil {
				t.Fatal(err)
			}
			fresh, err := NewStore(s.stateDir)
			if err != nil {
				t.Fatal(err)
			}
			pending, err := fresh.PendingOperations(id.ProjectID)
			if err != nil || len(pending) != 1 || !pending[0].Resources[0].Planned {
				t.Fatalf("pending intention: %+v %v", pending, err)
			}
			if _, err := fresh.BeginOperation(id.ProjectID, id.Revision, "up", ""); !errors.Is(err, ErrCollision) {
				t.Fatalf("overlapping intention: %v", err)
			}
			if err := fresh.RecoverOperation(id.ProjectID); !errors.Is(err, ErrCollision) {
				t.Fatalf("planned runtime recovery: %v", err)
			}
			for _, phase := range []string{"committed", "rolled_back"} {
				terminal := op
				terminal.Phase = phase
				if err := fresh.SaveOperation(terminal); !errors.Is(err, ErrInvalidIdentity) {
					t.Fatalf("terminal %s planned: %v", phase, err)
				}
			}
			lost := op
			lost.Resources = nil
			if err := fresh.SaveOperation(lost); !errors.Is(err, ErrOwnerMismatch) {
				t.Fatalf("forgotten plan: %v", err)
			}
			id.Resources = []ResourceRecord{resource}
			if err := fresh.SaveIdentity(id, id.Revision); !errors.Is(err, ErrInvalidIdentity) {
				t.Fatalf("plan in independent ledger: %v", err)
			}
			resolved := op
			resolved.Resources = append([]ResourceRecord{}, op.Resources...)
			resolved.Resources[0].Planned = false
			if resolution == "deleted" {
				resolved.Resources[0].Deleted = true
				resolved.Phase = "rollback"
			}
			if err := fresh.SaveOperation(resolved); err != nil {
				t.Fatal(err)
			}
			resurrected := resolved
			resurrected.Resources = append([]ResourceRecord{}, resolved.Resources...)
			resurrected.Resources[0].Planned = true
			if err := fresh.SaveOperation(resurrected); err == nil {
				t.Fatal("observed resource became planned")
			}
			if resolution == "deleted" {
				resolved.Phase = "rolled_back"
			} else {
				resolved.Phase = "committed"
			}
			if err := fresh.SaveOperation(resolved); err != nil {
				t.Fatal(err)
			}
		})
	}
}
