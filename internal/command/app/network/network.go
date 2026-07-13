// Package network implements `app network` — DNS, custom domains and edge
// observability (analytics, errors, logs, performance, cache purge).
package network

import (
	"fmt"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/cmdutil"
	"github.com/squarecloudofc/cli/internal/ui"
	"github.com/squarecloudofc/sdk-api-go/v2/rest"
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
func rangeFlags(cmd *cobra.Command) func() (time.Time, time.Time, error) {
	var startFlag, endFlag string

	cmd.Flags().StringVar(&startFlag, "start", "", "Start of the window (RFC 3339, default: 24h ago)")
	cmd.Flags().StringVar(&endFlag, "end", "", "End of the window (RFC 3339, default: now)")

	return func() (time.Time, time.Time, error) {
		end := time.Now()
		start := end.Add(-24 * time.Hour)

		var err error
		if startFlag != "" {
			if start, err = time.Parse(time.RFC3339, startFlag); err != nil {
				return start, end, fmt.Errorf("invalid --start: %w", err)
			}
		}
		if endFlag != "" {
			if end, err = time.Parse(time.RFC3339, endFlag); err != nil {
				return start, end, fmt.Errorf("invalid --end: %w", err)
			}
		}

		return start, end, nil
	}
}

func newDNSCommand(squareCli cli.SquareCLI) *cobra.Command {
	return &cobra.Command{
		Use:   "dns [app id]",
		Short: squareCli.I18n().T("metadata.commands.app.network.dns.short"),
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			appID, err := cmdutil.ResolveAppID(squareCli, args)
			if err != nil {
				return err
			}

			records, err := squareCli.Rest().GetApplicationDNS(appID)
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

			w := tabwriter.NewWriter(squareCli.Out(), 0, 0, 2, ' ', tabwriter.TabIndent)
			defer w.Flush()

			fmt.Fprintln(w, strings.Join([]string{"TYPE", "NAME", "VALUE", "STATUS"}, " \t "))
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

			id, err := cmdutil.ResolveAppID(squareCli, idArgs)
			if err != nil {
				return err
			}

			if err := squareCli.Rest().SetApplicationCustomDomain(id, args[0]); err != nil {
				return err
			}

			fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.app.network.domain.success", map[string]any{"Domain": args[0]}))
			fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.app.network.domain.hint"))
			return nil
		},
	}

	cmd.Flags().StringVar(&appID, "app", "", "Application ID")
	return cmd
}

func newAnalyticsCommand(squareCli cli.SquareCLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "analytics [app id]",
		Short: squareCli.I18n().T("metadata.commands.app.network.analytics.short"),
		Args:  cobra.MaximumNArgs(1),
	}

	window := rangeFlags(cmd)

	filters := map[string]*string{}
	for _, name := range []string{"country", "ip", "path", "status", "os", "browser", "protocol", "referer", "provider", "content-type", "bot"} {
		value := new(string)
		filters[name] = value
		cmd.Flags().StringVar(value, name, "", "Filter analytics by "+name)
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		appID, err := cmdutil.ResolveAppID(squareCli, args)
		if err != nil {
			return err
		}

		start, end, err := window()
		if err != nil {
			return err
		}

		var opts []rest.RequestOpt
		for name, value := range filters {
			if *value != "" {
				opts = append(opts, rest.WithQueryParam(strings.ReplaceAll(name, "-", "_"), *value))
			}
		}

		analytics, err := squareCli.Rest().GetApplicationAnalytics(appID, start, end, opts...)
		if err != nil {
			return err
		}

		if cmdutil.WantsJSON(cmd) {
			return cmdutil.PrintJSON(squareCli.Out(), analytics)
		}

		if len(analytics.Visits) == 0 {
			fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.app.network.analytics.empty"))
			return nil
		}

		var visits, requests, bytes int64
		for _, bucket := range analytics.Visits {
			visits += int64(bucket.Visits)
			requests += int64(bucket.Requests)
			bytes += bucket.Bytes
		}

		fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.app.network.analytics.summary", map[string]any{
			"Visits":   visits,
			"Requests": requests,
			"Traffic":  cmdutil.FormatBytes(bytes),
		}))

		w := tabwriter.NewWriter(squareCli.Out(), 0, 0, 2, ' ', tabwriter.TabIndent)
		defer w.Flush()

		fmt.Fprintln(w, strings.Join([]string{"COUNTRY", "REQUESTS", "VISITS"}, " \t "))
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
		Use:   "errors [app id]",
		Short: squareCli.I18n().T("metadata.commands.app.network.errors.short"),
		Args:  cobra.MaximumNArgs(1),
	}

	window := rangeFlags(cmd)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		appID, err := cmdutil.ResolveAppID(squareCli, args)
		if err != nil {
			return err
		}

		start, end, err := window()
		if err != nil {
			return err
		}

		var opts []rest.RequestOpt
		if include4xx {
			opts = append(opts, rest.WithQueryParam("include_4xx", "true"))
		}

		networkErrors, err := squareCli.Rest().GetApplicationNetworkErrors(appID, start, end, opts...)
		if err != nil {
			return err
		}

		if cmdutil.WantsJSON(cmd) {
			return cmdutil.PrintJSON(squareCli.Out(), networkErrors)
		}

		fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.app.network.errors.summary", map[string]any{
			"Total":    networkErrors.Summary.Total,
			"Class4xx": networkErrors.Summary.ByClass.Class4xx,
			"Class5xx": networkErrors.Summary.ByClass.Class5xx,
		}))

		if len(networkErrors.TopPaths) == 0 {
			return nil
		}

		w := tabwriter.NewWriter(squareCli.Out(), 0, 0, 2, ' ', tabwriter.TabIndent)
		defer w.Flush()

		fmt.Fprintln(w, strings.Join([]string{"PATH", "METHOD", "ERRORS"}, " \t "))
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

	cmd.Flags().BoolVar(&include4xx, "include-4xx", false, "Include 4xx responses (default: 5xx only)")
	return cmd
}

