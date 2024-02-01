package app

import (
	"fmt"
	"time"

	"strings"
	"text/tabwriter"

	"github.com/rvflash/elapsed"
	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
)

func NewStatusCommand(squareCli *cli.SquareCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show the status of your application",
		RunE:  runStatusCommand(squareCli),
	}

	return cmd
}

func runStatusCommand(squareCli *cli.SquareCli) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) (err error) {
		var appId string

		if len(args) > 0 {
			appId = args[0]
		}

		if len(args) < 1 {
			id, err := CreateApplicationSelection(squareCli)
			if err != nil {
				return err
			}

			appId = id
		}

		data, err := squareCli.Rest.ApplicationStatus(appId)

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
