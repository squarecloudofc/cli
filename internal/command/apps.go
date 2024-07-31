package command

import (
	"fmt"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
)

func NewAppsCommand(squareCli *cli.SquareCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apps",
		Short: "List all your Square Cloud applications",
		RunE:  runAppsCommand(squareCli),
	}

	cmd.PersistentFlags().StringP("search", "s", "", "Search for an application")
	return cmd
}

func runAppsCommand(squareCli *cli.SquareCli) RunEFunc {
	return func(cmd *cobra.Command, args []string) (err error) {
		rest := squareCli.Rest()
		self, err := rest.SelfUser()
		if err != nil {
			return
		}

		if len(self.Applications) < 1 {
			fmt.Fprintln(squareCli.Out(), "You don't have any application active")
			return
		}

		w := tabwriter.NewWriter(squareCli.Out(), 0, 0, 2, ' ', tabwriter.TabIndent)
		defer w.Flush()

		tags := []string{"NAME", "App ID", "MEMORY", "CLUSTER", "LANG", "WEBSITE"}
		fmt.Fprintln(w, strings.Join(tags, " \t "))

		for _, app := range self.Applications {
			values := []string{
				app.Name,
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
