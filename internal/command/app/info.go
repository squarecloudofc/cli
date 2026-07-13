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
		Use:   "info [app id]",
		Short: squareCli.I18n().T("metadata.commands.app.info.short"),
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			appID, err := cmdutil.ResolveAppID(squareCli, args)
			if err != nil {
				return err
			}

			app, err := squareCli.Rest().GetApplication(appID)
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

			rows := [][2]string{
				{"ID", app.ID},
				{"Name", app.Name},
				{"Description", app.Description},
				{"Language", app.Language},
				{"Memory", strconv.Itoa(app.RAM) + "MB"},
				{"Cluster", app.Cluster},
				{"Owner", app.Owner},
				{"Domain", value(app.Domain)},
				{"Custom domain", value(app.Custom)},
				{"Created at", app.CreatedAt.Format(time.DateTime)},
			}

			for _, row := range rows {
				if row[1] == "" {
					row[1] = "-"
				}
				fmt.Fprintf(squareCli.Out(), "%-14s %s\n", row[0]+":", row[1])
			}

			return nil
		},
	}
}
