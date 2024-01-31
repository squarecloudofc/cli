package command

import (
	"fmt"

	"github.com/erikgeiser/promptkit/textinput"
	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
)

func NewLoginCommand(squareCli *cli.SquareCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login to Square Cloud",
		RunE:  runLoginCommand(squareCli),
	}

	return cmd
}

func runLoginCommand(squareCli *cli.SquareCli) RunEFunc {
	return func(cmd *cobra.Command, args []string) (err error) {

		input := textinput.New("Your API Token:")
		input.Placeholder = "Insert your square cloud api token"
		input.Hidden = true
		input.Template = `
		{{- Bold .Prompt }} {{ "\n" }}
		{{- ">" }} {{ .Input }}
		`
		input.ResultTemplate = ""

		token, err := input.RunPrompt()
		if err != nil {
			return
		}

		squareCli.Config.AuthToken = token
		squareCli.Config.Save()

		fmt.Fprintln(squareCli.Out(), "Your API Token has successfuly changed")
		fmt.Fprintln(squareCli.Out(), "\nWith great power comes great responsibility!")
		return
	}
}
