package app

import (
	"fmt"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/rvflash/elapsed"
	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/ui/application_selector"
)

func NewStatusCommand(squareCli cli.SquareCLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: squareCli.I18n().T("metadata.commands.app.status.short"),
		RunE:  runStatusCommand(squareCli),
	}

	return cmd
}

func runStatusCommand(squareCli cli.SquareCLI) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) (err error) {
		var appId string
		rest := squareCli.Rest()

		if len(args) > 0 {
			appId = args[0]
		}

		if len(args) < 1 {
			m, err := application_selector.RunSelector(squareCli)
			if err != nil {
				return err
			}

			appId = m.ID
		}

		data, err := rest.GetApplicationStatus(appId)
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(squareCli.Out(), 0, 0, 2, ' ', tabwriter.TabIndent)
		defer w.Flush()

		uptime_elapsed := ""

		if data.Status == "running" {
			uptime_elapsed = elapsed.Time(time.Unix(0, (data.Uptime * int64(time.Millisecond))))
		}

		tags := []string{"APP ID", "CPU %", "MEM", "DISK", "STATUS", "UPTIME"}
		fmt.Fprintln(w, strings.Join(tags, " \t "))

		fmt.Fprintf(w, "%s \t %s \t %s \t %s \t %s \t %s \t\n", appId, data.CPU, data.RAM, data.Storage, data.Status, uptime_elapsed)

		return nil
	}
}
