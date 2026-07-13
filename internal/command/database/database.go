package database

import (
	"fmt"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/cmdutil"
	"github.com/squarecloudofc/cli/internal/command/app"
	"github.com/squarecloudofc/cli/internal/ui"
	"github.com/squarecloudofc/sdk-api-go/v2/squarecloud"
)

func newListCommand(squareCli cli.SquareCLI) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: squareCli.I18n().T("metadata.commands.database.list.short"),
		RunE: func(cmd *cobra.Command, args []string) error {
			databases, err := squareCli.Rest().GetDatabases()
			if err != nil {
				return err
			}

			if cmdutil.WantsJSON(cmd) {
				return cmdutil.PrintJSON(squareCli.Out(), databases)
			}

			if len(databases) == 0 {
				fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.database.list.empty"))
				return nil
			}

			w := tabwriter.NewWriter(squareCli.Out(), 0, 0, 2, ' ', tabwriter.TabIndent)
			defer w.Flush()

			fmt.Fprintln(w, strings.Join([]string{"NAME", "DB ID", "TYPE", "MEMORY", "CLUSTER"}, " \t "))
			for _, db := range databases {
				fmt.Fprintf(w, "%s \t %s \t %s \t %s \t %s \t\n",
					db.Name, db.ID, db.Type, strconv.Itoa(db.RAM)+"MB", db.Cluster)
			}

			return nil
		},
	}
}

func newCreateCommand(squareCli cli.SquareCLI) *cobra.Command {
	var options squarecloud.DatabaseCreateOptions

	cmd := &cobra.Command{
		Use:   "create",
		Short: squareCli.I18n().T("metadata.commands.database.create.short"),
		RunE: func(cmd *cobra.Command, args []string) error {
			if options.Name == "" || options.Memory == 0 || options.Type == "" || options.Version == "" {
				return cmd.Help()
			}

			created, err := squareCli.Rest().CreateDatabase(options)
			if err != nil {
				return err
			}

			if cmdutil.WantsJSON(cmd) {
				return cmdutil.PrintJSON(squareCli.Out(), created)
			}

			fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.database.create.success", map[string]any{"Name": created.Name, "Id": created.ID}))
			fmt.Fprintln(squareCli.Out(), ui.TextYellow.SetString(squareCli.I18n().T("commands.database.create.password_warning")))
			fmt.Fprintf(squareCli.Out(), "  PASSWORD: %s\n", created.Password)
			fmt.Fprintf(squareCli.Out(), "  URL: %s\n", created.ConnectionURL)
			return nil
		},
	}

	cmd.Flags().StringVar(&options.Name, "name", "", "Database name (1-32 chars)")
	cmd.Flags().IntVar(&options.Memory, "memory", 0, "Allocated memory in MB")
	cmd.Flags().StringVar((*string)(&options.Type), "type", "", "Engine: mongo, mysql, redis or postgres")
	cmd.Flags().StringVar(&options.Version, "version", "", "Engine version")
	return cmd
}

func newInfoCommand(squareCli cli.SquareCLI) *cobra.Command {
	return &cobra.Command{
		Use:   "info [db id]",
		Short: squareCli.I18n().T("metadata.commands.database.info.short"),
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dbID, err := cmdutil.ResolveDatabaseID(squareCli, args)
			if err != nil {
				return err
			}

			db, err := squareCli.Rest().GetDatabase(dbID)
			if err != nil {
				return err
			}

			if cmdutil.WantsJSON(cmd) {
				return cmdutil.PrintJSON(squareCli.Out(), db)
			}

			rows := [][2]string{
				{"ID", db.ID},
				{"Name", db.Name},
				{"Type", string(db.Type)},
				{"Memory", strconv.Itoa(db.RAM) + "MB"},
				{"Port", strconv.Itoa(db.Port)},
				{"Cluster", db.Cluster},
				{"Owner", db.Owner},
				{"Created at", db.CreatedAt.Format(time.DateTime)},
			}

			for _, row := range rows {
				fmt.Fprintf(squareCli.Out(), "%-12s %s\n", row[0]+":", row[1])
			}

			return nil
		},
	}
}

func newUpdateCommand(squareCli cli.SquareCLI) *cobra.Command {
	var name string
	var memory int

	cmd := &cobra.Command{
		Use:   "update [db id]",
		Short: squareCli.I18n().T("metadata.commands.database.update.short"),
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" && memory == 0 {
				return cmd.Help()
			}

			dbID, err := cmdutil.ResolveDatabaseID(squareCli, args)
			if err != nil {
				return err
			}

			options := squarecloud.DatabaseUpdateOptions{Name: name, RAM: memory}
			if err := squareCli.Rest().UpdateDatabase(dbID, options); err != nil {
				return err
			}

			fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.database.update.success"))
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "New database name")
	cmd.Flags().IntVar(&memory, "memory", 0, "New allocated memory in MB")
	return cmd
}

func newDeleteCommand(squareCli cli.SquareCLI) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete [db id]",
		Short: squareCli.I18n().T("metadata.commands.database.delete.short"),
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dbID, err := cmdutil.ResolveDatabaseID(squareCli, args)
			if err != nil {
				return err
			}

			message := squareCli.I18n().T("commands.database.delete.confirm", map[string]any{"Id": dbID})
			if !cmdutil.Confirm(squareCli, message, yes) {
				return nil
			}

			if err := squareCli.Rest().DeleteDatabase(dbID); err != nil {
				return err
			}

			fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.database.delete.success"))
			return nil
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip the confirmation prompt")
	return cmd
}

