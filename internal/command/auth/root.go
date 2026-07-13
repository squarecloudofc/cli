package auth

import (
	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
)

func NewAuthCommand(squareCli cli.SquareCLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: squareCli.I18n().T("metadata.commands.auth.root.short"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(
		NewLoginCommand(squareCli),
		NewLogoutCommand(squareCli),
		NewWhoamiCommand(squareCli),
	)

	return cmd
}
