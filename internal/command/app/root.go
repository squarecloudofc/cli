package app

import (
	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/command/app/commit"
	"github.com/squarecloudofc/cli/internal/command/backup"
)

type RunEFunc func(cmd *cobra.Command, args []string) error

func NewAppCommand(squareCli cli.SquareCLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "app",
		Short: "Do some actions with your applications",
		RunE:  runAppCommand(squareCli),
	}

	cmd.AddCommand(
		NewUploadCommand(squareCli),
		commit.NewCommand(squareCli),

		NewStartCommand(squareCli),
		NewRestartCommand(squareCli),
		NewStopCommand(squareCli),

		backup.NewCommand(squareCli),
		NewDeleteCommand(squareCli),
		NewLogsCommand(squareCli),
		NewStatusCommand(squareCli),
		NewListCommand(squareCli),
	)

	return cmd
}

func runAppCommand(_ cli.SquareCLI) RunEFunc {
	return func(cmd *cobra.Command, args []string) (err error) {
		cmd.Help()

		return nil
	}
}
