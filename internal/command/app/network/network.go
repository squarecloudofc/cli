// Package network implements `app network` — DNS, custom domains and edge
// observability (analytics, errors, logs, performance, cache purge).
package network

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/cmdutil"
	"github.com/squarecloudofc/cli/internal/ui"
	"github.com/squarecloudofc/sdk-api-go/v3"
)

func NewCommand(squareCli cli.SquareCLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "network",
		Short: squareCli.I18n().T("metadata.commands.app.network.root.short"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(
		newDNSCommand(squareCli),
		newDomainCommand(squareCli),
		newAnalyticsCommand(squareCli),
		newErrorsCommand(squareCli),
		newLogsCommand(squareCli),
		newPerformanceCommand(squareCli),
		newPurgeCacheCommand(squareCli),
	)

	return cmd
}

// rangeFlags registers --start/--end and returns a resolver that defaults to
// the last 24 hours (API max retention is 7 days).
func rangeFlags(cmd *cobra.Command, squareCli cli.SquareCLI) func() (time.Time, time.Time, error) {
	var startFlag, endFlag string

	cmd.Flags().StringVar(&startFlag, "start", "", squareCli.I18n().T("metadata.flags.start"))
	cmd.Flags().StringVar(&endFlag, "end", "", squareCli.I18n().T("metadata.flags.end"))

	return func() (time.Time, time.Time, error) {
		end := time.Now()
		start := end.Add(-24 * time.Hour)

		var err error
		if startFlag != "" {
			if start, err = time.Parse(time.RFC3339, startFlag); err != nil {
				return start, end, errors.New(squareCli.I18n().T("commands.app.network.invalid_time", map[string]any{"Flag": "--start", "Value": startFlag}))
			}
		}
		if endFlag != "" {
			if end, err = time.Parse(time.RFC3339, endFlag); err != nil {
				return start, end, errors.New(squareCli.I18n().T("commands.app.network.invalid_time", map[string]any{"Flag": "--end", "Value": endFlag}))
			}
		}

		return start, end, nil
	}
}

func newDNSCommand(squareCli cli.SquareCLI) *cobra.Command {
	return &cobra.Command{
		Use:         "dns [app id]",
		Short:       squareCli.I18n().T("metadata.commands.app.network.dns.short"),
		Annotations: map[string]string{"json": "true"},
		Args:        cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			appID, err := cmdutil.ResolveAppID(cmd.Context(), squareCli, args)
			if err != nil {
				return err
			}

			records, err := squareCli.API().Apps.Network.DNS(cmd.Context(), appID)
			if err != nil {
				return err
			}

			if cmdutil.WantsJSON(cmd) {
				return cmdutil.PrintJSON(squareCli.Out(), records)
			}

			if len(records) == 0 {
				fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.app.network.dns.empty"))
				return nil
			}

			w := cmdutil.NewTableWriter(squareCli.Out())
			defer w.Flush()

			t := squareCli.I18n()
			fmt.Fprintln(w, strings.Join([]string{t.T("ui.table.type"), t.T("ui.table.name"), t.T("ui.table.value"), t.T("ui.table.status")}, " \t "))
			for _, record := range records {
				fmt.Fprintf(w, "%s \t %s \t %s \t %s \t\n", record.Type, record.Name, record.Value, record.Status)
			}

			return nil
		},
	}
}

func newDomainCommand(squareCli cli.SquareCLI) *cobra.Command {
	var appID string

	cmd := &cobra.Command{
		Use:   "domain <hostname>",
		Short: squareCli.I18n().T("metadata.commands.app.network.domain.short"),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var idArgs []string
			if appID != "" {
				idArgs = []string{appID}
			}

			id, err := cmdutil.ResolveAppID(cmd.Context(), squareCli, idArgs)
			if err != nil {
				return err
			}

			if err := squareCli.API().Apps.Network.SetDomain(cmd.Context(), id, args[0]); err != nil {
				return err
			}

			fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.app.network.domain.success", map[string]any{"Domain": args[0]}))
			fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.app.network.domain.hint"))
			return nil
		},
	}

	cmd.Flags().StringVar(&appID, "app", "", squareCli.I18n().T("metadata.flags.app"))
	return cmd
}

