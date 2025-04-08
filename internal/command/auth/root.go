package auth

import (
	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
)

type RunEFunc func(cmd *cobra.Command, args []string) error

func NewAuthCommand(squareCli *cli.SquareCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage your authentication with Square Cloud",
		RunE:  runAppCommand(squareCli),
	}

	cmd.AddCommand(
		NewLoginCommand(squareCli),
		NewLogoutCommand(squareCli),
		NewWhoamiCommand(squareCli),
	)

	return cmd
}

func runAppCommand(_ *cli.SquareCli) RunEFunc {
	return func(cmd *cobra.Command, args []string) (err error) {
		cmd.Help()

		return nil
	}
}
