package command

import (
	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
)

type RunEFunc func(cmd *cobra.Command, args []string) error

func AddCommands(cmd *cobra.Command, squareCli *cli.SquareCli) {
	cmd.AddCommand(
		NewAppsCommand(squareCli),
	)
}