func newStartCommand(squareCli cli.SquareCLI) *cobra.Command {
	return &cobra.Command{
		Use:   "start [db id]",
		Short: squareCli.I18n().T("metadata.commands.database.start.short"),
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dbID, err := cmdutil.ResolveDatabaseID(squareCli, args)
			if err != nil {
				return err
			}

			if err := squareCli.Rest().StartDatabase(dbID); err != nil {
				return err
			}

			fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.database.signal.success", map[string]any{"Signal": "START"}))
			return nil
		},
	}
}

func newStopCommand(squareCli cli.SquareCLI) *cobra.Command {
	return &cobra.Command{
		Use:   "stop [db id]",
		Short: squareCli.I18n().T("metadata.commands.database.stop.short"),
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dbID, err := cmdutil.ResolveDatabaseID(squareCli, args)
			if err != nil {
				return err
			}

			if err := squareCli.Rest().StopDatabase(dbID); err != nil {
				return err
			}

			fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.database.signal.success", map[string]any{"Signal": "STOP"}))
			return nil
		},
	}
}

func newStatusCommand(squareCli cli.SquareCLI) *cobra.Command {
	var all bool
	var raw bool

	cmd := &cobra.Command{
		Use:   "status [db id]",
		Short: squareCli.I18n().T("metadata.commands.database.status.short"),
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rest := squareCli.Rest()

			if all {
				list, err := rest.GetDatabaseListStatus()
				if err != nil {
					return err
				}

				if cmdutil.WantsJSON(cmd) {
					return cmdutil.PrintJSON(squareCli.Out(), list)
				}

				w := tabwriter.NewWriter(squareCli.Out(), 0, 0, 2, ' ', tabwriter.TabIndent)
				defer w.Flush()

				fmt.Fprintln(w, strings.Join([]string{"DB ID", "RUNNING", "CPU", "MEM"}, " \t "))
				for _, item := range list {
					fmt.Fprintf(w, "%s \t %t \t %s \t %s \t\n", item.ID, item.Running, item.CPU, item.RAM)
				}

				return nil
			}

			dbID, err := cmdutil.ResolveDatabaseID(squareCli, args)
			if err != nil {
				return err
			}

			if raw {
				data, err := rest.GetDatabaseStatusRaw(dbID)
				if err != nil {
					return err
				}

				return cmdutil.PrintJSON(squareCli.Out(), data)
			}

			data, err := rest.GetDatabaseStatus(dbID)
			if err != nil {
				return err
			}

			if cmdutil.WantsJSON(cmd) {
				return cmdutil.PrintJSON(squareCli.Out(), data)
			}

			w := tabwriter.NewWriter(squareCli.Out(), 0, 0, 2, ' ', tabwriter.TabIndent)
			defer w.Flush()

			fmt.Fprintln(w, strings.Join([]string{"DB ID", "CPU %", "MEM", "DISK", "STATUS"}, " \t "))
			fmt.Fprintf(w, "%s \t %s \t %s \t %s \t %s \t\n", dbID, data.CPU, data.RAM, data.Storage, data.Status)

			return nil
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "Show the status of every database")
	cmd.Flags().BoolVar(&raw, "raw", false, "Print raw numeric stats as JSON")
	return cmd
}

func newMetricsCommand(squareCli cli.SquareCLI) *cobra.Command {
	return &cobra.Command{
		Use:   "metrics [db id]",
		Short: squareCli.I18n().T("metadata.commands.database.metrics.short"),
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dbID, err := cmdutil.ResolveDatabaseID(squareCli, args)
			if err != nil {
				return err
			}

			metrics, err := squareCli.Rest().GetDatabaseMetrics(dbID)
			if err != nil {
				return err
			}

			return app.PrintMetrics(squareCli, cmd, metrics)
		},
	}
}

func newCredentialsCommand(squareCli cli.SquareCLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "credentials",
		Short: squareCli.I18n().T("metadata.commands.database.credentials.root.short"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	certificate := &cobra.Command{
		Use:   "certificate [db id]",
		Short: squareCli.I18n().T("metadata.commands.database.credentials.certificate.short"),
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dbID, err := cmdutil.ResolveDatabaseID(squareCli, args)
			if err != nil {
				return err
			}

			result, err := squareCli.Rest().GetDatabaseCertificate(dbID)
			if err != nil {
				return err
			}

			fmt.Fprintln(squareCli.Out(), result.Certificate)
			return nil
		},
	}

	var yes bool
	reset := &cobra.Command{
		Use:   "reset <password|certificate> [db id]",
		Short: squareCli.I18n().T("metadata.commands.database.credentials.reset.short"),
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			resetType := squarecloud.DatabaseResetType(args[0])
			if resetType != squarecloud.DatabaseResetPassword && resetType != squarecloud.DatabaseResetCertificate {
				return fmt.Errorf("invalid reset type %q, expected password or certificate", args[0])
			}

			dbID, err := cmdutil.ResolveDatabaseID(squareCli, args[1:])
			if err != nil {
				return err
			}

			message := squareCli.I18n().T("commands.database.credentials.reset.confirm", map[string]any{"Type": string(resetType), "Id": dbID})
			if !cmdutil.Confirm(squareCli, message, yes) {
				return nil
			}

			result, err := squareCli.Rest().ResetDatabaseCredentials(dbID, resetType)
			if err != nil {
				return err
			}

			fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.database.credentials.reset.success"))
			if result.Password != "" {
				fmt.Fprintln(squareCli.Out(), ui.TextYellow.SetString(squareCli.I18n().T("commands.database.create.password_warning")))
				fmt.Fprintf(squareCli.Out(), "  PASSWORD: %s\n", result.Password)
			}

			return nil
		},
	}
	reset.Flags().BoolVarP(&yes, "yes", "y", false, "Skip the confirmation prompt")

	cmd.AddCommand(certificate, reset)
	return cmd
}
