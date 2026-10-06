package state

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func preparedCommit(t *testing.T) (*Store, Operation, Identity, Record) {
	t.Helper()
	s, id, rec := identifiedRecord(t)
	op, err := s.BeginOperation(id.ProjectID, id.Revision, "up", "desired")
	if err != nil {
		t.Fatal(err)
	}
	op.Phase = "route"
	if err := s.SaveOperation(op); err != nil {
		t.Fatal(err)
	}
	id.DesiredHash = "desired"
	id.AppliedHash = "desired"
	id.LastWriter = op.ID
	rec.Project.Name = "updated"
	return s, op, id, rec
}
func TestLifecycleCommitCrashRecovery(t *testing.T) {
	for _, boundary := range []string{"journal", "identity", "record", "committed"} {
		for _, unregister := range []bool{false, true} {
			t.Run(boundary+map[bool]string{true: "-unregister", false: "-update"}[unregister], func(t *testing.T) {
				s, op, id, rec := preparedCommit(t)
				resource := ResourceRecord{Kind: "volume", ID: "retained", Name: "database", InstallationID: id.InstallationID, ProjectID: id.ProjectID, Role: "database", CreationOperation: op.ID}
				prior, err := s.LoadIdentity(id.ProjectID)
				if err != nil {
					t.Fatal(err)
				}
				prior.Resources = []ResourceRecord{resource}
				if err := s.SaveIdentity(prior, prior.Revision); err != nil {
					t.Fatal(err)
				}
				id.Revision++
				id.Resources = []ResourceRecord{resource}
				var target *Record = &rec
				if unregister {
					id.Registered = false
					target = nil
				}
				fault := errors.New("simulated crash")
				s.transactionFault = func(name string) error {
					if name == boundary {
						return fault
					}
					return nil
				}
				if err := s.CommitOperation(op, id, target, id.Revision); !errors.Is(err, fault) {
					t.Fatalf("fault boundary: %v", err)
				}
				fresh, err := NewStore(s.stateDir)
				if err != nil {
					t.Fatal(err)
				}
				if err := fresh.RecoverOperation(id.ProjectID); err != nil {
					t.Fatal(err)
				}
				if err := fresh.RecoverOperation(id.ProjectID); err != nil {
					t.Fatal(err)
				}
				got, err := fresh.LoadIdentity(id.ProjectID)
				if err != nil {
					t.Fatal(err)
				}
				expected := id
				expected.Revision++
				if !reflect.DeepEqual(got, expected) {
					t.Fatalf("identity: %+v", got)
				}
				if unregister {
					if _, err := fresh.Load("demo"); !errors.Is(err, ErrNotFound) {
						t.Fatalf("record retained: %v", err)
					}
					if len(got.Resources) != 1 || got.Resources[0].Deleted {
						t.Fatal("retained volume lost")
					}
				} else {
					loaded, err := fresh.Load("demo")
					if err != nil || loaded.Project.Name != "updated" {
						t.Fatalf("projection: %+v %v", loaded, err)
					}
				}
				ops, err := fresh.PendingOperations(id.ProjectID)
				if err != nil || len(ops) != 0 {
					t.Fatalf("pending: %+v %v", ops, err)
				}
				if err := fresh.CommitOperation(op, id, target, id.Revision); err != nil {
					t.Fatalf("idempotent commit: %v", err)
				}
			})
		}
	}
}
func TestLifecycleCommitConflictFailsClosed(t *testing.T) {
	for _, scenario := range []string{"identity", "record", "payload", "save-identity", "save-operation"} {
		t.Run(scenario, func(t *testing.T) {
			s, op, id, rec := preparedCommit(t)
			s.transactionFault = func(name string) error {
				if name == "journal" {
					return errors.New("crash")
				}
				return nil
			}
			if err := s.CommitOperation(op, id, &rec, id.Revision); err == nil {
				t.Fatal("missing injected fault")
			}
			switch scenario {
			case "identity":
				current, err := s.LoadIdentity(id.ProjectID)
				if err != nil {
					t.Fatal(err)
				}
				current.EngineVersion = "external"
				if err := durableJSON(filepath.Join(s.stateDir, "identities", id.ProjectID+".json"), current); err != nil {
					t.Fatal(err)
				}
			case "record":
				if err := os.WriteFile(s.projectFile("demo"), []byte("external"), 0600); err != nil {
					t.Fatal(err)
				}
			case "payload":
				var journal Operation
				if err := readJSON(filepath.Join(s.stateDir, "operations", op.ID+".json"), &journal); err != nil {
					t.Fatal(err)
				}
				journal.Commit.TargetIdentity.InstallationID = "00000000-0000-0000-0000-000000000000"
				if err := durableJSON(filepath.Join(s.stateDir, "operations", op.ID+".json"), journal); err != nil {
					t.Fatal(err)
				}
			case "save-identity":
				if err := s.SaveIdentity(id, id.Revision); !errors.Is(err, ErrRevision) {
					t.Fatalf("SaveIdentity: %v", err)
				}
				return
			case "save-operation":
				var journal Operation
				if err := readJSON(filepath.Join(s.stateDir, "operations", op.ID+".json"), &journal); err != nil {
					t.Fatal(err)
				}
				journal.Commit.Record.Project.Name = "tampered"
				if err := s.SaveOperation(journal); !errors.Is(err, ErrOwnerMismatch) {
					t.Fatalf("SaveOperation: %v", err)
				}
				return
			}
			fresh, err := NewStore(s.stateDir)
			if err != nil {
				t.Fatal(err)
			}
			if err := fresh.RecoverOperation(id.ProjectID); err == nil {
				t.Fatal("recovery accepted conflict")
			}
		})
	}
}
func TestLifecycleCommitRejectsRevisionAndOwnership(t *testing.T) {
	for _, scenario := range []string{"revision", "record", "unregister", "resource-loss"} {
		t.Run(scenario, func(t *testing.T) {
			s, op, id, rec := preparedCommit(t)
			revision := id.Revision
			switch scenario {
			case "revision":
				revision++
			case "record":
				rec.ProjectID = "00000000-0000-0000-0000-000000000000"
			case "unregister":
				id.Registered = false
			case "resource-loss":
				old := id
				old.Resources = []ResourceRecord{{Kind: "volume", ID: "retained", Name: "db", InstallationID: id.InstallationID, ProjectID: id.ProjectID, Role: "db", CreationOperation: op.ID}}
				if err := s.SaveIdentity(old, old.Revision); err != nil {
					t.Fatal(err)
				}
				id.Revision++
				revision++
			}
			if err := s.CommitOperation(op, id, &rec, revision); err == nil {
				t.Fatal("invalid commit accepted")
			}
		})
	}
}
func TestIdentityForPathLostInstallationFailsClosed(t *testing.T) {
	s, id, _ := identifiedRecord(t)
	if err := os.Remove(filepath.Join(s.stateDir, "installation.json")); err != nil {
		t.Fatal(err)
	}
	fresh, err := NewStore(s.stateDir)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fresh.IdentityForPath(id.CanonicalPath); !errors.Is(err, ErrOwnerMismatch) {
		t.Fatalf("lost installation: %v", err)
	}
	if _, err := fresh.LoadIdentity("00000000-0000-0000-0000-000000000000"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("absent identity: %v", err)
	}
}

