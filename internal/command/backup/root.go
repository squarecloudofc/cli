package backup

import (
	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
)

func NewCommand(squareCli cli.SquareCLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Manage your backups",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(
		NewCreateCommand(squareCli),
	)

	return cmd
}
