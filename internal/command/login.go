package command

import (
	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
)

func NewLoginCommand(squareCli *cli.SquareCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login to Square Cloud",
		RunE:  runLoginCommand(squareCli),
	}

	cmd.PersistentFlags().StringP("search", "s", "", "Search for an application")
	return cmd
}

func runLoginCommand(squareCli *cli.SquareCli) RunEFunc {
	return func(cmd *cobra.Command, args []string) (err error) {

		return nil
	}
}
