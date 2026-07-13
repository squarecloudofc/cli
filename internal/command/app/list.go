package app

import (
	"fmt"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/cmdutil"
)

func NewListCommand(squareCli cli.SquareCLI) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: squareCli.I18n().T("metadata.commands.app.list.short"),
		RunE: func(cmd *cobra.Command, args []string) error {
			applications, err := squareCli.Rest().GetApplications()
			if err != nil {
				return err
			}

			if cmdutil.WantsJSON(cmd) {
				return cmdutil.PrintJSON(squareCli.Out(), applications)
			}

			if len(applications) < 1 {
				fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.app.list.empty"))
				return nil
			}

			w := tabwriter.NewWriter(squareCli.Out(), 0, 0, 2, ' ', tabwriter.TabIndent)
			defer w.Flush()

			fmt.Fprintln(w, strings.Join([]string{"NAME", "APP ID", "MEMORY", "CLUSTER", "LANG", "DOMAIN"}, " \t "))

			for _, app := range applications {
				domain := "-"
				if app.Custom != nil && *app.Custom != "" {
					domain = *app.Custom
				} else if app.Domain != nil && *app.Domain != "" {
					domain = *app.Domain
				}

				values := []string{
					app.Name,
					app.ID,
					strconv.Itoa(app.RAM) + "MB",
					app.Cluster,
					app.Lang,
					domain,
				}
				fmt.Fprintln(w, strings.Join(values, " \t "))
			}

			return nil
		},
	}
}
