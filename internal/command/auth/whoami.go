package auth

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
		rest := squareCli.Rest()
		self, err := rest.SelfUser()
		if err != nil || self.Name == "" {
			fmt.Fprintf(squareCli.Out(), "No user associated with current Square Cloud Token\n")
			return err
		}

		username := ui.GreenText.SetString(self.Name)

		fmt.Fprintf(squareCli.Out(), "Currently logged as %s\n", username)
		return
	}
}
