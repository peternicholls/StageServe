package state

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/peternicholls/stageserve/core/config"
	"github.com/peternicholls/stageserve/core/runtime"
)

func identifiedRecord(t *testing.T) (*Store, Identity, Record) {
	t.Helper()
	s, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	id, err := s.CreateIdentity(t.TempDir(), "demo")
	if err != nil {
		t.Fatal(err)
	}
	rec := Record{ProjectID: id.ProjectID, InstallationID: id.InstallationID, Project: config.ProjectConfig{Slug: id.Slug, Dir: id.CanonicalPath, RuntimeBackend: runtime.BackendAppleContainer}, AttachmentState: StateAttached}
	if err := s.Save(rec); err != nil {
		t.Fatal(err)
	}
	return s, id, rec
}

func TestRecordIdentityFreshStoreFailsClosed(t *testing.T) {
	for _, scenario := range []string{"partial", "invalid", "installation", "project", "slug", "path", "lost-installation", "lost-ledger", "unregistered"} {
		t.Run(scenario, func(t *testing.T) {
			s, id, rec := identifiedRecord(t)
			switch scenario {
			case "partial":
				rec.InstallationID = ""
			case "invalid":
				rec.ProjectID = "invalid"
			case "installation":
				rec.InstallationID = "00000000-0000-0000-0000-000000000000"
			case "project":
				other, err := s.CreateIdentity(t.TempDir(), "other")
				if err != nil {
					t.Fatal(err)
				}
				rec.ProjectID = other.ProjectID
			case "slug":
				rec.Project.Slug = "other"
			case "path":
				rec.Project.Dir = t.TempDir()
			case "lost-installation":
				if err := os.Remove(filepath.Join(s.stateDir, "installation.json")); err != nil {
					t.Fatal(err)
				}
			case "lost-ledger":
				if err := os.Remove(filepath.Join(s.stateDir, "identities", id.ProjectID+".json")); err != nil {
					t.Fatal(err)
				}
			case "unregistered":
				id.Registered = false
				if err := s.SaveIdentity(id, id.Revision); err != nil {
					t.Fatal(err)
				}
			}
			rec.SchemaVersion = SchemaVersion
			if err := durableJSON(s.projectFile("demo"), rec); err != nil {
				t.Fatal(err)
			}
			fresh, err := NewStore(s.stateDir)
			if err != nil {
				t.Fatal(err)
			}
			for name, check := range map[string]func() error{
				"load":     func() error { _, e := fresh.Load("demo"); return e },
				"registry": func() error { _, e := fresh.Registry(); return e },
				"selector": func() error { _, _, e := fresh.StateFileForSelector("demo"); return e },
				"save":     func() error { return fresh.Save(rec) },
			} {
				if err := check(); err == nil {
					t.Errorf("%s accepted %s", name, scenario)
				}
			}
			if scenario != "unregistered" {
				if err := fresh.Remove("demo"); err == nil {
					t.Fatal("remove accepted damaged identity")
				}
			}
		})
	}
}

func TestRecordIdentityCannotBeReplacedOrAdopted(t *testing.T) {
	s, _, rec := identifiedRecord(t)
	for _, mutate := range []func(*Record){func(r *Record) { r.ProjectID = ""; r.InstallationID = "" }, func(r *Record) { r.ProjectID = "00000000-0000-0000-0000-000000000000" }} {
		changed := rec
		mutate(&changed)
		if err := s.Save(changed); !errors.Is(err, ErrOwnerMismatch) {
			t.Fatalf("identity replacement: %v", err)
		}
	}
	// A record predating ownership persistence must not become identified by Save.
	if err := s.Remove("demo"); err != nil {
		t.Fatal(err)
	}
	legacy := rec
	legacy.ProjectID = ""
	legacy.InstallationID = ""
	if err := s.Save(legacy); err != nil {
		t.Fatal(err)
	}
	if err := s.Save(rec); !errors.Is(err, ErrOwnerMismatch) {
		t.Fatalf("adoption: %v", err)
	}
}

func TestRecordRemovalRetainsUnregisteredOwnership(t *testing.T) {
	s, id, _ := identifiedRecord(t)
	id.Registered = false
	if err := s.SaveIdentity(id, id.Revision); err != nil {
		t.Fatal(err)
	}
	fresh, err := NewStore(s.stateDir)
	if err != nil {
		t.Fatal(err)
	}
	if err := fresh.Remove("demo"); err != nil {
		t.Fatal(err)
	}
	retained, err := fresh.LoadIdentity(id.ProjectID)
	if err != nil || retained.Registered || retained.CanonicalPath != id.CanonicalPath {
		t.Fatalf("retained ownership: %+v %v", retained, err)
	}
	same, err := fresh.CreateIdentity(id.CanonicalPath, id.Slug)
	if err != nil || same.ProjectID != id.ProjectID || same.Registered {
		t.Fatalf("recreation adopted ownership: %+v %v", same, err)
	}
}

func TestUUIDFreeRecordDoesNotCreateIdentity(t *testing.T) {
	s, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	rec := Record{Project: config.ProjectConfig{Slug: "legacy", RuntimeBackend: runtime.BackendAppleContainer}}
	if err := s.Save(rec); err != nil {
		t.Fatal(err)
	}
	fresh, err := NewStore(s.stateDir)
	if err != nil {
		t.Fatal(err)
	}
	got, err := fresh.Load("legacy")
	if err != nil || got.ProjectID != "" || got.InstallationID != "" {
		t.Fatalf("load: %+v %v", got, err)
	}
	for _, path := range []string{"installation.json", "identities"} {
		if _, err := os.Stat(filepath.Join(s.stateDir, path)); !os.IsNotExist(err) {
			t.Fatalf("ownership created: %s %v", path, err)
		}
	}
}

func TestNewRecordIdentityRequiresExistingRegisteredLedger(t *testing.T) {
	for _, scenario := range []string{"partial", "invalid", "unknown", "unregistered"} {
		t.Run(scenario, func(t *testing.T) {
			s, id, rec := identifiedRecord(t)
			if err := s.Remove("demo"); err != nil {
				t.Fatal(err)
			}
			switch scenario {
			case "partial":
				rec.InstallationID = ""
			case "invalid":
				rec.ProjectID = "invalid"
			case "unknown":
				rec.ProjectID = "00000000-0000-0000-0000-000000000000"
			case "unregistered":
				id.Registered = false
				if err := s.SaveIdentity(id, id.Revision); err != nil {
					t.Fatal(err)
				}
			}
			if err := s.Save(rec); err == nil {
				t.Fatal("saved invalid ownership")
			}
			if _, err := os.Stat(s.projectFile("demo")); !os.IsNotExist(err) {
				t.Fatalf("invalid record persisted: %v", err)
			}
		})
	}
}

func TestRecordIdentityAcceptsCanonicalAlias(t *testing.T) {
	s, id, rec := identifiedRecord(t)
	alias := filepath.Join(t.TempDir(), "alias")
	if err := os.Symlink(id.CanonicalPath, alias); err != nil {
		t.Fatal(err)
	}
	rec.Project.Dir = alias
	if err := s.Save(rec); err != nil {
		t.Fatal(err)
	}
	fresh, err := NewStore(s.stateDir)
	if err != nil {
		t.Fatal(err)
	}
	got, err := fresh.Load("demo")
	if err != nil || got.ProjectID != id.ProjectID {
		t.Fatalf("alias: %+v %v", got, err)
	}
}
