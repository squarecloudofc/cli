package app

import (
	"fmt"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
)

func NewListCommand(squareCli cli.SquareCLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all your Square Cloud applications",
		RunE:  runAppListCommand(squareCli),
	}

	return cmd
}

func runAppListCommand(squareCli cli.SquareCLI) RunEFunc {
	return func(cmd *cobra.Command, args []string) (err error) {
		rest := squareCli.Rest()
		applications, err := rest.GetApplications()
		if err != nil {
			return
		}

		if len(applications) < 1 {
			fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.app.list.empty"))
			return
		}

		w := tabwriter.NewWriter(squareCli.Out(), 0, 0, 2, ' ', tabwriter.TabIndent)
		defer w.Flush()

		tags := []string{"NAME", "App ID", "MEMORY", "CLUSTER", "LANG"}
		fmt.Fprintln(w, strings.Join(tags, " \t "))

		for _, app := range applications {
			values := []string{
				app.Name,
				app.ID,
				strconv.Itoa(app.RAM) + "mb",
				app.Cluster,
				app.Lang,
			}
			fmt.Fprintln(w, strings.Join(values, " \t "))
		}

		return nil
	}
}
