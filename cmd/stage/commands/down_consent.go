package commands

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/peternicholls/stageserve/core/config"
	"github.com/peternicholls/stageserve/core/state"
	"github.com/spf13/cobra"
)

func authorizeDownVolumes(cmd *cobra.Command, cfg config.ProjectConfig, confirmed string) error {
	if confirmed == cfg.Slug {
		return nil
	}
	input, inOK := cmd.InOrStdin().(*os.File)
	output, outOK := cmd.OutOrStdout().(*os.File)
	if !inOK || !outOK || !isatty.IsTerminal(input.Fd()) || !isatty.IsTerminal(output.Fd()) {
		return fmt.Errorf("volume deletion requires --confirm-project %s in noninteractive use", cfg.Slug)
	}
	store, err := state.NewStore(cfg.StateDir)
	if err != nil {
		return err
	}
	identity, err := store.IdentityForPath(cfg.Dir)
	if err != nil {
		return err
	}
	record, err := store.Load(cfg.Slug)
	if err != nil {
		return err
	}
	if identity.Slug != cfg.Slug || record.ProjectID != identity.ProjectID || record.InstallationID != identity.InstallationID {
		return state.ErrOwnerMismatch
	}
	return readVolumeConfirmation(cmd.Context(), cmd.InOrStdin(), cmd.OutOrStdout(), identity)
}

// This preview grants consent only; lifecycle revalidates ownership before mutation.
func readVolumeConfirmation(ctx context.Context, input io.Reader, output io.Writer, identity state.Identity) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	var preview strings.Builder
	fmt.Fprintf(&preview, "Delete data for %s (%s):\n", identity.Slug, identity.ProjectID)
	for _, resource := range identity.Resources {
		if resource.Kind == "volume" && !resource.Deleted {
			fmt.Fprintf(&preview, "  %s (%s)\n", resource.Name, resource.ID)
		}
	}
	fmt.Fprintf(&preview, "Project files and settings are retained. Default: cancel.\nType %s to delete these volumes, or press Enter to cancel: ", identity.Slug)
	if _, err := io.WriteString(output, preview.String()); err != nil {
		return err
	}
	answer := make(chan string, 1)
	readError := make(chan error, 1)
	go func() {
		scanner := bufio.NewScanner(input)
		scanner.Buffer(make([]byte, 128), 4096)
		if scanner.Scan() {
			answer <- strings.TrimSpace(scanner.Text())
		} else {
			readError <- scanner.Err()
		}
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-readError:
		if err != nil {
			return err
		}
		return context.Canceled
	case value := <-answer:
		if value != identity.Slug {
			return context.Canceled
		}
		return nil
	}
}
