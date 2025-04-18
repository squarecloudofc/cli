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
		Short:       "Login to Square Cloud",
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
			input := textinput.New("Your API Token:")
			input.Placeholder = "Insert your square cloud api token"
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
			fmt.Fprintf(squareCli.Out(), "No user associated for this Square Cloud Token\n")
			return
		}

		configuration := squareCli.Config()
		configuration.AuthToken = token
		configuration.Save()

		fmt.Fprintf(squareCli.Out(), "Your API Token has successfuly changed! You are now logged in a %s\n", self.Name)
		fmt.Fprintln(squareCli.Out(), "\nWith great power comes great responsibility!")
		return
	}
}
