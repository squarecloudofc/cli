package database

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/cmdutil"
	"github.com/squarecloudofc/cli/internal/command/app"
	"github.com/squarecloudofc/cli/internal/ui"
	"github.com/squarecloudofc/sdk-api-go/v3"
)

func newListCommand(squareCli cli.SquareCLI) *cobra.Command {
	return &cobra.Command{
		Use:         "list",
		Short:       squareCli.I18n().T("metadata.commands.database.list.short"),
		Annotations: map[string]string{"json": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			me, err := squareCli.API().Account.Me(cmd.Context())
			if err != nil {
				return err
			}
			databases := me.Databases

			if cmdutil.WantsJSON(cmd) {
				return cmdutil.PrintJSON(squareCli.Out(), databases)
			}

			if len(databases) == 0 {
				fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.database.list.empty"))
				return nil
			}

			w := cmdutil.NewTableWriter(squareCli.Out())
			defer w.Flush()

			t := squareCli.I18n()
			fmt.Fprintln(w, strings.Join([]string{t.T("ui.table.name"), t.T("ui.table.db_id"), t.T("ui.table.type"), t.T("ui.table.memory"), t.T("ui.table.cluster")}, " \t "))
			for _, db := range databases {
				fmt.Fprintf(w, "%s \t %s \t %s \t %s \t %s \t\n",
					db.Name, db.ID, db.Type, strconv.Itoa(db.RAM)+"MB", db.Cluster)
			}

			return nil
		},
	}
}

func newCreateCommand(squareCli cli.SquareCLI) *cobra.Command {
	var options squarecloud.DatabaseCreate

	cmd := &cobra.Command{
		Use:         "create",
		Short:       squareCli.I18n().T("metadata.commands.database.create.short"),
		Annotations: map[string]string{"json": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if options.Name == "" || options.Memory == 0 || options.Type == "" || options.Version == "" {
				return cmd.Help()
			}

			created, err := squareCli.API().Databases.Create(cmd.Context(), options)
			if err != nil {
				return err
			}

			if cmdutil.WantsJSON(cmd) {
				return cmdutil.PrintJSON(squareCli.Out(), created)
			}

			fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.database.create.success", map[string]any{"Name": created.Name, "Id": created.ID}))
			fmt.Fprintln(squareCli.Out(), ui.TextYellow.SetString(squareCli.I18n().T("commands.database.create.password_warning")))
			fmt.Fprintf(squareCli.Out(), "  %s %s\n", squareCli.I18n().T("ui.field.password"), created.Password)
			fmt.Fprintf(squareCli.Out(), "  %s %s\n", squareCli.I18n().T("ui.field.url"), created.ConnectionURL)
			return nil
		},
	}

	cmd.Flags().StringVar(&options.Name, "name", "", squareCli.I18n().T("metadata.flags.db_name"))
	cmd.Flags().IntVar(&options.Memory, "memory", 0, squareCli.I18n().T("metadata.flags.db_memory"))
	cmd.Flags().StringVar((*string)(&options.Type), "type", "", squareCli.I18n().T("metadata.flags.db_type"))
	cmd.Flags().StringVar(&options.Version, "version", "", squareCli.I18n().T("metadata.flags.db_version"))
	return cmd
}

