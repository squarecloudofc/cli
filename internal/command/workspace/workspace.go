// Package workspace implements `workspace` — team workspace management.
package workspace

import (
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
		Use:     "workspace",
		Aliases: []string{"ws"},
		Short:   squareCli.I18n().T("metadata.commands.workspace.root.short"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(
		newListCommand(squareCli),
		newCreateCommand(squareCli),
		newInfoCommand(squareCli),
		newDeleteCommand(squareCli),
		newLeaveCommand(squareCli),
		newAppCommand(squareCli),
		newMemberCommand(squareCli),
	)

	return cmd
}

func newListCommand(squareCli cli.SquareCLI) *cobra.Command {
	return &cobra.Command{
		Use:         "list",
		Short:       squareCli.I18n().T("metadata.commands.workspace.list.short"),
		Annotations: map[string]string{"json": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			workspaces, err := squareCli.API().Workspaces.List(cmd.Context())
			if err != nil {
				return err
			}

			if cmdutil.WantsJSON(cmd) {
				return cmdutil.PrintJSON(squareCli.Out(), workspaces)
			}

			if len(workspaces) == 0 {
				fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.workspace.list.empty"))
				return nil
			}

			w := cmdutil.NewTableWriter(squareCli.Out())
			defer w.Flush()

			t := squareCli.I18n()
			fmt.Fprintln(w, strings.Join([]string{t.T("ui.table.name"), t.T("ui.table.workspace_id"), t.T("ui.table.members"), t.T("ui.table.apps"), t.T("ui.table.created")}, " \t "))
			for _, ws := range workspaces {
				fmt.Fprintf(w, "%s \t %s \t %d \t %d \t %s \t\n",
					ws.Name, ws.ID, len(ws.Members), len(ws.Applications), ws.CreatedAt.Format(time.DateOnly))
			}

			return nil
		},
	}
}

func newCreateCommand(squareCli cli.SquareCLI) *cobra.Command {
	return &cobra.Command{
		Use:   "create <name>",
		Short: squareCli.I18n().T("metadata.commands.workspace.create.short"),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			created, err := squareCli.API().Workspaces.Create(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.workspace.create.success", map[string]any{
				"Name": created.Name,
				"Id":   created.ID,
			}))
			return nil
		},
	}
}

func newInfoCommand(squareCli cli.SquareCLI) *cobra.Command {
	return &cobra.Command{
		Use:         "info <workspace id>",
		Short:       squareCli.I18n().T("metadata.commands.workspace.info.short"),
		Annotations: map[string]string{"json": "true"},
		Args:        cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ws, err := squareCli.API().Workspaces.Get(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			if cmdutil.WantsJSON(cmd) {
				return cmdutil.PrintJSON(squareCli.Out(), ws)
			}

			t := squareCli.I18n()
			w := cmdutil.NewTableWriter(squareCli.Out())
			fmt.Fprintf(w, "%s\t%s\n", t.T("ui.field.id"), ws.ID)
			fmt.Fprintf(w, "%s\t%s\n", t.T("ui.field.name"), ws.Name)
			fmt.Fprintf(w, "%s\t%s\n", t.T("ui.field.owner"), ws.Owner)
			fmt.Fprintf(w, "%s\t%s\n", t.T("ui.field.created_at"), ws.CreatedAt.Format(time.DateTime))
			w.Flush()

			if len(ws.Members) > 0 {
				fmt.Fprintf(squareCli.Out(), "\n%s\n", squareCli.I18n().T("commands.workspace.info.members"))

				w := cmdutil.NewTableWriter(squareCli.Out())
				fmt.Fprintln(w, strings.Join([]string{"  " + t.T("ui.table.name"), t.T("ui.table.member_id"), t.T("ui.table.group"), t.T("ui.table.joined")}, " \t "))
				for _, member := range ws.Members {
					name := "-"
					if member.Name != nil {
						name = *member.Name
					}
					fmt.Fprintf(w, "  %s \t %s \t %s \t %s \t\n",
						name, member.ID, member.Group, member.JoinedAt.Format(time.DateOnly))
				}
				w.Flush()
			}

			if len(ws.Applications) > 0 {
				fmt.Fprintf(squareCli.Out(), "\n%s\n", squareCli.I18n().T("commands.workspace.info.applications"))

				w := cmdutil.NewTableWriter(squareCli.Out())
				fmt.Fprintln(w, strings.Join([]string{"  " + t.T("ui.table.name"), t.T("ui.table.app_id"), t.T("ui.table.lang"), t.T("ui.table.memory")}, " \t "))
				for _, app := range ws.Applications {
					fmt.Fprintf(w, "  %s \t %s \t %s \t %dMB \t\n", app.Name, app.ID, app.Lang, app.RAM)
				}
				w.Flush()
			}

			return nil
		},
	}
}

