package auth

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/ui"
	"github.com/squarecloudofc/cli/internal/ui/textinput"
	"github.com/squarecloudofc/sdk-api-go/v2/rest"
)

func NewLoginCommand(squareCli cli.SquareCLI) *cobra.Command {
	var token string

	cmd := &cobra.Command{
		Use:         "login",
		Short:       squareCli.I18n().T("metadata.commands.auth.login.short"),
		Annotations: map[string]string{"skipAuthCheck": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if token == "" {
				input := textinput.New(squareCli.I18n().T("commands.auth.login.input.label"))
				input.Placeholder = squareCli.I18n().T("commands.auth.login.input.placeholder")
				input.Hidden = true

				var err error
				if token, err = input.RunPrompt(); err != nil {
					return err
				}
			}

			self, err := squareCli.Rest().SelfUser(rest.WithToken(token))
			if err != nil || self.Name == "" {
				fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.auth.login.error"))
				return nil
			}

			configuration := squareCli.Config()
			configuration.AuthToken = token
			if err := configuration.Save(); err != nil {
				return err
			}

			fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.auth.login.success.0", map[string]any{"User": self.Name}))
			fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.auth.login.success.1"))

			if self.Plan.Name == "free" {
				fmt.Fprintln(squareCli.Out(), ui.TextDanger.Render(squareCli.I18n().T("commands.auth.login.warnings.no_plan", map[string]any{
					"Link": "https://squarecloud.app/pricing",
				})))
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&token, "token", "", "Square Cloud API token")
	return cmd
}