func newAnalyticsCommand(squareCli cli.SquareCLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "analytics [app id]",
		Short:       squareCli.I18n().T("metadata.commands.app.network.analytics.short"),
		Annotations: map[string]string{"json": "true"},
		Args:        cobra.MaximumNArgs(1),
	}

	window := rangeFlags(cmd, squareCli)

	var filters squarecloud.AnalyticsFilters
	for name, value := range map[string]*string{
		"country": &filters.Country, "ip": &filters.IP, "path": &filters.Path, "status": &filters.Status,
		"os": &filters.OS, "browser": &filters.Browser, "protocol": &filters.Protocol, "referer": &filters.Referer,
		"provider": &filters.Provider, "content-type": &filters.ContentType, "bot": &filters.Bot,
	} {
		cmd.Flags().StringVar(value, name, "", squareCli.I18n().T("metadata.flags.filter", map[string]any{"Name": name}))
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		appID, err := cmdutil.ResolveAppID(cmd.Context(), squareCli, args)
		if err != nil {
			return err
		}

		start, end, err := window()
		if err != nil {
			return err
		}

		analytics, err := squareCli.API().Apps.Network.Analytics(cmd.Context(), appID, start, end, filters)
		if err != nil {
			return err
		}

		if cmdutil.WantsJSON(cmd) {
			return cmdutil.PrintJSON(squareCli.Out(), analytics)
		}

		if analytics == nil || len(analytics.Visits) == 0 {
			fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.app.network.analytics.empty"))
			return nil
		}

		var visits, requests, bytes int64
		for _, bucket := range analytics.Visits {
			visits += bucket.Visits
			requests += bucket.Requests
			bytes += bucket.Bytes
		}

		fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.app.network.analytics.summary", map[string]any{
			"Visits":   visits,
			"Requests": requests,
			"Traffic":  cmdutil.FormatBytes(bytes),
		}))

		w := cmdutil.NewTableWriter(squareCli.Out())
		defer w.Flush()

		t := squareCli.I18n()
		fmt.Fprintln(w, strings.Join([]string{t.T("ui.table.country"), t.T("ui.table.requests"), t.T("ui.table.visits")}, " \t "))
		for i, bucket := range analytics.Countries {
			if i >= 10 {
				break
			}
			fmt.Fprintf(w, "%s \t %d \t %d \t\n", bucket.Type, bucket.Requests, bucket.Visits)
		}

		return nil
	}

	return cmd
}

func newErrorsCommand(squareCli cli.SquareCLI) *cobra.Command {
	var include4xx bool

	cmd := &cobra.Command{
		Use:         "errors [app id]",
		Short:       squareCli.I18n().T("metadata.commands.app.network.errors.short"),
		Annotations: map[string]string{"json": "true"},
		Args:        cobra.MaximumNArgs(1),
	}

	window := rangeFlags(cmd, squareCli)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		appID, err := cmdutil.ResolveAppID(cmd.Context(), squareCli, args)
		if err != nil {
			return err
		}

		start, end, err := window()
		if err != nil {
			return err
		}

		networkErrors, err := squareCli.API().Apps.Network.Errors(cmd.Context(), appID, start, end, include4xx)
		if err != nil {
			return err
		}

		if cmdutil.WantsJSON(cmd) {
			return cmdutil.PrintJSON(squareCli.Out(), networkErrors)
		}

		if networkErrors == nil { // no traffic in the window
			networkErrors = &squarecloud.NetworkErrors{}
		}

		fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.app.network.errors.summary", map[string]any{
			"Total":    networkErrors.Summary.Total,
			"Class4xx": networkErrors.Summary.ByClass["4xx"],
			"Class5xx": networkErrors.Summary.ByClass["5xx"],
		}))

		if len(networkErrors.TopPaths) == 0 {
			return nil
		}

		w := cmdutil.NewTableWriter(squareCli.Out())
		defer w.Flush()

		t := squareCli.I18n()
		fmt.Fprintln(w, strings.Join([]string{t.T("ui.table.path"), t.T("ui.table.method"), t.T("ui.table.errors")}, " \t "))
		for i, path := range networkErrors.TopPaths {
			if i >= 10 {
				break
			}

			method := "-"
			if path.Method != nil {
				method = *path.Method
			}

			fmt.Fprintf(w, "%s \t %s \t %d \t\n", path.Path, method, path.Total)
		}

		return nil
	}

	cmd.Flags().BoolVar(&include4xx, "include-4xx", false, squareCli.I18n().T("metadata.flags.include_4xx"))
	return cmd
}

