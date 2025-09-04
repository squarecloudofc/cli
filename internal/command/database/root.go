package database

import (
	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/command/app/snapshot"
)

type RunEFunc func(cmd *cobra.Command, args []string) error

func NewCommand(squareCli cli.SquareCLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "database",
		Short: squareCli.I18n().T("metadata.commands.app.root.short"),
		RunE:  runDatabaseCommand(squareCli),
	}

	cmd.AddCommand(
		snapshot.NewCommand(squareCli),
	)

	return cmd
}

func runDatabaseCommand(_ cli.SquareCLI) RunEFunc {
	return func(cmd *cobra.Command, args []string) (err error) {
		cmd.Help()

		return nil
	}
}
