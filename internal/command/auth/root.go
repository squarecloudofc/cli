package auth

import (
	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
)

type RunEFunc func(cmd *cobra.Command, args []string) error

func NewAuthCommand(squareCli cli.SquareCLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: squareCli.I18n().T("metadata.commands.auth.root.short"),
		RunE:  runAppCommand(squareCli),
	}

	cmd.AddCommand(
		NewLoginCommand(squareCli),
		NewLogoutCommand(squareCli),
		NewWhoamiCommand(squareCli),
	)

	return cmd
}

func runAppCommand(_ cli.SquareCLI) RunEFunc {
	return func(cmd *cobra.Command, args []string) (err error) {
		cmd.Help()

		return nil
	}
}
