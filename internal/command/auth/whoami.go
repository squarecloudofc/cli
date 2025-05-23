package auth

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/ui"
)

func NewWhoamiCommand(squareCli cli.SquareCLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "whoami",
		Short: "Print the user informations associated with current Square Cloud Token",
		RunE:  runWhoamiCommand(squareCli),
	}

	return cmd
}

func runWhoamiCommand(squareCli cli.SquareCLI) RunEFunc {
	return func(cmd *cobra.Command, args []string) (err error) {
		rest := squareCli.Rest()
		self, err := rest.SelfUser()
		if err != nil || self.Name == "" {
			fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.whoami.none"))
			return err
		}

		username := ui.TextGreen.SetString(self.Name)

		fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.whoami.logged", map[string]any{"User": username}))
		return
	}
}
