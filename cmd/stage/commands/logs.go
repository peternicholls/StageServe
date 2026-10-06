// stage logs: stream logs for one service.
package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/peternicholls/stageserve/infra/applecontainer"
	"github.com/peternicholls/stageserve/observability/logs"
)

func NewLogs(flags *SharedFlags) *cobra.Command {
	var (
		service         string
		projectSelector string
		follow          bool
	)
	cmd := &cobra.Command{
		Use:   "logs [service]",
		Short: "View project logs",
		Long:  "Shows log output for the current project by default. Use [service] or --service to choose a service, or --project to inspect another recorded project.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			serviceName, err := resolveLogService(service, args)
			if err != nil {
				return err
			}
			cfg, err := loadConfig(flags)
			if err != nil {
				return err
			}
			cfg, err = recordedProjectForRead(cfg, projectSelector)
			if err != nil {
				return fmt.Errorf("logs: cannot resolve recorded project: %w", err)
			}
			ctx, cancel := contextWithSignal(cmd.Context())
			defer cancel()
			s := &logs.Streamer{Runtime: applecontainer.NewManager(nil)}
			return s.Stream(ctx, cfg.ComposeProjectName, serviceName, follow, cmd.OutOrStdout())
		},
	}
	cmd.Flags().StringVar(&service, "service", "", "Project service name (default: nginx)")
	cmd.Flags().StringVar(&projectSelector, "project", "", "Recorded project selector (slug, name, hostname, or path)")
	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "Follow log output")
	return cmd
}

func resolveLogService(flagValue string, args []string) (string, error) {
	service := flagValue
	if len(args) > 0 {
		if service != "" && service != args[0] {
			return "", fmt.Errorf("logs: use either [service] or --service, not both")
		}
		service = args[0]
	}
	if service == "" {
		return "nginx", nil
	}
	return service, nil
}
