// Package workspace implements `workspace` — team workspace management.
package workspace

import (
	"fmt"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/cmdutil"
	"github.com/squarecloudofc/cli/internal/ui"
	"github.com/squarecloudofc/sdk-api-go/v2/squarecloud"
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
		Use:   "list",
		Short: squareCli.I18n().T("metadata.commands.workspace.list.short"),
		RunE: func(cmd *cobra.Command, args []string) error {
			workspaces, err := squareCli.Rest().GetWorkspaces()
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

			w := tabwriter.NewWriter(squareCli.Out(), 0, 0, 2, ' ', tabwriter.TabIndent)
			defer w.Flush()

			fmt.Fprintln(w, strings.Join([]string{"NAME", "WORKSPACE ID", "MEMBERS", "APPS", "CREATED"}, " \t "))
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
			created, err := squareCli.Rest().CreateWorkspace(args[0])
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
		Use:   "info <workspace id>",
		Short: squareCli.I18n().T("metadata.commands.workspace.info.short"),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ws, err := squareCli.Rest().GetWorkspace(args[0])
			if err != nil {
				return err
			}

			if cmdutil.WantsJSON(cmd) {
				return cmdutil.PrintJSON(squareCli.Out(), ws)
			}

			fmt.Fprintf(squareCli.Out(), "%-12s %s\n", "ID:", ws.ID)
			fmt.Fprintf(squareCli.Out(), "%-12s %s\n", "Name:", ws.Name)
			fmt.Fprintf(squareCli.Out(), "%-12s %s\n", "Owner:", ws.Owner)
			fmt.Fprintf(squareCli.Out(), "%-12s %s\n", "Created at:", ws.CreatedAt.Format(time.DateTime))

			if len(ws.Members) > 0 {
				fmt.Fprintf(squareCli.Out(), "\n%s\n", squareCli.I18n().T("commands.workspace.info.members"))

				w := tabwriter.NewWriter(squareCli.Out(), 0, 0, 2, ' ', tabwriter.TabIndent)
				fmt.Fprintln(w, strings.Join([]string{"  NAME", "MEMBER ID", "GROUP", "JOINED"}, " \t "))
				for _, member := range ws.Members {
					fmt.Fprintf(w, "  %s \t %s \t %s \t %s \t\n",
						member.Name, member.ID, member.Group, member.JoinedAt.Format(time.DateOnly))
				}
				w.Flush()
			}

			if len(ws.Applications) > 0 {
				fmt.Fprintf(squareCli.Out(), "\n%s\n", squareCli.I18n().T("commands.workspace.info.applications"))

				w := tabwriter.NewWriter(squareCli.Out(), 0, 0, 2, ' ', tabwriter.TabIndent)
				fmt.Fprintln(w, strings.Join([]string{"  NAME", "APP ID", "LANG", "MEMORY"}, " \t "))
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
			if !cmdutil.Confirm(squareCli, message, yes) {
				return nil
			}

			if err := squareCli.Rest().DeleteWorkspace(args[0]); err != nil {
				return err
			}

			fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.workspace.delete.success"))
			return nil
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip the confirmation prompt")
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
			if !cmdutil.Confirm(squareCli, message, yes) {
				return nil
			}

			if err := squareCli.Rest().LeaveWorkspace(args[0]); err != nil {
				return err
			}

			fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.workspace.leave.success"))
			return nil
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip the confirmation prompt")
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
			if err := squareCli.Rest().AddWorkspaceApplication(args[0], args[1]); err != nil {
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
			if err := squareCli.Rest().RemoveWorkspaceApplication(args[0], args[1]); err != nil {
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
			group := squarecloud.WorkspaceMemberGroup(args[2])
			if err := squareCli.Rest().AddWorkspaceMember(args[0], args[1], group); err != nil {
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
			group := squarecloud.WorkspaceMemberGroup(args[2])
			if err := squareCli.Rest().UpdateWorkspaceMember(args[0], args[1], group); err != nil {
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
			if !cmdutil.Confirm(squareCli, message, yes) {
				return nil
			}

			if err := squareCli.Rest().RemoveWorkspaceMember(args[0], args[1]); err != nil {
				return err
			}

			fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.workspace.member.remove.success"))
			return nil
		},
	}
	remove.Flags().BoolVarP(&yes, "yes", "y", false, "Skip the confirmation prompt")

	inviteCode := &cobra.Command{
		Use:   "invite-code",
		Short: squareCli.I18n().T("metadata.commands.workspace.member.invite_code.short"),
		RunE: func(cmd *cobra.Command, args []string) error {
			code, err := squareCli.Rest().GetWorkspaceInviteCode()
			if err != nil {
				return err
			}

			fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.workspace.member.invite_code.success"))
			fmt.Fprintf(squareCli.Out(), "  %s\n", code.Code)
			return nil
		},
	}

	cmd.AddCommand(add, update, remove, inviteCode)
	return cmd
}