func TestLifecycleCommitFirstProjection(t *testing.T) {
	s, op, id, rec := preparedCommit(t)
	if err := s.Remove("demo"); err != nil {
		t.Fatal(err)
	}
	s.transactionFault = func(name string) error {
		if name == "identity" {
			return errors.New("crash")
		}
		return nil
	}
	if err := s.CommitOperation(op, id, &rec, id.Revision); err == nil {
		t.Fatal("fault missing")
	}
	fresh, err := NewStore(s.stateDir)
	if err != nil {
		t.Fatal(err)
	}
	if err := fresh.RecoverOperation(id.ProjectID); err != nil {
		t.Fatal(err)
	}
	loaded, err := fresh.Load("demo")
	if err != nil || loaded.Project.Name != "updated" {
		t.Fatalf("first projection: %+v %v", loaded, err)
	}
}

func TestLifecycleCommitPreservesJournalResources(t *testing.T) {
	for _, scenario := range []string{"omitted", "provenance", "tombstone", "retained", "deleted-leftover"} {
		t.Run(scenario, func(t *testing.T) {
			s, op, id, rec := preparedCommit(t)
			resource := ResourceRecord{Kind: "volume", ID: "created", Name: "database", InstallationID: id.InstallationID, ProjectID: id.ProjectID, Role: "database", CreationOperation: op.ID}
			if scenario == "deleted-leftover" {
				resource.Deleted = true
				op.Leftovers = []ResourceRecord{resource}
			} else {
				op.Resources = []ResourceRecord{resource}
			}
			if err := s.SaveOperation(op); err != nil {
				t.Fatal(err)
			}
			switch scenario {
			case "provenance":
				changed := resource
				changed.Role = "other"
				id.Resources = []ResourceRecord{changed}
			case "tombstone":
				changed := resource
				changed.Deleted = true
				id.Resources = []ResourceRecord{changed}
			case "retained", "deleted-leftover":
				id.Resources = []ResourceRecord{resource}
			}
			err := s.CommitOperation(op, id, &rec, id.Revision)
			if scenario == "retained" || scenario == "deleted-leftover" {
				if err != nil {
					t.Fatal(err)
				}
			} else {
				if !errors.Is(err, ErrOwnerMismatch) {
					t.Fatalf("forgotten resource: %v", err)
				}
			}
		})
	}
}

