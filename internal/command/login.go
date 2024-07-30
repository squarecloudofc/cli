package command

import (
	"fmt"

	"github.com/erikgeiser/promptkit/textinput"
	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/rest"
)

func NewLoginCommand(squareCli *cli.SquareCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "login",
		Short:       "Login to Square Cloud",
		Annotations: map[string]string{"skipAuthCheck": "true"},
		RunE:        runLoginCommand(squareCli),
	}

	cmd.Flags().String("token", "", "")
	return cmd
}

func runLoginCommand(squareCli *cli.SquareCli) RunEFunc {
	return func(cmd *cobra.Command, args []string) (err error) {
		var token string
		tkn, err := cmd.Flags().GetString("token")

		if tkn == "" {
			input := textinput.New("Your API Token:")
			input.Placeholder = "Insert your square cloud api token"
			input.Hidden = true
			input.Template = `
		{{- Bold .Prompt }} {{ "\n" }}
		{{- ">" }} {{ .Input }}
		`
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

		restClient := rest.NewClient(token)

		self, err := restClient.SelfUser()
		if err != nil {
			return
		}

		if self == nil || self.User.Name == "" {
			fmt.Fprintf(squareCli.Out(), "No user associated for this Square Cloud Token\n")
			return
		}

		squareCli.Config.AuthToken = token
		squareCli.Config.Save()

		fmt.Fprintf(squareCli.Out(), "Your API Token has successfuly changed! You are now logged in a %s\n", self.User.Name)
		fmt.Fprintln(squareCli.Out(), "\nWith great power comes great responsibility!")
		return
	}
}
