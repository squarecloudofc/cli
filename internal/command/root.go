package command

import (
	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/command/app"
	"github.com/squarecloudofc/cli/internal/command/auth"
	"github.com/squarecloudofc/cli/internal/command/database"
	"github.com/squarecloudofc/cli/internal/command/workspace"
)

func AddCommands(cmd *cobra.Command, squareCli cli.SquareCLI) {
	cmd.AddCommand(
		NewZipCommand(squareCli),
		NewStatusCommand(squareCli),

		app.NewAppCommand(squareCli),
		// upload/commit are also exposed at the top level — the headline UX.
		app.NewUploadCommand(squareCli),
		app.NewCommitCommand(squareCli),

		database.NewCommand(squareCli),
		workspace.NewCommand(squareCli),
		auth.NewAuthCommand(squareCli),
	)
}