func TestLifecycleRecoveryRefusesUnresolvedRuntimeJournal(t *testing.T) {
	s, op, id, _ := preparedCommit(t)
	fresh, err := NewStore(s.stateDir)
	if err != nil {
		t.Fatal(err)
	}
	if err := fresh.RecoverOperation(id.ProjectID); !errors.Is(err, ErrCollision) {
		t.Fatalf("unresolved route journal: %v", err)
	}
	current, err := fresh.operation(op.ID)
	if err != nil || current.Phase != "route" {
		t.Fatalf("journal changed: %+v %v", current, err)
	}
}

func TestLifecycleRecoveryRefusesMultiplePendingJournals(t *testing.T) {
	s, op, id, rec := preparedCommit(t)
	s.transactionFault = func(name string) error {
		if name == "journal" {
			return errors.New("crash")
		}
		return nil
	}
	if err := s.CommitOperation(op, id, &rec, id.Revision); err == nil {
		t.Fatal("missing fault")
	}
	duplicate := op
	duplicate.ID, _ = newUUID()
	if err := durableJSON(filepath.Join(s.stateDir, "operations", duplicate.ID+".json"), duplicate); err != nil {
		t.Fatal(err)
	}
	fresh, err := NewStore(s.stateDir)
	if err != nil {
		t.Fatal(err)
	}
	if err := fresh.RecoverOperation(id.ProjectID); !errors.Is(err, ErrCollision) {
		t.Fatalf("multiple pending: %v", err)
	}
}

func TestLifecycleCommitRetainsRecreatedIncarnations(t *testing.T) {
	for _, omitTombstone := range []bool{false, true} {
		t.Run(map[bool]string{false: "retained", true: "omitted"}[omitTombstone], func(t *testing.T) {
			s, op, id, rec := preparedCommit(t)
			oldOperation, err := newUUID()
			if err != nil {
				t.Fatal(err)
			}
			tombstone := ResourceRecord{Kind: "volume", ID: "named-volume", Name: "named-volume", InstallationID: id.InstallationID, ProjectID: id.ProjectID, Role: "database", CreationOperation: oldOperation, Deleted: true}
			prior, err := s.LoadIdentity(id.ProjectID)
			if err != nil {
				t.Fatal(err)
			}
			prior.Resources = []ResourceRecord{tombstone}
			if err := s.SaveIdentity(prior, prior.Revision); err != nil {
				t.Fatal(err)
			}
			id.Revision++
			recreated := tombstone
			recreated.CreationOperation = op.ID
			recreated.Deleted = false
			op.Resources = []ResourceRecord{recreated}
			if err := s.SaveOperation(op); err != nil {
				t.Fatal(err)
			}
			id.Resources = []ResourceRecord{recreated}
			if !omitTombstone {
				id.Resources = append([]ResourceRecord{tombstone}, id.Resources...)
			}
			err = s.CommitOperation(op, id, &rec, id.Revision)
			if omitTombstone {
				if !errors.Is(err, ErrOwnerMismatch) {
					t.Fatalf("tombstone lost: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			retained, err := s.LoadIdentity(id.ProjectID)
			if err != nil || len(retained.Resources) != 2 || !retained.Resources[0].Deleted || retained.Resources[1].Deleted {
				t.Fatalf("incarnations: %+v %v", retained, err)
			}
		})
	}
}

func TestLifecycleCommitRefusesUnresolvedResourceIntent(t *testing.T) {
	s, op, id, rec := preparedCommit(t)
	resource := ResourceRecord{Kind: "volume", ID: "intended", Name: "intended", InstallationID: id.InstallationID, ProjectID: id.ProjectID, Role: "database", CreationOperation: op.ID, Planned: true}
	op.Resources = []ResourceRecord{resource}
	if err := s.SaveOperation(op); err != nil {
		t.Fatal(err)
	}
	id.Resources = []ResourceRecord{resource}
	if err := s.CommitOperation(op, id, &rec, id.Revision); !errors.Is(err, ErrInvalidIdentity) {
		t.Fatalf("planned commit: %v", err)
	}
	op.Resources[0].Planned = false
	id.Resources[0].Planned = false
	if err := s.SaveOperation(op); err != nil {
		t.Fatal(err)
	}
	if err := s.CommitOperation(op, id, &rec, id.Revision); err != nil {
		t.Fatal(err)
	}
}
