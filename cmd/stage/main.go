// stage CLI entrypoint. Wires cobra subcommands to the lifecycle
// orchestrator.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/peternicholls/stageserve/cmd/stage/commands"
	"github.com/spf13/cobra"
)

// version is overridden at build time via -ldflags.
var version = "dev"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	code := executeCommand(ctx, commands.NewRoot(version), os.Stderr)
	stop()
	if code != 0 {
		os.Exit(code)
	}
}

// executeCommand keeps command execution and process exit semantics testable.
func executeCommand(ctx context.Context, root *cobra.Command, stderr io.Writer) int {
	err := root.ExecuteContext(ctx)
	if err == nil {
		return 0
	}
	// Subprocesses may report a killed process rather than context.Canceled.
	// The root context records cancellation from either SIGINT or SIGTERM.
	if errors.Is(err, context.Canceled) || errors.Is(ctx.Err(), context.Canceled) {
		fmt.Fprintln(stderr, commands.RenderCommandError(context.Canceled))
		return 130
	}
	var exitCoder commands.ExitCoder
	if errors.As(err, &exitCoder) {
		if !exitCoder.Silent() {
			fmt.Fprintln(stderr, err)
		}
		return exitCoder.ExitCode()
	}
	fmt.Fprintln(stderr, commands.RenderCommandError(err))
	return 1
}