func newInfoCommand(squareCli cli.SquareCLI) *cobra.Command {
	return &cobra.Command{
		Use:         "info [db id]",
		Short:       squareCli.I18n().T("metadata.commands.database.info.short"),
		Annotations: map[string]string{"json": "true"},
		Args:        cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dbID, err := cmdutil.ResolveDatabaseID(cmd.Context(), squareCli, args)
			if err != nil {
				return err
			}

			db, err := squareCli.API().Databases.Get(cmd.Context(), dbID)
			if err != nil {
				return err
			}

			if cmdutil.WantsJSON(cmd) {
				return cmdutil.PrintJSON(squareCli.Out(), db)
			}

			t := squareCli.I18n()
			rows := [][2]string{
				{t.T("ui.field.id"), db.ID},
				{t.T("ui.field.name"), db.Name},
				{t.T("ui.field.type"), string(db.Type)},
				{t.T("ui.field.memory"), strconv.Itoa(db.RAM) + "MB"},
				{t.T("ui.field.port"), strconv.Itoa(db.Port)},
				{t.T("ui.field.cluster"), db.Cluster},
				{t.T("ui.field.owner"), db.Owner},
				{t.T("ui.field.created_at"), db.CreatedAt.Format(time.DateTime)},
			}

			w := cmdutil.NewTableWriter(squareCli.Out())
			for _, row := range rows {
				fmt.Fprintf(w, "%s\t%s\n", row[0], row[1])
			}

			return w.Flush()
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

			dbID, err := cmdutil.ResolveDatabaseID(cmd.Context(), squareCli, args)
			if err != nil {
				return err
			}

			options := squarecloud.DatabaseUpdate{Name: name, RAM: memory}
			if err := squareCli.API().Databases.Update(cmd.Context(), dbID, options); err != nil {
				return err
			}

			fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.database.update.success"))
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", squareCli.I18n().T("metadata.flags.db_new_name"))
	cmd.Flags().IntVar(&memory, "memory", 0, squareCli.I18n().T("metadata.flags.db_new_memory"))
	return cmd
}

func newDeleteCommand(squareCli cli.SquareCLI) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete [db id]",
		Short: squareCli.I18n().T("metadata.commands.database.delete.short"),
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dbID, err := cmdutil.ResolveDatabaseID(cmd.Context(), squareCli, args)
			if err != nil {
				return err
			}

			message := squareCli.I18n().T("commands.database.delete.confirm", map[string]any{"Id": dbID})
			if err := cmdutil.Confirm(squareCli, message, yes); err != nil {
				return err
			}

			if err := squareCli.API().Databases.Delete(cmd.Context(), dbID); err != nil {
				return err
			}

			fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.database.delete.success"))
			return nil
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, squareCli.I18n().T("metadata.flags.yes"))
	return cmd
}

