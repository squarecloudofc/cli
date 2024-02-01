package app

import (
	"fmt"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/erikgeiser/promptkit/selection"
	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
)

type RunEFunc func(cmd *cobra.Command, args []string) error

func CreateApplicationSelection(squareCli *cli.SquareCli) (string, error) {
	rapps, err := squareCli.Rest.SelfUser()
	if err != nil {
		return "", err
	}

	var apps []string

	for _, app := range rapps.Applications {
		apps = append(apps, fmt.Sprintf("%s (%s)", app.Tag, app.ID))
	}

	sp := selection.New("What application do you want to restart?", apps)
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
		NewDeleteCommand(squareCli),
		NewStartCommand(squareCli),
		NewRestartCommand(squareCli),
		NewStopCommand(squareCli),
		NewStatusCommand(squareCli),
	)

	return cmd
}

func runAppCommand(squareCli *cli.SquareCli) RunEFunc {
	return func(cmd *cobra.Command, args []string) (err error) {
		self, err := squareCli.Rest.SelfUser()
		if err != nil {
			return
		}

		if self == nil || self.User.Tag == "" {
			fmt.Fprintf(squareCli.Out(), "No user associated with current Square Cloud Token\n")
			return
		}

		if len(self.Applications) < 1 {
			fmt.Fprintln(squareCli.Out(), "You does not have any application active")
			return
		}

		w := tabwriter.NewWriter(squareCli.Out(), 0, 0, 2, ' ', tabwriter.TabIndent)
		defer w.Flush()

		tags := []string{"NAME", "App ID", "MEMORY", "CLUSTER", "LANG", "WEBSITE"}
		fmt.Fprintln(w, strings.Join(tags, " \t "))

		for _, app := range self.Applications {
			values := []string{
				app.Tag,
				app.ID,
				strconv.Itoa(app.RAM) + "mb",
				app.Cluster,
				app.Lang,
				strconv.FormatBool(app.IsWebsite),
			}
			fmt.Fprintln(w, strings.Join(values, " \t "))
		}

		return nil
	}
}
