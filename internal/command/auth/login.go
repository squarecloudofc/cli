package auth

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/ui/textinput"
	"github.com/squarecloudofc/sdk-api-go/rest"
)

func NewLoginCommand(squareCli cli.SquareCLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "login",
		Short:       squareCli.I18n().T("commands.auth.login.metadata.short"),
		Annotations: map[string]string{"skipAuthCheck": "true"},
		RunE:        runLoginCommand(squareCli),
	}

	cmd.Flags().String("token", "", "")
	return cmd
}

func runLoginCommand(squareCli cli.SquareCLI) RunEFunc {
	return func(cmd *cobra.Command, args []string) (err error) {
		var token string
		tkn, err := cmd.Flags().GetString("token")

		if tkn == "" {
			input := textinput.New(squareCli.I18n().T("commands.auth.login.input.label"))
			input.Placeholder = squareCli.I18n().T("commands.auth.login.input.placeholder")
			input.Hidden = true
			input.ResultTemplate = ""

			token, err = input.RunPrompt()
			if err != nil {
				return
			}
		}

		if tkn != "" {
			if err != nil {
				return err
			}

			token = tkn
		}

		self, err := squareCli.Rest().SelfUser(rest.WithToken(token))
		if err != nil || self.Name == "" {
			fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.auth.login.error"))
			return
		}

		configuration := squareCli.Config()
		configuration.AuthToken = token
		configuration.Save()

		fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.auth.login.success.0", map[string]any{"User": self.Name}))
		fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.auth.login.success.1"))
		return
	}
}
