package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/peternicholls/stageserve/cmd/stage/commands"
	"github.com/spf13/cobra"
)

type commandExitError struct {
	code   int
	silent bool
}

func (e commandExitError) Error() string { return "operational failure" }
func (e commandExitError) ExitCode() int { return e.code }
func (e commandExitError) Silent() bool  { return e.silent }

func TestExecuteCommandExitContract(t *testing.T) {
	for _, tc := range []struct {
		name       string
		err        error
		cancel     bool
		want       int
		wantOutput bool
	}{
		{name: "success", want: 0},
		{name: "validation failure", err: errors.New("invalid input"), want: 1, wantOutput: true},
		{name: "operational code", err: commandExitError{code: 3}, want: 3, wantOutput: true},
		{name: "wrapped silent operational code", err: fmt.Errorf("setup: %w", commandExitError{code: 2, silent: true}), want: 2},
		{name: "canceled error", err: context.Canceled, want: 130, wantOutput: true},
		{name: "wrapped canceled error", err: fmt.Errorf("logs: %w", context.Canceled), want: 130, wantOutput: true},
		{name: "signal killed subprocess", err: errors.New("signal: killed"), cancel: true, want: 130, wantOutput: true},
		{name: "completed before signal", cancel: true, want: 0},
		{name: "deadline failure", err: context.DeadlineExceeded, want: 1, wantOutput: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			root := &cobra.Command{Use: "stage", SilenceErrors: true, SilenceUsage: true,
				RunE: func(cmd *cobra.Command, args []string) error {
					if tc.cancel {
						cancel()
					}
					return tc.err
				},
			}
			root.SetArgs(nil)
			var stderr bytes.Buffer
			if got := executeCommand(ctx, root, &stderr); got != tc.want {
				t.Fatalf("exit = %d, want %d", got, tc.want)
			}
			if got := stderr.Len() > 0; got != tc.wantOutput {
				t.Fatalf("stderr = %q, want output %v", stderr.String(), tc.wantOutput)
			}
		})
	}
}

func TestExecuteCommandProductionRootValidation(t *testing.T) {
	root := commands.NewRoot("test")
	root.SetArgs([]string{"unknown-command"})
	var stderr bytes.Buffer
	if got := executeCommand(context.Background(), root, &stderr); got != 1 {
		t.Fatalf("exit = %d, want 1", got)
	}
	if !strings.Contains(stderr.String(), "unknown command") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}
