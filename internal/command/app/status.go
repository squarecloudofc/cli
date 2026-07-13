package app

import (
	"fmt"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/cmdutil"
)

func NewStatusCommand(squareCli cli.SquareCLI) *cobra.Command {
	var all bool
	var raw bool

	cmd := &cobra.Command{
		Use:   "status [app id]",
		Short: squareCli.I18n().T("metadata.commands.app.status.short"),
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rest := squareCli.Rest()

			if all {
				list, err := rest.GetApplicationListStatus()
				if err != nil {
					return err
				}

				if cmdutil.WantsJSON(cmd) {
					return cmdutil.PrintJSON(squareCli.Out(), list)
				}

				w := tabwriter.NewWriter(squareCli.Out(), 0, 0, 2, ' ', tabwriter.TabIndent)
				defer w.Flush()

				fmt.Fprintln(w, strings.Join([]string{"APP ID", "RUNNING", "CPU", "MEM"}, " \t "))
				for _, item := range list {
					fmt.Fprintf(w, "%s \t %t \t %s \t %s \t\n", item.ID, item.Running, item.CPU, item.RAM)
				}

				return nil
			}

			appID, err := cmdutil.ResolveAppID(squareCli, args)
			if err != nil {
				return err
			}

			if raw {
				data, err := rest.GetApplicationStatusRaw(appID)
				if err != nil {
					return err
				}

				return cmdutil.PrintJSON(squareCli.Out(), data)
			}

			data, err := rest.GetApplicationStatus(appID)
			if err != nil {
				return err
			}

			if cmdutil.WantsJSON(cmd) {
				return cmdutil.PrintJSON(squareCli.Out(), data)
			}

			w := tabwriter.NewWriter(squareCli.Out(), 0, 0, 2, ' ', tabwriter.TabIndent)
			defer w.Flush()

			fmt.Fprintln(w, strings.Join([]string{"APP ID", "CPU %", "MEM", "DISK", "STATUS", "UPTIME"}, " \t "))
			fmt.Fprintf(w, "%s \t %s \t %s \t %s \t %s \t %s \t\n",
				appID, data.CPU, data.RAM, data.Storage, data.Status, formatUptime(data.Uptime))

			return nil
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "Show the status of every application")
	cmd.Flags().BoolVar(&raw, "raw", false, "Print raw numeric stats as JSON")
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
