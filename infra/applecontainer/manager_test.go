package applecontainer

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	coreruntime "github.com/peternicholls/stageserve/core/runtime"
)

type fakeRunner struct {
	calls   [][]string
	outputs map[string][]byte
	errors  map[string]error
	stream  string
}

func (f *fakeRunner) Run(_ context.Context, args ...string) ([]byte, error) {
	f.calls = append(f.calls, append([]string(nil), args...))
	key := strings.Join(args, " ")
	return f.outputs[key], f.errors[key]
}

func (f *fakeRunner) Stream(_ context.Context, args ...string) (io.ReadCloser, error) {
	f.calls = append(f.calls, append([]string(nil), args...))
	return io.NopCloser(strings.NewReader(f.stream)), nil
}

func TestManagerStartBuildsAppleContainerCommands(t *testing.T) {
	dir := t.TempDir()
	definition := filepath.Join(dir, "runtime.json")
	body := `{
  "volumes":["${PROJECT_DATABASE_VOLUME}"],
  "services":[{
    "name":"web","image":"nginx:alpine",
    "environment":{"SITE":"${PROJECT_HOSTNAME}"},
    "mounts":["${PROJECT_ROOT}:/srv:ro"],
    "ports":["127.0.0.1:${HOST_PORT}:80"]
  }]
}`
	if err := os.WriteFile(definition, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	runner := &fakeRunner{outputs: map[string][]byte{
		"volume list --quiet":      nil,
		"list --all --format json": []byte("[]"),
	}, errors: map[string]error{}}
	manager := NewManager(runner)
	err := manager.Start(context.Background(), coreruntime.StartOptions{
		Definition: definition, ProjectName: "stage-demo",
		Env: []string{"PROJECT_DATABASE_VOLUME=stage-demo-db", "PROJECT_HOSTNAME=demo.test", "PROJECT_ROOT=/tmp/demo", "HOST_PORT=8080"},
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	wantRun := []string{"run", "--detach", "--name", "stage-demo-web",
		"--label", "io.stageserve.project=stage-demo", "--label", "io.stageserve.service=web",
		"--network", "default", "--env", "SITE=demo.test", "--volume", "/tmp/demo:/srv:ro",
		"--publish", "127.0.0.1:8080:80", "nginx:alpine"}
	if !hasCall(runner.calls, wantRun) {
		t.Fatalf("run call missing\ncalls=%v\nwant=%v", runner.calls, wantRun)
	}
	if !hasCall(runner.calls, []string{"volume", "create", "--label", "io.stageserve.managed=true", "stage-demo-db"}) {
		t.Fatalf("volume create missing: %v", runner.calls)
	}
}

func TestManagerListsOnlyProjectServices(t *testing.T) {
	runner := &fakeRunner{outputs: map[string][]byte{
		"list --all --format json": []byte(`[
          {"id":"stage-demo-web","state":"running","ip":"192.168.64.2"},
          {"id":"stage-other-web","state":"running","ip":"192.168.64.3"}
        ]`),
	}, errors: map[string]error{}}
	services, err := NewManager(runner).ListServices(context.Background(), "stage-demo")
	if err != nil {
		t.Fatal(err)
	}
	if len(services) != 1 || services[0].Service != "web" || services[0].Address != "192.168.64.2" {
		t.Fatalf("services=%+v", services)
	}
}

func TestManagerExecAndLogsUseStableServiceName(t *testing.T) {
	runner := &fakeRunner{outputs: map[string][]byte{}, errors: map[string]error{}, stream: "hello\n"}
	manager := NewManager(runner)
	out, err := manager.Exec(context.Background(), coreruntime.ExecOptions{
		ProjectName: "stage-demo", Service: "web", WorkingDir: "/srv", Command: []string{"php", "-v"},
	})
	if err != nil || out != "" {
		t.Fatalf("Exec output=%q err=%v", out, err)
	}
	var logs bytes.Buffer
	manager.Stdout = &logs
	if err := manager.Logs(context.Background(), coreruntime.LogsOptions{ProjectName: "stage-demo", Service: "web", Follow: true}); err != nil {
		t.Fatal(err)
	}
	if logs.String() != "hello\n" {
		t.Fatalf("logs=%q", logs.String())
	}
	if !hasCall(runner.calls, []string{"exec", "--workdir", "/srv", "stage-demo-web", "php", "-v"}) ||
		!hasCall(runner.calls, []string{"logs", "--follow", "stage-demo-web"}) {
		t.Fatalf("calls=%v", runner.calls)
	}
}

func TestProbeIsReadOnlyAndFailClosed(t *testing.T) {
	runner := &fakeRunner{outputs: map[string][]byte{}, errors: map[string]error{
		"system status --format json": errors.New("not running"),
	}}
	probe := Probe{
		Runner: runner, GOOS: "darwin", GOARCH: "arm64",
		OSVersion: func(context.Context) (string, error) { return "26.6.2", nil },
		LookPath:  func(string) (string, error) { return "/usr/local/bin/container", nil },
	}
	result := probe.Check(context.Background())
	if result.ServiceRunning || !strings.Contains(result.Message, "not running") {
		t.Fatalf("result=%+v", result)
	}
	if len(runner.calls) != 1 || !slices.Equal(runner.calls[0], []string{"system", "status", "--format", "json"}) {
		t.Fatalf("readiness performed mutating calls: %v", runner.calls)
	}
}

func hasCall(calls [][]string, want []string) bool {
	for _, call := range calls {
		if slices.Equal(call, want) {
			return true
		}
	}
	return false
}
