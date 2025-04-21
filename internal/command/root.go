package command

import (
	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/command/app"
	"github.com/squarecloudofc/cli/internal/command/app/commit"
	"github.com/squarecloudofc/cli/internal/command/app/upload"
	"github.com/squarecloudofc/cli/internal/command/auth"
	"github.com/squarecloudofc/cli/internal/command/backup"
)

type RunEFunc func(cmd *cobra.Command, args []string) error

func AddCommands(cmd *cobra.Command, squareCli cli.SquareCLI) {
	cmd.AddCommand(
		NewZipCommand(squareCli),

		app.NewAppCommand(squareCli),
		auth.NewAuthCommand(squareCli),

		upload.NewCommand(squareCli),
		commit.NewCommand(squareCli),
		backup.NewCommand(squareCli),
	)
}