func newDeleteCommand(squareCli cli.SquareCLI) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <workspace id>",
		Short: squareCli.I18n().T("metadata.commands.workspace.delete.short"),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			message := squareCli.I18n().T("commands.workspace.delete.confirm", map[string]any{"Id": args[0]})
			if err := cmdutil.Confirm(squareCli, message, yes); err != nil {
				return err
			}

			if err := squareCli.API().Workspaces.Delete(cmd.Context(), args[0]); err != nil {
				return err
			}

			fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.workspace.delete.success"))
			return nil
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, squareCli.I18n().T("metadata.flags.yes"))
	return cmd
}

func newLeaveCommand(squareCli cli.SquareCLI) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "leave <workspace id>",
		Short: squareCli.I18n().T("metadata.commands.workspace.leave.short"),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			message := squareCli.I18n().T("commands.workspace.leave.confirm", map[string]any{"Id": args[0]})
			if err := cmdutil.Confirm(squareCli, message, yes); err != nil {
				return err
			}

			if err := squareCli.API().Workspaces.Leave(cmd.Context(), args[0]); err != nil {
				return err
			}

			fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.workspace.leave.success"))
			return nil
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, squareCli.I18n().T("metadata.flags.yes"))
	return cmd
}

func newAppCommand(squareCli cli.SquareCLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "app",
		Short: squareCli.I18n().T("metadata.commands.workspace.app.root.short"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	add := &cobra.Command{
		Use:   "add <workspace id> <app id>",
		Short: squareCli.I18n().T("metadata.commands.workspace.app.add.short"),
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := squareCli.API().Workspaces.Apps.Add(cmd.Context(), args[0], args[1]); err != nil {
				return err
			}

			fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.workspace.app.add.success"))
			return nil
		},
	}

	remove := &cobra.Command{
		Use:   "remove <workspace id> <app id>",
		Short: squareCli.I18n().T("metadata.commands.workspace.app.remove.short"),
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := squareCli.API().Workspaces.Apps.Remove(cmd.Context(), args[0], args[1]); err != nil {
				return err
			}

			fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.workspace.app.remove.success"))
			return nil
		},
	}

	cmd.AddCommand(add, remove)
	return cmd
}

func newMemberCommand(squareCli cli.SquareCLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "member",
		Short: squareCli.I18n().T("metadata.commands.workspace.member.root.short"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	add := &cobra.Command{
		Use:   "add <workspace id> <invite code> <group>",
		Short: squareCli.I18n().T("metadata.commands.workspace.member.add.short"),
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			group := squarecloud.WorkspaceGroup(args[2])
			if err := squareCli.API().Workspaces.Members.Add(cmd.Context(), args[0], args[1], group); err != nil {
				return err
			}

			fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.workspace.member.add.success"))
			return nil
		},
	}

	update := &cobra.Command{
		Use:   "update <workspace id> <member id> <group>",
		Short: squareCli.I18n().T("metadata.commands.workspace.member.update.short"),
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			group := squarecloud.WorkspaceGroup(args[2])
			if err := squareCli.API().Workspaces.Members.Update(cmd.Context(), args[0], args[1], group); err != nil {
				return err
			}

			fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.workspace.member.update.success"))
			return nil
		},
	}

	var yes bool
	remove := &cobra.Command{
		Use:   "remove <workspace id> <member id>",
		Short: squareCli.I18n().T("metadata.commands.workspace.member.remove.short"),
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			message := squareCli.I18n().T("commands.workspace.member.remove.confirm", map[string]any{"Id": args[1]})
			if err := cmdutil.Confirm(squareCli, message, yes); err != nil {
				return err
			}

			if err := squareCli.API().Workspaces.Members.Remove(cmd.Context(), args[0], args[1]); err != nil {
				return err
			}

			fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.workspace.member.remove.success"))
			return nil
		},
	}
	remove.Flags().BoolVarP(&yes, "yes", "y", false, squareCli.I18n().T("metadata.flags.yes"))

	inviteCode := &cobra.Command{
		Use:   "invite-code",
		Short: squareCli.I18n().T("metadata.commands.workspace.member.invite_code.short"),
		RunE: func(cmd *cobra.Command, args []string) error {
			code, err := squareCli.API().Workspaces.Members.InviteCode(cmd.Context())
			if err != nil {
				return err
			}

			fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.workspace.member.invite_code.success"))
			fmt.Fprintf(squareCli.Out(), "  %s\n", code)
			return nil
		},
	}

	cmd.AddCommand(add, update, remove, inviteCode)
	return cmd
}
