package commands

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/peternicholls/stageserve/core/state"
)

func TestDownDeletionConsentFailsBeforeRuntime(t *testing.T) {
	for _, tc := range []struct {
		flags []string
		want  string
	}{
		{[]string{"--all", "--volumes"}, "--all --volumes"},
		{[]string{"--confirm-project", "demo"}, "requires --volumes"},
		{[]string{"--volumes"}, "requires --confirm-project demo"},
		{[]string{"--volumes", "--confirm-project", "other"}, "must match"},
	} {
		t.Run(tc.want, func(t *testing.T) {
			dir, stack := t.TempDir(), t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, ".env.stageserve"), []byte("SITE_NAME=demo\n"), 0600); err != nil {
				t.Fatal(err)
			}
			root := NewRoot("test")
			root.SetIn(strings.NewReader("demo\n")) // redirected input must never be read
			var output bytes.Buffer
			root.SetOut(&output)
			root.SetArgs(append([]string{"down", "--project-dir", dir, "--stack-home", stack}, tc.flags...))
			if err := root.Execute(); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error=%v, want %q", err, tc.want)
			}
			if output.Len() != 0 {
				t.Fatalf("unexpected prompt/output: %s", output.String())
			}
			if _, err := os.Stat(filepath.Join(stack, ".stageserve-state")); !os.IsNotExist(err) {
				t.Fatalf("state created before consent: %v", err)
			}
		})
	}
}

func TestVolumeConfirmationDefaultCancelExactScope(t *testing.T) {
	identity := state.Identity{Slug: "demo", ProjectID: "immutable-project", Resources: []state.ResourceRecord{
		{Kind: "volume", ID: "custom-id", Name: "custom-data"},
		{Kind: "volume", ID: "deleted-id", Name: "removed-data", Deleted: true},
		{Kind: "container", ID: "web-id", Name: "web"},
	}}
	for _, input := range []string{"\n", "yes\n", "other\n", "", "demo\n"} {
		var output bytes.Buffer
		err := readVolumeConfirmation(context.Background(), strings.NewReader(input), &output, identity)
		if input == "demo\n" {
			if err != nil {
				t.Fatal(err)
			}
		} else if !errors.Is(err, context.Canceled) {
			t.Fatalf("input=%q error=%v", input, err)
		}
		for _, want := range []string{"immutable-project", "custom-data (custom-id)", "Default: cancel", "files and settings are retained"} {
			if !strings.Contains(output.String(), want) {
				t.Errorf("preview missing %q", want)
			}
		}
		if strings.Contains(output.String(), "removed-data") || strings.Contains(output.String(), "web-id") {
			t.Fatal("preview includes resources that will not be deleted")
		}
	}
}

func TestVolumeConfirmationCancellationDoesNotWaitForInput(t *testing.T) {
	reader, writer := io.Pipe()
	defer reader.Close()
	defer writer.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := readVolumeConfirmation(ctx, reader, io.Discard, state.Identity{Slug: "demo"}); !errors.Is(err, context.Canceled) {
		t.Fatalf("error=%v", err)
	}
}

func TestDownVolumeDryRunNeedsNoConsentOrRuntime(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".env.stageserve"), []byte("SITE_NAME=demo\n"), 0600); err != nil {
		t.Fatal(err)
	}
	root := NewRoot("test")
	var output bytes.Buffer
	root.SetOut(&output)
	root.SetArgs([]string{"down", "--project-dir", dir, "--stack-home", t.TempDir(), "--volumes", "--dry-run"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "delete its owned named volumes") {
		t.Fatal(output.String())
	}
}