func newLogsCommand(squareCli cli.SquareCLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs [app id]",
		Short: squareCli.I18n().T("metadata.commands.app.network.logs.short"),
		Args:  cobra.MaximumNArgs(1),
	}

	window := rangeFlags(cmd)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		appID, err := cmdutil.ResolveAppID(squareCli, args)
		if err != nil {
			return err
		}

		start, end, err := window()
		if err != nil {
			return err
		}

		logs, err := squareCli.Rest().GetApplicationNetworkLogs(appID, start, end)
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

		w := tabwriter.NewWriter(squareCli.Out(), 0, 0, 2, ' ', tabwriter.TabIndent)
		defer w.Flush()

		fmt.Fprintln(w, strings.Join([]string{"TIME", "STATUS", "METHOD", "PATH", "IP", "COUNTRY"}, " \t "))
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
		Use:   "performance [app id]",
		Short: squareCli.I18n().T("metadata.commands.app.network.performance.short"),
		Args:  cobra.MaximumNArgs(1),
	}

	window := rangeFlags(cmd)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		appID, err := cmdutil.ResolveAppID(squareCli, args)
		if err != nil {
			return err
		}

		start, end, err := window()
		if err != nil {
			return err
		}

		performance, err := squareCli.Rest().GetApplicationNetworkPerformance(appID, start, end)
		if err != nil {
			return err
		}

		if cmdutil.WantsJSON(cmd) {
			return cmdutil.PrintJSON(squareCli.Out(), performance)
		}

		fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.app.network.performance.summary", map[string]any{
			"Requests": performance.Summary.Requests,
			"EdgeP50":  performance.Summary.Edge.P50,
			"EdgeP95":  performance.Summary.Edge.P95,
			"OrigP50":  performance.Summary.Origin.P50,
			"OrigP95":  performance.Summary.Origin.P95,
		}))

		if len(performance.SlowestPaths) == 0 {
			return nil
		}

		w := tabwriter.NewWriter(squareCli.Out(), 0, 0, 2, ' ', tabwriter.TabIndent)
		defer w.Flush()

		fmt.Fprintln(w, strings.Join([]string{"SLOWEST PATH", "P95 MS", "P99 MS", "REQUESTS"}, " \t "))
		for i, path := range performance.SlowestPaths {
			if i >= 10 {
				break
			}
			fmt.Fprintf(w, "%s \t %.0f \t %.0f \t %d \t\n", path.Path, path.P95, path.P99, path.Requests)
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
			appID, err := cmdutil.ResolveAppID(squareCli, args)
			if err != nil {
				return err
			}

			message := squareCli.I18n().T("commands.app.network.purge_cache.confirm", map[string]any{"Appid": appID})
			if !cmdutil.Confirm(squareCli, message, yes) {
				return nil
			}

			if err := squareCli.Rest().PurgeApplicationCache(appID); err != nil {
				return err
			}

			fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.app.network.purge_cache.success"))
			return nil
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip the confirmation prompt")
	return cmd
}
