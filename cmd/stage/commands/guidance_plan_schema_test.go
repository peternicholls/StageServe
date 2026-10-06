package commands

import (
	"bytes"
	"encoding/json"
	"io"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/peternicholls/stageserve/core/guidance"
)

func TestGuidancePlanVersionOneJSONContract(t *testing.T) {
	const secret = "schema-test-private-password"
	t.Setenv("MYSQL_PASSWORD", secret)
	t.Setenv("MYSQL_ROOT_PASSWORD", secret)
	root := NewRoot("test")
	var out, stderr bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&stderr)
	projectDir := t.TempDir()
	root.SetArgs([]string{"guidance-plan", "--stack-home", t.TempDir(), "--project-dir", projectDir, "--skip-readiness"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("unexpected stderr: %s", &stderr)
	}
	for _, forbidden := range []string{"\x1b", secret, "MYSQL_PASSWORD", "MYSQL_ROOT_PASSWORD"} {
		if strings.Contains(out.String(), forbidden) {
			t.Fatalf("JSON contains %q", forbidden)
		}
	}
	dec := json.NewDecoder(&out)
	var got map[string]json.RawMessage
	if err := dec.Decode(&got); err != nil {
		t.Fatal(err)
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		t.Fatalf("expected single document, got %v", err)
	}
	if string(got["schema_version"]) != "1" {
		t.Fatalf("schema_version=%s", got["schema_version"])
	}
	var scope map[string]any
	if err := json.Unmarshal(got["project_scope"], &scope); err != nil {
		t.Fatal(err)
	}
	canonicalDir, err := filepath.EvalSymlinks(projectDir)
	if err != nil {
		t.Fatal(err)
	}
	if scope["dir"] != canonicalDir || scope["slug"] == "" {
		t.Fatalf("project scope=%#v", scope)
	}
	if _, exists := scope["project_id"]; exists {
		t.Fatal("unregistered project must not infer a UUID")
	}
	var actions []map[string]any
	if err := json.Unmarshal(got["decision_items"], &actions); err != nil {
		t.Fatal(err)
	}
	if len(actions) == 0 || actions[0]["ID"] != "init" {
		t.Fatalf("existing action key casing changed: %#v", actions)
	}
	if _, exists := actions[0]["id"]; exists {
		t.Fatal("nested action keys must preserve Go field casing")
	}
}

func TestGuidancePlanOptionalFieldOmissionContract(t *testing.T) {
	raw, err := json.Marshal(guidancePlanView{SchemaVersion: 1, Situation: guidance.SituationUnknownError, StatusHeader: "Unavailable"})
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	want := map[string]any{"schema_version": float64(1), "situation": "unknown_error", "status_header": "Unavailable"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("optional field omissions changed: %#v", got)
	}
}

func TestGuidancePlanConfigErrorOmitsUnresolvedScope(t *testing.T) {
	root := NewRoot("test")
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"guidance-plan", "--stack-home", t.TempDir(), "--project-dir", filepath.Join(t.TempDir(), "missing"), "--skip-readiness"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	var got map[string]json.RawMessage
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if _, exists := got["project_scope"]; exists {
		t.Fatalf("unresolved scope must be omitted: %s", out.String())
	}
	if string(got["schema_version"]) != "1" {
		t.Fatalf("schema_version=%s", got["schema_version"])
	}
}
