package command

import (
	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/command/app"
	"github.com/squarecloudofc/cli/internal/command/auth"
)

type RunEFunc func(cmd *cobra.Command, args []string) error

func AddCommands(cmd *cobra.Command, squareCli *cli.SquareCli) {
	cmd.AddCommand(
		NewAppsCommand(squareCli),
		NewCommitCommand(squareCli),
		NewUploadCommand(squareCli),
		NewZipCommand(squareCli),

		app.NewAppCommand(squareCli),
		auth.NewAuthCommand(squareCli),
	)
}
