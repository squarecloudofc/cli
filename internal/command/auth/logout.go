package auth

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
)

func NewLogoutCommand(squareCli cli.SquareCLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Logout from your Square Cloud account",
		RunE:  runLogoutCommand(squareCli),
	}

	cmd.Flags().String("token", "", "")
	return cmd
}

func runLogoutCommand(squareCli cli.SquareCLI) RunEFunc {
	return func(cmd *cobra.Command, args []string) (err error) {
		configuration := squareCli.Config()
		configuration.AuthToken = ""
		configuration.Save()

		fmt.Fprintf(squareCli.Out(), "You have successfully logged out of your Square Cloud account.\n")
		return
	}
}
