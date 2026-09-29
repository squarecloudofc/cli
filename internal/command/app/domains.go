package app

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/cmdutil"
)

func NewDomainsCommand(squareCli cli.SquareCLI) *cobra.Command {
	return &cobra.Command{
		Use:         "domains",
		Short:       squareCli.I18n().T("metadata.commands.app.domains.short"),
		Annotations: map[string]string{"json": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			domains, err := squareCli.API().Apps.Domains(cmd.Context())
			if err != nil {
				return err
			}

			if cmdutil.WantsJSON(cmd) {
				return cmdutil.PrintJSON(squareCli.Out(), domains)
			}

			if len(domains) == 0 {
				fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.app.domains.empty"))
				return nil
			}

			w := cmdutil.NewTableWriter(squareCli.Out())
			defer w.Flush()

			t := squareCli.I18n()
			fmt.Fprintln(w, strings.Join([]string{t.T("ui.table.hostname"), t.T("ui.table.type"), t.T("ui.table.app_id")}, " \t "))
			for _, domain := range domains {
				fmt.Fprintf(w, "%s \t %s \t %s \t\n", domain.Hostname, domain.Type, domain.AppID)
			}

			return nil
		},
	}
}

func NewLoadBalancersCommand(squareCli cli.SquareCLI) *cobra.Command {
	return &cobra.Command{
		Use:         "load-balancers",
		Aliases:     []string{"lb"},
		Short:       squareCli.I18n().T("metadata.commands.app.load_balancers.short"),
		Annotations: map[string]string{"json": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			lbs, err := squareCli.API().Apps.LoadBalancers(cmd.Context())
			if err != nil {
				return err
			}

			if cmdutil.WantsJSON(cmd) {
				return cmdutil.PrintJSON(squareCli.Out(), lbs)
			}

			if len(lbs.Balancers) == 0 {
				fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.app.load_balancers.empty"))
				return nil
			}

			fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.app.load_balancers.limit", map[string]any{"Limit": lbs.Limit}))

			w := cmdutil.NewTableWriter(squareCli.Out())
			defer w.Flush()

			t := squareCli.I18n()
			fmt.Fprintln(w, strings.Join([]string{t.T("ui.table.hostname"), t.T("ui.table.app_id"), t.T("ui.table.app_name"), t.T("ui.table.cluster")}, " \t "))
			for _, balancer := range lbs.Balancers {
				for _, app := range balancer.Apps {
					fmt.Fprintf(w, "%s \t %s \t %s \t %s \t\n", balancer.Hostname, app.ID, app.Name, app.Cluster)
				}
			}

			return nil
		},
	}
}
