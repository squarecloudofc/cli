package app

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/cmdutil"
)

func NewListCommand(squareCli cli.SquareCLI) *cobra.Command {
	return &cobra.Command{
		Use:         "list",
		Short:       squareCli.I18n().T("metadata.commands.app.list.short"),
		Annotations: map[string]string{"json": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			me, err := squareCli.API().Account.Me(cmd.Context())
			if err != nil {
				return err
			}
			applications := me.Applications

			if cmdutil.WantsJSON(cmd) {
				return cmdutil.PrintJSON(squareCli.Out(), applications)
			}

			if len(applications) < 1 {
				fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.app.list.empty"))
				return nil
			}

			w := cmdutil.NewTableWriter(squareCli.Out())
			defer w.Flush()

			t := squareCli.I18n()
			fmt.Fprintln(w, strings.Join([]string{t.T("ui.table.name"), t.T("ui.table.app_id"), t.T("ui.table.memory"), t.T("ui.table.cluster"), t.T("ui.table.lang"), t.T("ui.table.domain")}, " \t "))

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
