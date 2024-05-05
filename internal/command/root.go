package command

import (
	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/command/app"
)

type RunEFunc func(cmd *cobra.Command, args []string) error

func AddCommands(cmd *cobra.Command, squareCli *cli.SquareCli) {
	cmd.AddCommand(
		NewLoginCommand(squareCli),
		NewWhoamiCommand(squareCli),

		NewAppsCommand(squareCli),
		NewCommitCommand(squareCli),
		NewUploadCommand(squareCli),
		NewZipCommand(squareCli),

		app.NewAppCommand(squareCli),
	)
}
