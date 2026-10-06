package applecontainer

import (
	"context"
	"os"
	"testing"
)

func TestServiceReadinessRequiresRunningStatus(t *testing.T) {
	unregistered, err := os.ReadFile("testdata/1.4.1/system-status.stdout")
	if err != nil {
		t.Fatal(err)
	}
	// The unregistered payload is captured from the real CLI. Other cases are
	// synthetic discriminator tests, not evidence of a running Apple runtime.
	for _, tc := range []struct {
		name  string
		body  []byte
		ready bool
	}{
		{"captured-unregistered", unregistered, false},
		{"not-running", []byte(`{"status":"not running"}`), false},
		{"unknown", []byte(`{"status":"starting"}`), false},
		{"missing", []byte(`{}`), false},
		{"null", []byte(`null`), false},
		{"empty", nil, false},
		{"malformed", []byte(`{`), false},
		{"running-additive", []byte(`{"status":"running","client":{"version":"1.4.1"}}`), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			runner := &fakeRunner{outputs: map[string][]byte{"system status --format json": tc.body}}
			if err := NewManager(runner).Available(context.Background()); (err == nil) != tc.ready {
				t.Fatalf("Available error=%v, want ready=%v", err, tc.ready)
			}
			probe := Probe{Runner: runner, GOOS: "darwin", GOARCH: "arm64",
				OSVersion: func(context.Context) (string, error) { return "26.0", nil },
				LookPath:  func(string) (string, error) { return "/usr/local/bin/container", nil }}
			if result := probe.Check(context.Background()); result.ServiceRunning != tc.ready {
				t.Fatalf("Probe=%+v, want ready=%v", result, tc.ready)
			}
		})
	}
}
