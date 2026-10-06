package state

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestInstallationCommitRecoveryIsolation(t *testing.T) {
	for _, boundary := range []string{"journal", "identity", "committed"} {
		t.Run(boundary, func(t *testing.T) {
			s, project, rec := identifiedRecord(t)
			// A real project named installation must never be inspected or removed by a shared commit.
			reserved, err := s.CreateIdentity(t.TempDir(), "installation")
			if err != nil {
				t.Fatal(err)
			}
			reservedRecord := rec
			reservedRecord.ProjectID = reserved.ProjectID
			reservedRecord.InstallationID = reserved.InstallationID
			reservedRecord.Project.Slug = reserved.Slug
			reservedRecord.Project.Dir = reserved.CanonicalPath
			if err := s.Save(reservedRecord); err != nil {
				t.Fatal(err)
			}
			before, err := os.ReadFile(s.projectFile("installation"))
			if err != nil {
				t.Fatal(err)
			}
			shared, err := s.InstallationIdentity(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			if shared.ProjectID != shared.InstallationID || shared.ProjectID == project.ProjectID {
				t.Fatal("ownership scope confusion")
			}
			op, err := s.BeginOperation(shared.ProjectID, shared.Revision, "gateway", "desired")
			if err != nil {
				t.Fatal(err)
			}
			resource := ResourceRecord{Kind: "container", ID: "gateway", Name: "shared", InstallationID: shared.InstallationID, ProjectID: shared.ProjectID, Role: "gateway", CreationOperation: op.ID}
			op.Resources = []ResourceRecord{resource}
			op.Phase = "route"
			if err := s.SaveOperation(op); err != nil {
				t.Fatal(err)
			}
			shared.Resources = []ResourceRecord{resource}
			shared.LastWriter = op.ID
			fault := errors.New("crash")
			s.transactionFault = func(name string) error {
				if name == boundary {
					return fault
				}
				return nil
			}
			if err := s.CommitInstallationOperation(op, shared, shared.Revision); !errors.Is(err, fault) {
				t.Fatalf("boundary: %v", err)
			}
			fresh, err := NewStore(s.stateDir)
			if err != nil {
				t.Fatal(err)
			}
			if err := fresh.RecoverOperation(shared.ProjectID); err != nil {
				t.Fatal(err)
			}
			got, err := fresh.LoadIdentity(shared.ProjectID)
			if err != nil {
				t.Fatal(err)
			}
			shared.Revision++
			if !reflect.DeepEqual(got, shared) {
				t.Fatalf("shared target: %+v", got)
			}
			after, err := os.ReadFile(s.projectFile("installation"))
			if err != nil || !bytes.Equal(before, after) {
				t.Fatal("shared commit touched project registry")
			}
			rows, err := fresh.Registry()
			if err != nil || len(rows) != 2 {
				t.Fatalf("registry: %+v %v", rows, err)
			}
			if err := fresh.CommitOperation(op, shared, &reservedRecord, shared.Revision); !errors.Is(err, ErrOwnerMismatch) {
				t.Fatalf("project commit accepted shared owner: %v", err)
			}
			if err := fresh.CommitInstallationOperation(Operation{ProjectID: project.ProjectID, InstallationID: project.InstallationID}, project, project.Revision); !errors.Is(err, ErrOwnerMismatch) {
				t.Fatalf("shared commit accepted project owner: %v", err)
			}
		})
	}
}

func TestInstallationLedgerFailsClosed(t *testing.T) {
	for _, scenario := range []string{"lost-installation", "lost-shared-ledger", "wrong-owner", "payload-project", "payload-record"} {
		t.Run(scenario, func(t *testing.T) {
			s, err := NewStore(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			shared, err := s.InstallationIdentity(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			op, err := s.BeginOperation(shared.ProjectID, shared.Revision, "gateway", "desired")
			if err != nil {
				t.Fatal(err)
			}
			op.Phase = "route"
			if err := s.SaveOperation(op); err != nil {
				t.Fatal(err)
			}
			s.transactionFault = func(name string) error {
				if name == "journal" {
					return errors.New("crash")
				}
				return nil
			}
			if err := s.CommitInstallationOperation(op, shared, shared.Revision); err == nil {
				t.Fatal("missing crash")
			}
			switch scenario {
			case "lost-installation":
				if err := os.Remove(filepath.Join(s.stateDir, "installation.json")); err != nil {
					t.Fatal(err)
				}
			case "lost-shared-ledger":
				if err := os.Remove(filepath.Join(s.stateDir, "installation-resources.json")); err != nil {
					t.Fatal(err)
				}
			case "wrong-owner":
				shared.ProjectID = "00000000-0000-0000-0000-000000000000"
				if err := durableJSON(filepath.Join(s.stateDir, "installation-resources.json"), shared); err != nil {
					t.Fatal(err)
				}
			default:
				var journal Operation
				if err := readJSON(filepath.Join(s.stateDir, "operations", op.ID+".json"), &journal); err != nil {
					t.Fatal(err)
				}
				if scenario == "payload-project" {
					journal.Commit.Installation = false
				} else {
					journal.Commit.Record = &Record{}
				}
				if err := durableJSON(filepath.Join(s.stateDir, "operations", op.ID+".json"), journal); err != nil {
					t.Fatal(err)
				}
			}
			fresh, err := NewStore(s.stateDir)
			if err != nil {
				t.Fatal(err)
			}
			if err := fresh.RecoverOperation(op.ProjectID); err == nil {
				t.Fatal("recovery adopted damaged shared state")
			}
			if scenario == "lost-shared-ledger" {
				if _, err := fresh.InstallationIdentity(shared.CanonicalPath); !errors.Is(err, ErrOwnerMismatch) {
					t.Fatalf("lost ledger adopted: %v", err)
				}
			}
			if scenario == "lost-installation" {
				if _, err := fresh.InstallationIdentity(shared.CanonicalPath); !errors.Is(err, ErrOwnerMismatch) {
					t.Fatalf("lost installation adopted: %v", err)
				}
			}
		})
	}
}

func TestInstallationIdentityCannotBecomeProjectRecord(t *testing.T) {
	s, _, rec := identifiedRecord(t)
	shared, err := s.InstallationIdentity(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	rec.ProjectID = shared.ProjectID
	rec.InstallationID = shared.InstallationID
	rec.Project.Slug = shared.Slug
	rec.Project.Dir = shared.CanonicalPath
	if err := s.Save(rec); !errors.Is(err, ErrOwnerMismatch) {
		t.Fatalf("shared registry save: %v", err)
	}
	if _, err := os.Stat(s.projectFile("installation")); !os.IsNotExist(err) {
		t.Fatalf("shared projection created: %v", err)
	}
	// Forged project-ledger storage cannot alias the reserved installation ID.
	if err := durableJSON(filepath.Join(s.stateDir, "identities", shared.ProjectID+".json"), shared); err != nil {
		t.Fatal(err)
	}
	if _, err := s.IdentityForPath(shared.CanonicalPath); !errors.Is(err, ErrOwnerMismatch) {
		t.Fatalf("shared owner adopted as project: %v", err)
	}
}

func TestInstallationLedgerHasIndependentSlugNamespace(t *testing.T) {
	s, _, _ := identifiedRecord(t)
	if _, err := s.CreateIdentity(t.TempDir(), "installation"); err != nil {
		t.Fatal(err)
	}
	shared, err := s.InstallationIdentity(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	shared.EngineVersion = "verified"
	if err := s.SaveIdentity(shared, shared.Revision); err != nil {
		t.Fatalf("shared slug collided with project: %v", err)
	}
}
