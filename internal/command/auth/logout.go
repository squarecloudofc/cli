package auth

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
)

func NewLogoutCommand(squareCli *cli.SquareCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Logout from your Square Cloud account",
		RunE:  runLogoutCommand(squareCli),
	}

	cmd.Flags().String("token", "", "")
	return cmd
}

func runLogoutCommand(squareCli *cli.SquareCli) RunEFunc {
	return func(cmd *cobra.Command, args []string) (err error) {
		squareCli.Config.AuthToken = ""
		squareCli.Config.Save()

		fmt.Fprintf(squareCli.Out(), "You have successfully logged out of your Square Cloud account.\n")
		return
	}
}
