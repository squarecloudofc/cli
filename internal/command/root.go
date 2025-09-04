package command

import (
	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/command/app"
	"github.com/squarecloudofc/cli/internal/command/auth"
	// "github.com/squarecloudofc/cli/internal/command/snapshot"
)

type RunEFunc func(cmd *cobra.Command, args []string) error

func AddCommands(cmd *cobra.Command, squareCli cli.SquareCLI) {
	cmd.AddCommand(
		NewZipCommand(squareCli),

		app.NewAppCommand(squareCli),
		app.NewUploadCommand(squareCli),
		app.NewCommitCommand(squareCli),
		// snapshot.NewCommand(squareCli),

		auth.NewAuthCommand(squareCli),
	)
}
