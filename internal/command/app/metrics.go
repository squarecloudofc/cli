package app

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/cmdutil"
	"github.com/squarecloudofc/cli/internal/ui"
	"github.com/squarecloudofc/sdk-api-go/v4"
)

func NewMetricsCommand(squareCli cli.SquareCLI) *cobra.Command {
	var rng string

	cmd := &cobra.Command{
		Use:         "metrics [app id]",
		Short:       squareCli.I18n().T("metadata.commands.app.metrics.short"),
		Annotations: map[string]string{"json": "true"},
		Args:        cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := CheckMetricsRange(squareCli, rng); err != nil {
				return err
			}

			appID, err := cmdutil.ResolveAppID(cmd.Context(), squareCli, args)
			if err != nil {
				return err
			}

			metrics, err := squareCli.API().Apps.Metrics(cmd.Context(), appID, rng)
			if err != nil {
				return err
			}

			return PrintMetrics(squareCli, cmd, metrics, rng)
		},
	}

	AddMetricsRangeFlag(squareCli, cmd, &rng)
	return cmd
}

// AddMetricsRangeFlag adds --range to `app metrics` and `db metrics`. No
// default: "" is the API's 24h, and cobra's "(default ...)" isn't translated.
func AddMetricsRangeFlag(squareCli cli.SquareCLI, cmd *cobra.Command, rng *string) {
	cmd.Flags().StringVar(rng, "range", "", squareCli.I18n().T("metadata.flags.range"))
}

// CheckMetricsRange rejects a --range the API would answer with 400.
func CheckMetricsRange(squareCli cli.SquareCLI, rng string) error {
	if rng != "" && rng != "24h" && rng != "7d" {
		return errors.New(squareCli.I18n().T("commands.app.metrics.invalid_range", map[string]any{"Range": rng}))
	}
	return nil
}

// PrintMetrics renders a metrics time series; shared with `db metrics`. The
// API answers 7d with 24h, without an error, to plans and apps that don't
// qualify: a note on stderr says so.
func PrintMetrics(squareCli cli.SquareCLI, cmd *cobra.Command, metrics []squarecloud.MetricPoint, rng string) error {
	// 7d points carry the hour's peaks; 24h points never do.
	peaks := slices.ContainsFunc(metrics, func(p squarecloud.MetricPoint) bool { return p.CPUMax != nil || p.RAMMax != nil })
	if rng == "7d" && len(metrics) > 0 && !peaks {
		fmt.Fprintln(squareCli.Err(), ui.TextYellow.Render(squareCli.I18n().T("commands.app.metrics.only_24h")))
	}

	if cmdutil.WantsJSON(cmd) {
		return cmdutil.PrintJSON(squareCli.Out(), metrics)
	}

	if len(metrics) == 0 {
		fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.app.metrics.empty"))
		return nil
	}

	w := cmdutil.NewTableWriter(squareCli.Out())
	defer w.Flush()

	t := squareCli.I18n()
	headers := []string{t.T("ui.table.date"), "CPU %", "RAM MB"}
	if peaks {
		headers = append(headers, "CPU MAX %", "RAM MAX MB")
	}
	headers = append(headers, t.T("ui.table.net"), t.T("ui.table.disk_io"), t.T("ui.table.storage"))
	fmt.Fprintln(w, strings.Join(headers, " \t "))

	for _, point := range metrics {
		row := []string{point.Date.Local().Format(time.DateTime), fmt.Sprintf("%.2f", point.CPU), fmt.Sprintf("%.1f", point.RAM)}
		if peaks {
			row = append(row, orDash(point.CPUMax, "%.2f"), orDash(point.RAMMax, "%.1f"))
		}
		storage := "-"
		if point.Storage != nil {
			storage = cmdutil.FormatBytes(*point.Storage)
		}
		row = append(row, bytePair(point.Net), bytePair(point.Disk), storage)
		fmt.Fprintln(w, strings.Join(row, " \t "))
	}

	return nil
}

func orDash(v *float64, format string) string {
	if v == nil {
		return "-"
	}
	return fmt.Sprintf(format, *v)
}

// bytePair renders [in, out] byte counters as "in/out", or "-" without a
// reading.
func bytePair[T int64 | float64](v []T) string {
	if len(v) == 0 {
		return "-"
	}
	s := make([]string, len(v))
	for i, n := range v {
		s[i] = cmdutil.FormatBytes(int64(n))
	}
	return strings.Join(s, "/")
}
