package database

import (
	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/command/database/snapshot"
)

func NewCommand(squareCli cli.SquareCLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "db",
		Aliases: []string{"database"},
		Short:   squareCli.I18n().T("metadata.commands.database.root.short"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(
		newListCommand(squareCli),
		newCreateCommand(squareCli),
		newInfoCommand(squareCli),
		newUpdateCommand(squareCli),
		newDeleteCommand(squareCli),
		newStartCommand(squareCli),
		newStopCommand(squareCli),
		newStatusCommand(squareCli),
		newMetricsCommand(squareCli),
		newCredentialsCommand(squareCli),
		snapshot.NewCommand(squareCli),
	)

	return cmd
}