func newStartCommand(squareCli cli.SquareCLI) *cobra.Command {
	return &cobra.Command{
		Use:   "start [db id]",
		Short: squareCli.I18n().T("metadata.commands.database.start.short"),
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dbID, err := cmdutil.ResolveDatabaseID(cmd.Context(), squareCli, args)
			if err != nil {
				return err
			}

			if err := squareCli.API().Databases.Start(cmd.Context(), dbID); err != nil {
				if cmdutil.AlreadyInState(squareCli, err) {
					return nil
				}
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
			dbID, err := cmdutil.ResolveDatabaseID(cmd.Context(), squareCli, args)
			if err != nil {
				return err
			}

			if err := squareCli.API().Databases.Stop(cmd.Context(), dbID); err != nil {
				if cmdutil.AlreadyInState(squareCli, err) {
					return nil
				}
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
		Use:         "status [db id]",
		Short:       squareCli.I18n().T("metadata.commands.database.status.short"),
		Annotations: map[string]string{"json": "true"},
		Args:        cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			api := squareCli.API()

			if all {
				list, err := api.Databases.StatusAll(cmd.Context())
				if err != nil {
					return err
				}

				if cmdutil.WantsJSON(cmd) {
					return cmdutil.PrintJSON(squareCli.Out(), list)
				}

				w := cmdutil.NewTableWriter(squareCli.Out())
				defer w.Flush()

				t := squareCli.I18n()
				fmt.Fprintln(w, strings.Join([]string{t.T("ui.table.db_id"), t.T("ui.table.running"), "CPU", t.T("ui.table.mem")}, " \t "))
				for _, item := range list {
					fmt.Fprintf(w, "%s \t %t \t %s \t %s \t\n", item.ID, item.Running, item.CPU, item.RAM)
				}

				return nil
			}

			dbID, err := cmdutil.ResolveDatabaseID(cmd.Context(), squareCli, args)
			if err != nil {
				return err
			}

			if raw {
				data, err := api.Databases.StatusRaw(cmd.Context(), dbID)
				if err != nil {
					return err
				}

				return cmdutil.PrintJSON(squareCli.Out(), data)
			}

			data, err := api.Databases.Status(cmd.Context(), dbID)
			if err != nil {
				return err
			}

			if cmdutil.WantsJSON(cmd) {
				return cmdutil.PrintJSON(squareCli.Out(), data)
			}

			w := cmdutil.NewTableWriter(squareCli.Out())
			defer w.Flush()

			t := squareCli.I18n()
			fmt.Fprintln(w, strings.Join([]string{t.T("ui.table.db_id"), "CPU %", t.T("ui.table.mem"), t.T("ui.table.disk"), t.T("ui.table.status")}, " \t "))
			fmt.Fprintf(w, "%s \t %s \t %s \t %s \t %s \t\n", dbID, data.CPU, data.RAM, data.Storage, data.Status)

			return nil
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, squareCli.I18n().T("metadata.flags.status_all_databases"))
	cmd.Flags().BoolVar(&raw, "raw", false, squareCli.I18n().T("metadata.flags.raw"))
	return cmd
}

func newMetricsCommand(squareCli cli.SquareCLI) *cobra.Command {
	return &cobra.Command{
		Use:         "metrics [db id]",
		Short:       squareCli.I18n().T("metadata.commands.database.metrics.short"),
		Annotations: map[string]string{"json": "true"},
		Args:        cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dbID, err := cmdutil.ResolveDatabaseID(cmd.Context(), squareCli, args)
			if err != nil {
				return err
			}

			metrics, err := squareCli.API().Databases.Metrics(cmd.Context(), dbID)
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
			dbID, err := cmdutil.ResolveDatabaseID(cmd.Context(), squareCli, args)
			if err != nil {
				return err
			}

			result, err := squareCli.API().Databases.Certificate(cmd.Context(), dbID)
			if err != nil {
				return err
			}

			fmt.Fprintln(squareCli.Out(), result)
			return nil
		},
	}

	var yes bool
	reset := &cobra.Command{
		Use:   "reset <password|certificate> [db id]",
		Short: squareCli.I18n().T("metadata.commands.database.credentials.reset.short"),
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			resetType := squarecloud.DatabaseReset(args[0])
			if resetType != squarecloud.ResetPassword && resetType != squarecloud.ResetCertificate {
				return errors.New(squareCli.I18n().T("commands.database.credentials.reset.invalid_type", map[string]any{"Type": args[0]}))
			}

			dbID, err := cmdutil.ResolveDatabaseID(cmd.Context(), squareCli, args[1:])
			if err != nil {
				return err
			}

			message := squareCli.I18n().T("commands.database.credentials.reset.confirm", map[string]any{"Type": string(resetType), "Id": dbID})
			if err := cmdutil.Confirm(squareCli, message, yes); err != nil {
				return err
			}

			password, err := squareCli.API().Databases.ResetCredentials(cmd.Context(), dbID, resetType)
			if err != nil {
				return err
			}

			fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.database.credentials.reset.success"))
			if password != "" {
				fmt.Fprintln(squareCli.Out(), ui.TextYellow.SetString(squareCli.I18n().T("commands.database.create.password_warning")))
				fmt.Fprintf(squareCli.Out(), "  %s %s\n", squareCli.I18n().T("ui.field.password"), password)
			}

			return nil
		},
	}
	reset.Flags().BoolVarP(&yes, "yes", "y", false, squareCli.I18n().T("metadata.flags.yes"))

	cmd.AddCommand(certificate, reset)
	return cmd
}
