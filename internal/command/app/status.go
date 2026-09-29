package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/cmdutil"
)

func NewStatusCommand(squareCli cli.SquareCLI) *cobra.Command {
	var all bool
	var raw bool

	cmd := &cobra.Command{
		Use:         "status [app id]",
		Short:       squareCli.I18n().T("metadata.commands.app.status.short"),
		Annotations: map[string]string{"json": "true"},
		Args:        cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			api := squareCli.API()

			if all {
				list, err := api.Apps.StatusAll(cmd.Context(), "")
				if err != nil {
					return err
				}

				if cmdutil.WantsJSON(cmd) {
					return cmdutil.PrintJSON(squareCli.Out(), list)
				}

				w := cmdutil.NewTableWriter(squareCli.Out())
				defer w.Flush()

				t := squareCli.I18n()
				fmt.Fprintln(w, strings.Join([]string{t.T("ui.table.app_id"), t.T("ui.table.running"), "CPU", t.T("ui.table.mem")}, " \t "))
				for _, item := range list {
					fmt.Fprintf(w, "%s \t %t \t %s \t %s \t\n", item.ID, item.Running, item.CPU, item.RAM)
				}

				return nil
			}

			appID, err := cmdutil.ResolveAppID(cmd.Context(), squareCli, args)
			if err != nil {
				return err
			}

			if raw {
				data, err := api.Apps.StatusRaw(cmd.Context(), appID)
				if err != nil {
					return err
				}

				return cmdutil.PrintJSON(squareCli.Out(), data)
			}

			data, err := api.Apps.Status(cmd.Context(), appID)
			if err != nil {
				return err
			}

			if cmdutil.WantsJSON(cmd) {
				return cmdutil.PrintJSON(squareCli.Out(), data)
			}

			w := cmdutil.NewTableWriter(squareCli.Out())
			defer w.Flush()

			t := squareCli.I18n()
			fmt.Fprintln(w, strings.Join([]string{t.T("ui.table.app_id"), "CPU %", t.T("ui.table.mem"), t.T("ui.table.disk"), t.T("ui.table.status"), t.T("ui.table.uptime")}, " \t "))
			fmt.Fprintf(w, "%s \t %s \t %s \t %s \t %s \t %s \t\n",
				appID, data.CPU, data.RAM, data.Storage, data.Status, formatUptime(data.Uptime))

			return nil
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, squareCli.I18n().T("metadata.flags.status_all_apps"))
	cmd.Flags().BoolVar(&raw, "raw", false, squareCli.I18n().T("metadata.flags.raw"))
	return cmd
}

// formatUptime renders a startup timestamp (Unix ms) as a compact elapsed
// string, e.g. "3d4h" or "12m".
func formatUptime(startedAt *int64) string {
	if startedAt == nil {
		return "-"
	}

	elapsed := time.Since(time.UnixMilli(*startedAt))
	if elapsed < 0 {
		return "-"
	}

	days := int(elapsed.Hours()) / 24
	hours := int(elapsed.Hours()) % 24
	minutes := int(elapsed.Minutes()) % 60

	switch {
	case days > 0:
		return fmt.Sprintf("%dd%dh", days, hours)
	case hours > 0:
		return fmt.Sprintf("%dh%dm", hours, minutes)
	default:
		return fmt.Sprintf("%dm", minutes)
	}
}
