package app

import (
	"fmt"
	"strings"

	"github.com/erikgeiser/promptkit/selection"
	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
)

type RunEFunc func(cmd *cobra.Command, args []string) error

func CreateApplicationSelection(squareCli *cli.SquareCli) (string, error) {
	rest := squareCli.Rest()
	rapps, err := rest.SelfUser()
	if err != nil {
		return "", err
	}

	var apps []string

	for _, app := range rapps.Applications {
		apps = append(apps, fmt.Sprintf("%s (%s)", app.Tag, app.ID))
	}

	sp := selection.New("Which application do you want to use for this action?", apps)
	choice, err := sp.RunPrompt()
	if err != nil {
		return "", err
	}

	id := strings.TrimSuffix(strings.Split(choice, "(")[1], ")")

	return id, nil
}

func NewAppCommand(squareCli *cli.SquareCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "app",
		Short: "Do some actions with your applications",
		RunE:  runAppCommand(squareCli),
	}

	cmd.AddCommand(
		NewBackupCommand(squareCli),
		NewDeleteCommand(squareCli),
		NewLogsCommand(squareCli),
		NewStartCommand(squareCli),
		NewRestartCommand(squareCli),
		NewStopCommand(squareCli),
		NewStatusCommand(squareCli),
	)

	return cmd
}

func runAppCommand(squareCli *cli.SquareCli) RunEFunc {
	return func(cmd *cobra.Command, args []string) (err error) {
		cmd.Help()

		return nil
	}
}
