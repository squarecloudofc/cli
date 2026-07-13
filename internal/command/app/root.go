package app

import (
	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/command/app/deploy"
	"github.com/squarecloudofc/cli/internal/command/app/env"
	"github.com/squarecloudofc/cli/internal/command/app/file"
	"github.com/squarecloudofc/cli/internal/command/app/network"
	"github.com/squarecloudofc/cli/internal/command/app/snapshot"
)

func NewAppCommand(squareCli cli.SquareCLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "app",
		Short: squareCli.I18n().T("metadata.commands.app.root.short"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(
		NewUploadCommand(squareCli),
		NewCommitCommand(squareCli),

		NewStartCommand(squareCli),
		NewRestartCommand(squareCli),
		NewStopCommand(squareCli),

		NewListCommand(squareCli),
		NewInfoCommand(squareCli),
		NewStatusCommand(squareCli),
		NewLogsCommand(squareCli),
		NewMetricsCommand(squareCli),
		NewRealtimeCommand(squareCli),
		NewDomainsCommand(squareCli),
		NewLoadBalancersCommand(squareCli),
		NewDeleteCommand(squareCli),

		env.NewCommand(squareCli),
		file.NewCommand(squareCli),
		deploy.NewCommand(squareCli),
		network.NewCommand(squareCli),
		snapshot.NewCommand(squareCli),
	)

	return cmd
}
