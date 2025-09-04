package snapshot

import (
	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
)

func NewCommand(squareCli cli.SquareCLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snapshot",
		Short: squareCli.I18n().T("metadata.commands.snapshot.root.short"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(
		NewCreateCommand(squareCli),
		NewListCommand(squareCli),
	)

	return cmd
}
