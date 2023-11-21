package command

import (
	"fmt"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
)

type RunEFunc func(cmd *cobra.Command, args []string) error

func NewAppCommand(squareCli *cli.SquareCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "app",
		Short: "Do some actions with your applications",
		RunE:  runAppCommand(squareCli),
	}

	cmd.AddCommand(
		NewStartCommand(squareCli),
		NewRestartCommand(squareCli),
		NewStopCommand(squareCli),
	)

	return cmd
}

func runAppCommand(squareCli *cli.SquareCli) RunEFunc {
	return func(cmd *cobra.Command, args []string) (err error) {
		user, err := squareCli.Rest.SelfUser()
		if err != nil {
			return
		}
		if len(user.Applications) < 1 {
			fmt.Println("You does not have any application active")
			return
		}

		w := tabwriter.NewWriter(squareCli.Out(), 0, 0, 2, ' ', tabwriter.TabIndent)
		defer w.Flush()

		tags := []string{"NAME", "App ID", "MEMORY", "CLUSTER", "LANG", "WEBSITE"}
		fmt.Fprintln(w, strings.Join(tags, " \t "))

		for _, app := range user.Applications {
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
