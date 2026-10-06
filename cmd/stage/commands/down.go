// stage down: stop a project's stack.
package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewDown(flags *SharedFlags) *cobra.Command {
	var removeVolumes bool
	var confirmProject string
	cmd := &cobra.Command{
		Use:   "down",
		Args:  cobra.NoArgs,
		Short: "Stop this project",
		Long:  "Stops the current project and keeps its StageServe record so you can run it again later. Use --all to stop every recorded project.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if flags.All && removeVolumes {
				return fmt.Errorf("--all --volumes is not supported; delete data for one explicitly confirmed project at a time")
			}
			if confirmProject != "" && !removeVolumes {
				return fmt.Errorf("--confirm-project requires --volumes")
			}
			cfg, err := loadConfig(flags)
			if err != nil {
				return err
			}
			cfg.All = flags.All
			if confirmProject != "" && confirmProject != cfg.Slug {
				return fmt.Errorf("--confirm-project must match the selected project %q", cfg.Slug)
			}
			if flags.DryRun {
				if flags.All {
					fmt.Fprintln(cmd.OutOrStdout(), "DRY RUN: would down every recorded project")
					return nil
				}
				if removeVolumes {
					fmt.Fprintf(cmd.OutOrStdout(), "DRY RUN: would stop project %s and delete its owned named volumes\n", cfg.Slug)
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "DRY RUN: would down project %s\n", cfg.Slug)
				}
				return nil
			}
			if removeVolumes {
				if err := authorizeDownVolumes(cmd, cfg, confirmProject); err != nil {
					return err
				}
			}
			orch, err := buildOrchestrator(cfg)
			if err != nil {
				return err
			}
			ctx, cancel := contextWithSignal(cmd.Context())
			defer cancel()
			if flags.All {
				return orch.DownAll(ctx, cfg, removeVolumes)
			}
			return orch.Down(ctx, cfg, removeVolumes)
		},
	}
	cmd.Flags().BoolVarP(&removeVolumes, "volumes", "v", false, "Remove named volumes")
	cmd.Flags().StringVar(&confirmProject, "confirm-project", "", "Confirm deletion of the selected project's volumes by exact slug")
	return cmd
}
