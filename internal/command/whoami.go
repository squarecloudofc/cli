package command

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/ui"
)

func NewWhoamiCommand(squareCli *cli.SquareCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "whoami",
		Short: "Print the user informations associated with current Square Cloud Token",
		RunE:  runWhoamiCommand(squareCli),
	}

	return cmd
}

func runWhoamiCommand(squareCli *cli.SquareCli) RunEFunc {
	return func(cmd *cobra.Command, args []string) (err error) {
		self, err := squareCli.Rest.SelfUser()
		if err != nil {
			return err
		}

		if self == nil || self.User.Tag == "" {
			fmt.Fprintf(squareCli.Out(), "No user associated with current Square Cloud Token\n")
			return
		}

		username := ui.GreenText.SetString(self.User.Tag)

		fmt.Fprintf(squareCli.Out(), "Currently logged as %s\n", username)
		return
	}
}
