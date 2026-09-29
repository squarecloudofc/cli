package app

import (
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/cmdutil"
)

func NewInfoCommand(squareCli cli.SquareCLI) *cobra.Command {
	return &cobra.Command{
		Use:         "info [app id]",
		Short:       squareCli.I18n().T("metadata.commands.app.info.short"),
		Annotations: map[string]string{"json": "true"},
		Args:        cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			appID, err := cmdutil.ResolveAppID(cmd.Context(), squareCli, args)
			if err != nil {
				return err
			}

			app, err := squareCli.API().Apps.Get(cmd.Context(), appID)
			if err != nil {
				return err
			}

			if cmdutil.WantsJSON(cmd) {
				return cmdutil.PrintJSON(squareCli.Out(), app)
			}

			value := func(s *string) string {
				if s == nil || *s == "" {
					return "-"
				}
				return *s
			}

			t := squareCli.I18n()
			rows := [][2]string{
				{t.T("ui.field.id"), app.ID},
				{t.T("ui.field.name"), app.Name},
				{t.T("ui.field.description"), app.Description},
				{t.T("ui.field.language"), app.Language},
				{t.T("ui.field.memory"), strconv.Itoa(app.RAM) + "MB"},
				{t.T("ui.field.cluster"), app.Cluster},
				{t.T("ui.field.owner"), app.Owner},
				{t.T("ui.field.domain"), value(app.Domain)},
				{t.T("ui.field.custom_domain"), value(app.Custom)},
				{t.T("ui.field.created_at"), app.CreatedAt.Format(time.DateTime)},
			}

			w := cmdutil.NewTableWriter(squareCli.Out())
			for _, row := range rows {
				if row[1] == "" {
					row[1] = "-"
				}
				fmt.Fprintf(w, "%s\t%s\n", row[0], row[1])
			}

			return w.Flush()
		},
	}
}
