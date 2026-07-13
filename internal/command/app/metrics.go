package app

import (
	"fmt"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/cmdutil"
	"github.com/squarecloudofc/sdk-api-go/v2/squarecloud"
)

func NewMetricsCommand(squareCli cli.SquareCLI) *cobra.Command {
	return &cobra.Command{
		Use:   "metrics [app id]",
		Short: squareCli.I18n().T("metadata.commands.app.metrics.short"),
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			appID, err := cmdutil.ResolveAppID(squareCli, args)
			if err != nil {
				return err
			}

			metrics, err := squareCli.Rest().GetApplicationMetrics(appID)
			if err != nil {
				return err
			}

			return PrintMetrics(squareCli, cmd, metrics)
		},
	}
}

// PrintMetrics renders a metrics time series; shared with `db metrics`.
func PrintMetrics(squareCli cli.SquareCLI, cmd *cobra.Command, metrics []squarecloud.MetricPoint) error {
	if cmdutil.WantsJSON(cmd) {
		return cmdutil.PrintJSON(squareCli.Out(), metrics)
	}

	if len(metrics) == 0 {
		fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.app.metrics.empty"))
		return nil
	}

	w := tabwriter.NewWriter(squareCli.Out(), 0, 0, 2, ' ', tabwriter.TabIndent)
	defer w.Flush()

	fmt.Fprintln(w, strings.Join([]string{"DATE", "CPU %", "RAM MB", "NET"}, " \t "))
	for _, point := range metrics {
		net := make([]string, len(point.Net))
		for i, n := range point.Net {
			net[i] = fmt.Sprintf("%.0f", n)
		}

		fmt.Fprintf(w, "%s \t %.2f \t %.1f \t %s \t\n",
			point.Date.Local().Format(time.DateTime), point.CPU, point.RAM, strings.Join(net, "/"))
	}

	return nil
}
