package command

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
)

func NewWhoamiCommand(squareCli *cli.SquareCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "whoami",
		Short: "Print the user informations associated with current Square Cloud Token",
		RunE:  runWhoamiCommand(squareCli),
	}

	cmd.PersistentFlags().StringP("search", "s", "", "Search for an application")
	return cmd
}

func runWhoamiCommand(squareCli *cli.SquareCli) RunEFunc {
	return func(cmd *cobra.Command, args []string) (err error) {
		self, err := squareCli.Rest.SelfUser()
		if err != nil {
			return err
		}

		fmt.Fprintf(squareCli.Out(), "currently logged as \x1b[32m%s\x1b[30m\n", self.User.Tag)
		return
	}
}
