// Package file implements `app file` — the remote file manager.
package file

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/cmdutil"
	"github.com/squarecloudofc/cli/internal/ui"
)

func NewCommand(squareCli cli.SquareCLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "file",
		Short: squareCli.I18n().T("metadata.commands.app.file.root.short"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(
		newListCommand(squareCli),
		newReadCommand(squareCli),
		newWriteCommand(squareCli),
		newMoveCommand(squareCli),
		newDeleteCommand(squareCli),
	)

	return cmd
}

func appFlag(cmd *cobra.Command, squareCli cli.SquareCLI, appID *string) {
	cmd.Flags().StringVar(appID, "app", "", squareCli.I18n().T("metadata.flags.app"))
}

func resolveApp(cmd *cobra.Command, squareCli cli.SquareCLI, appID string) (string, error) {
	var args []string
	if appID != "" {
		args = []string{appID}
	}
	return cmdutil.ResolveAppID(cmd.Context(), squareCli, args)
}

func newListCommand(squareCli cli.SquareCLI) *cobra.Command {
	var appID string

	cmd := &cobra.Command{
		Use:         "list [path]",
		Short:       squareCli.I18n().T("metadata.commands.app.file.list.short"),
		Annotations: map[string]string{"json": "true"},
		Args:        cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := resolveApp(cmd, squareCli, appID)
			if err != nil {
				return err
			}

			path := "/"
			if len(args) > 0 {
				path = args[0]
			}

			files, err := squareCli.API().Apps.Files.List(cmd.Context(), id, path)
			if err != nil {
				return err
			}

			if cmdutil.WantsJSON(cmd) {
				return cmdutil.PrintJSON(squareCli.Out(), files)
			}

			if len(files) == 0 {
				fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.app.file.list.empty"))
				return nil
			}

			w := cmdutil.NewTableWriter(squareCli.Out())
			defer w.Flush()

			t := squareCli.I18n()
			fmt.Fprintln(w, strings.Join([]string{t.T("ui.table.type"), t.T("ui.table.name"), t.T("ui.table.size"), t.T("ui.table.modified")}, " \t "))
			for _, f := range files {
				size := "-"
				if f.Type == "file" {
					size = cmdutil.FormatBytes(f.Size)
				}

				modified := "-"
				if f.LastModified != nil {
					modified = time.UnixMilli(int64(*f.LastModified)).Format(time.DateTime)
				}

				fmt.Fprintf(w, "%s \t %s \t %s \t %s \t\n", f.Type, f.Name, size, modified)
			}

			return nil
		},
	}

	appFlag(cmd, squareCli, &appID)
	return cmd
}

func newReadCommand(squareCli cli.SquareCLI) *cobra.Command {
	var appID string
	var output string
	var yes bool

	cmd := &cobra.Command{
		Use:   "read <path>",
		Short: squareCli.I18n().T("metadata.commands.app.file.read.short"),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := resolveApp(cmd, squareCli, appID)
			if err != nil {
				return err
			}

			if _, err := os.Stat(output); output != "" && err == nil {
				message := squareCli.I18n().T("commands.app.file.read.overwrite", map[string]any{"File": output})
				if err := cmdutil.Confirm(squareCli, message, yes); err != nil {
					return err
				}
			}

			content, err := squareCli.API().Apps.Files.Read(cmd.Context(), id, args[0])
			if err != nil {
				return err
			}

			if output != "" {
				// 0600: app files often hold secrets. An existing file
				// keeps its mode (WriteFile only applies it on create).
				if err := os.WriteFile(output, content, 0600); err != nil {
					return err
				}

				fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.app.file.read.saved", map[string]any{"File": output}))
				return nil
			}

			_, err = squareCli.Out().Write(content)
			return err
		},
	}

	appFlag(cmd, squareCli, &appID)
	cmd.Flags().StringVarP(&output, "output", "o", "", squareCli.I18n().T("metadata.flags.output"))
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, squareCli.I18n().T("metadata.flags.yes"))
	return cmd
}

func newWriteCommand(squareCli cli.SquareCLI) *cobra.Command {
	var appID string
	var from string

	cmd := &cobra.Command{
		Use:   "write <path>",
		Short: squareCli.I18n().T("metadata.commands.app.file.write.short"),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := resolveApp(cmd, squareCli, appID)
			if err != nil {
				return err
			}

			var content []byte
			if from != "" {
				content, err = os.ReadFile(from)
			} else {
				content, err = io.ReadAll(squareCli.In())
			}
			if err != nil {
				return err
			}

			if err := squareCli.API().Apps.Files.Write(cmd.Context(), id, args[0], content); err != nil {
				return err
			}

			fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.app.file.write.success", map[string]any{"File": args[0]}))
			return nil
		},
	}

	appFlag(cmd, squareCli, &appID)
	cmd.Flags().StringVar(&from, "from", "", squareCli.I18n().T("metadata.flags.from"))
	return cmd
}

func newMoveCommand(squareCli cli.SquareCLI) *cobra.Command {
	var appID string

	cmd := &cobra.Command{
		Use:     "move <path> <new path>",
		Aliases: []string{"mv", "rename"},
		Short:   squareCli.I18n().T("metadata.commands.app.file.move.short"),
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := resolveApp(cmd, squareCli, appID)
			if err != nil {
				return err
			}

			if err := squareCli.API().Apps.Files.Move(cmd.Context(), id, args[0], args[1]); err != nil {
				return err
			}

			fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.app.file.move.success", map[string]any{"From": args[0], "To": args[1]}))
			return nil
		},
	}

	appFlag(cmd, squareCli, &appID)
	return cmd
}

func newDeleteCommand(squareCli cli.SquareCLI) *cobra.Command {
	var appID string
	var yes bool

	cmd := &cobra.Command{
		Use:     "delete <path>",
		Aliases: []string{"rm"},
		Short:   squareCli.I18n().T("metadata.commands.app.file.delete.short"),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := resolveApp(cmd, squareCli, appID)
			if err != nil {
				return err
			}

			message := squareCli.I18n().T("commands.app.file.delete.confirm", map[string]any{"File": args[0]})
			if err := cmdutil.Confirm(squareCli, message, yes); err != nil {
				return err
			}

			if err := squareCli.API().Apps.Files.Delete(cmd.Context(), id, args[0]); err != nil {
				return err
			}

			fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.app.file.delete.success", map[string]any{"File": args[0]}))
			return nil
		},
	}

	appFlag(cmd, squareCli, &appID)
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, squareCli.I18n().T("metadata.flags.yes"))
	return cmd
}
