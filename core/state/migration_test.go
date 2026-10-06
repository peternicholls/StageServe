package state

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/peternicholls/stageserve/core/config"
	"github.com/peternicholls/stageserve/core/runtime"
)

// Refusal must protect the original bytes, including when a caller attempts
// to replace an incompatible record with a freshly generated Apple record.
func TestIncompatibleRecordCannotBeOverwritten(t *testing.T) {
	for _, tc := range []struct {
		name, data string
		want       error
	}{
		{"unversioned", `{"project":{"Slug":"demo"}}`, ErrSchema},
		{"older", `{"schema_version":1,"project":{"Slug":"demo"}}`, ErrSchema},
		{"newer", `{"schema_version":999,"project":{"Slug":"demo","RuntimeBackend":"apple-container"}}`, ErrSchema},
		{"missing-backend", `{"schema_version":2,"project":{"Slug":"demo"}}`, ErrLegacyState},
		{"docker", `{"schema_version":2,"project":{"Slug":"demo","RuntimeBackend":"docker"}}`, ErrLegacyState},
		{"unknown", `{"schema_version":2,"project":{"Slug":"demo","RuntimeBackend":"future"}}`, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, err := NewStore(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			path := s.projectFile("demo")
			if err := os.WriteFile(path, []byte(tc.data), 0600); err != nil {
				t.Fatal(err)
			}
			replacement := Record{Project: config.ProjectConfig{Slug: "demo", RuntimeBackend: runtime.BackendAppleContainer}}
			for _, action := range []func() error{
				func() error { _, err := s.Load("demo"); return err },
				func() error { _, _, err := s.StateFileForSelector("demo"); return err },
				func() error { return s.Save(replacement) },
				func() error { return s.Remove("demo") },
			} {
				err := action()
				if err == nil || (tc.want != nil && !errors.Is(err, tc.want)) {
					t.Fatalf("refusal = %v; want %v", err, tc.want)
				}
				got, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(got, []byte(tc.data)) {
					t.Fatal("refusal changed source record")
				}
			}
		})
	}
}

func TestRetainedLedgerOwnerAndSchemaRefusal(t *testing.T) {
	for _, target := range []string{"installation", "identity", "operation"} {
		for _, fault := range []string{"older-schema", "newer-schema", "owner"} {
			t.Run(target+"/"+fault, func(t *testing.T) {
				dir := t.TempDir()
				s, err := NewStore(dir)
				if err != nil {
					t.Fatal(err)
				}
				v, err := s.CreateIdentity(t.TempDir(), "demo")
				if err != nil {
					t.Fatal(err)
				}
				op, err := s.BeginOperation(v.ProjectID, v.Revision, "start", "hash")
				if err != nil {
					t.Fatal(err)
				}
				path := filepath.Join(dir, "identities", v.ProjectID+".json")
				if target == "installation" {
					path = filepath.Join(dir, "installation.json")
				}
				if target == "operation" {
					path = filepath.Join(dir, "operations", op.ID+".json")
				}
				data, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				var doc map[string]any
				if err := json.Unmarshal(data, &doc); err != nil {
					t.Fatal(err)
				}
				want := ErrSchema
				switch fault {
				case "older-schema":
					doc["schema_version"] = 0
				case "newer-schema":
					doc["schema_version"] = 999
				case "owner":
					want = ErrOwnerMismatch
					if target == "installation" {
						doc["id"] = "11111111-1111-4111-8111-111111111111"
					} else {
						doc["installation_id"] = "11111111-1111-4111-8111-111111111111"
					}
				}
				data, err = json.Marshal(doc)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, data, 0600); err != nil {
					t.Fatal(err)
				}
				fresh, err := NewStore(dir)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := fresh.PendingOperations(v.ProjectID); !errors.Is(err, want) {
					t.Fatalf("pending = %v; want %v", err, want)
				}
				if _, err := fresh.BeginOperation(v.ProjectID, v.Revision, "stop", ""); !errors.Is(err, want) {
					t.Fatalf("begin = %v; want %v", err, want)
				}
				got, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(got, data) {
					t.Fatal("refusal changed durable ledger")
				}
			})
		}
	}
}
