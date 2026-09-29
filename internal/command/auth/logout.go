package auth

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
)

func NewLogoutCommand(squareCli cli.SquareCLI) *cobra.Command {
	return &cobra.Command{
		Use:         "logout",
		Short:       squareCli.I18n().T("metadata.commands.auth.logout.short"),
		Annotations: map[string]string{"skipAuthCheck": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			configuration := squareCli.Config()
			if configuration.AuthToken == "" { // already logged out: not an error
				fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.auth.logout.none"))
				return nil
			}

			configuration.SetKey("", "", nil, nil)
			if err := configuration.Save(); err != nil {
				return err
			}

			fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.auth.logout.success"))
			return nil
		},
	}
}