func newLogsCommand(squareCli cli.SquareCLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "logs [app id]",
		Short:       squareCli.I18n().T("metadata.commands.app.network.logs.short"),
		Annotations: map[string]string{"json": "true"},
		Args:        cobra.MaximumNArgs(1),
	}

	window := rangeFlags(cmd, squareCli)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		appID, err := cmdutil.ResolveAppID(cmd.Context(), squareCli, args)
		if err != nil {
			return err
		}

		start, end, err := window()
		if err != nil {
			return err
		}

		logs, err := squareCli.API().Apps.Network.Logs(cmd.Context(), appID, start, end)
		if err != nil {
			return err
		}

		if cmdutil.WantsJSON(cmd) {
			return cmdutil.PrintJSON(squareCli.Out(), logs)
		}

		if len(logs) == 0 {
			fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.app.network.logs.empty"))
			return nil
		}

		w := cmdutil.NewTableWriter(squareCli.Out())
		defer w.Flush()

		t := squareCli.I18n()
		fmt.Fprintln(w, strings.Join([]string{t.T("ui.table.time"), t.T("ui.table.status"), t.T("ui.table.method"), t.T("ui.table.path"), "IP", t.T("ui.table.country")}, " \t "))
		for _, log := range logs {
			ip, country := "-", "-"
			if log.Client.IP != nil {
				ip = *log.Client.IP
			}
			if log.Client.Country != nil {
				country = *log.Client.Country
			}

			fmt.Fprintf(w, "%s \t %d \t %s \t %s \t %s \t %s \t\n",
				log.Timestamp.Local().Format(time.TimeOnly), log.Response.Status,
				log.Request.Method, log.Request.Path, ip, country)
		}

		return nil
	}

	return cmd
}

func newPerformanceCommand(squareCli cli.SquareCLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "performance [app id]",
		Short:       squareCli.I18n().T("metadata.commands.app.network.performance.short"),
		Annotations: map[string]string{"json": "true"},
		Args:        cobra.MaximumNArgs(1),
	}

	window := rangeFlags(cmd, squareCli)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		appID, err := cmdutil.ResolveAppID(cmd.Context(), squareCli, args)
		if err != nil {
			return err
		}

		start, end, err := window()
		if err != nil {
			return err
		}

		performance, err := squareCli.API().Apps.Network.Performance(cmd.Context(), appID, start, end)
		if err != nil {
			return err
		}

		if cmdutil.WantsJSON(cmd) {
			return cmdutil.PrintJSON(squareCli.Out(), performance)
		}

		if performance == nil { // no traffic in the window
			performance = &squarecloud.NetworkPerformance{}
		}

		fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.app.network.performance.summary", map[string]any{
			"Requests": performance.Summary.Requests,
			"EdgeP50":  ms(performance.Summary.Edge.P50),
			"EdgeP95":  ms(performance.Summary.Edge.P95),
			"OrigP50":  ms(performance.Summary.Origin.P50),
			"OrigP95":  ms(performance.Summary.Origin.P95),
		}))

		if len(performance.SlowestPaths) == 0 {
			return nil
		}

		w := cmdutil.NewTableWriter(squareCli.Out())
		defer w.Flush()

		t := squareCli.I18n()
		fmt.Fprintln(w, strings.Join([]string{t.T("ui.table.slowest_path"), "P95 MS", "P99 MS", t.T("ui.table.requests")}, " \t "))
		for i, path := range performance.SlowestPaths {
			if i >= 10 {
				break
			}
			fmt.Fprintf(w, "%s \t %s \t %s \t %d \t\n", path.Path, ms(path.P95), ms(path.P99), path.Requests)
		}

		return nil
	}

	return cmd
}

func newPurgeCacheCommand(squareCli cli.SquareCLI) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "purge-cache [app id]",
		Short: squareCli.I18n().T("metadata.commands.app.network.purge_cache.short"),
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			appID, err := cmdutil.ResolveAppID(cmd.Context(), squareCli, args)
			if err != nil {
				return err
			}

			message := squareCli.I18n().T("commands.app.network.purge_cache.confirm", map[string]any{"Appid": appID})
			if err := cmdutil.Confirm(squareCli, message, yes); err != nil {
				return err
			}

			if err := squareCli.API().Apps.Network.PurgeCache(cmd.Context(), appID); err != nil {
				return err
			}

			fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.app.network.purge_cache.success"))
			return nil
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, squareCli.I18n().T("metadata.flags.yes"))
	return cmd
}

// ms renders a latency the API may send as null.
func ms(v *float64) string {
	if v == nil {
		return "-"
	}
	return fmt.Sprintf("%.0f", *v)
}
