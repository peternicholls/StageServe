package onboarding_test

import (
	"bytes"
	"encoding/json"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/peternicholls/stageserve/core/onboarding"
)

func TestJSONProjectorVersionOneContract(t *testing.T) {
	for _, status := range []onboarding.Status{onboarding.StatusReady, onboarding.StatusNeedsAction, onboarding.StatusError} {
		t.Run(string(status), func(t *testing.T) {
			result := onboarding.BuildResult([]onboarding.StepResult{{ID: "check", Label: "Check", Status: status, Message: "Checked"}}, nil, nil)
			result.SchemaVersion = 999 // Callers cannot accidentally change the emitted contract.
			var out bytes.Buffer
			if err := (&onboarding.JSONProjector{W: &out}).Project(result); err != nil {
				t.Fatal(err)
			}
			if strings.Contains(out.String(), "\x1b") {
				t.Fatal("JSON contains terminal escapes")
			}
			decoder := json.NewDecoder(&out)
			var got map[string]any
			if err := decoder.Decode(&got); err != nil {
				t.Fatal(err)
			}
			var extra any
			if err := decoder.Decode(&extra); err != io.EOF {
				t.Fatalf("expected exactly one JSON document, got %v", err)
			}
			want := map[string]any{"schema_version": float64(1), "overall_status": string(result.OverallStatus), "exit_code": float64(result.ExitCode), "steps": []any{map[string]any{"id": "check", "label": "Check", "status": string(status), "message": "Checked", "remediation": nil}}}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("contract changed: got %#v want %#v", got, want)
			}
			if result.SchemaVersion != 999 {
				t.Fatal("projector mutated caller result")
			}
		})
	}
}

func TestJSONProjectorPreservesOptionalPayloads(t *testing.T) {
	remediation := "stage setup"
	result := onboarding.BuildResult([]onboarding.StepResult{{ID: "runtime", Label: "Runtime", Status: onboarding.StatusNeedsAction, Message: "Unavailable", Remediation: &remediation, Code: "runtime-unavailable", Meta: map[string]any{"checked": true}}}, map[string]any{"path": "/project/.env"}, []string{"stage doctor"})
	var out bytes.Buffer
	if err := (&onboarding.JSONProjector{W: &out}).Project(result); err != nil {
		t.Fatal(err)
	}
	var got onboarding.CommandResult
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	result.SchemaVersion = 1
	if !reflect.DeepEqual(got, result) {
		t.Fatalf("payload changed: %#v", got)
	}
}

func TestJSONProjectorEmptyStepsContract(t *testing.T) {
	var out bytes.Buffer
	if err := (&onboarding.JSONProjector{W: &out}).Project(onboarding.BuildResult(nil, nil, nil)); err != nil {
		t.Fatal(err)
	}
	var got map[string]json.RawMessage
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if string(got["steps"]) != "[]" {
		t.Fatalf("steps=%s", got["steps"])
	}
	for _, key := range []string{"result", "next_steps"} {
		if _, exists := got[key]; exists {
			t.Fatalf("empty optional %s must remain omitted", key)
		}
	}
}

func TestJSONProjectorProjectScopeContract(t *testing.T) {
	for _, projectID := range []string{"", "validated-project-id"} {
		result := onboarding.BuildResult(nil, nil, nil)
		result.ProjectScope = &onboarding.ProjectScope{Dir: "/project", Slug: "project", ProjectID: projectID}
		var out bytes.Buffer
		if err := (&onboarding.JSONProjector{W: &out}).Project(result); err != nil {
			t.Fatal(err)
		}
		var envelope map[string]json.RawMessage
		if err := json.Unmarshal(out.Bytes(), &envelope); err != nil {
			t.Fatal(err)
		}
		var scope map[string]any
		if err := json.Unmarshal(envelope["project_scope"], &scope); err != nil {
			t.Fatal(err)
		}
		want := map[string]any{"dir": "/project", "slug": "project"}
		if projectID != "" {
			want["project_id"] = projectID
		}
		if !reflect.DeepEqual(scope, want) {
			t.Fatalf("scope=%#v want %#v", scope, want)
		}
	}
}
