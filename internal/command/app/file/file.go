// Package file implements `app file` — the remote file manager.
package file

import (
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
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

func appFlag(cmd *cobra.Command, appID *string) {
	cmd.Flags().StringVar(appID, "app", "", "Application ID")
}

func resolveApp(squareCli cli.SquareCLI, appID string) (string, error) {
	var args []string
	if appID != "" {
		args = []string{appID}
	}
	return cmdutil.ResolveAppID(squareCli, args)
}

func newListCommand(squareCli cli.SquareCLI) *cobra.Command {
	var appID string

	cmd := &cobra.Command{
		Use:   "list [path]",
		Short: squareCli.I18n().T("metadata.commands.app.file.list.short"),
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := resolveApp(squareCli, appID)
			if err != nil {
				return err
			}

			path := "/"
			if len(args) > 0 {
				path = args[0]
			}

			files, err := squareCli.Rest().GetApplicationFiles(id, path)
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

			w := tabwriter.NewWriter(squareCli.Out(), 0, 0, 2, ' ', tabwriter.TabIndent)
			defer w.Flush()

			fmt.Fprintln(w, strings.Join([]string{"TYPE", "NAME", "SIZE", "MODIFIED"}, " \t "))
			for _, f := range files {
				size := "-"
				if f.Type == "file" {
					size = cmdutil.FormatBytes(int64(f.Size))
				}

				fmt.Fprintf(w, "%s \t %s \t %s \t %s \t\n",
					f.Type, f.Name, size, time.UnixMilli(f.LastModified).Format(time.DateTime))
			}

			return nil
		},
	}

	appFlag(cmd, &appID)
	return cmd
}

func newReadCommand(squareCli cli.SquareCLI) *cobra.Command {
	var appID string
	var output string

	cmd := &cobra.Command{
		Use:   "read <path>",
		Short: squareCli.I18n().T("metadata.commands.app.file.read.short"),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := resolveApp(squareCli, appID)
			if err != nil {
				return err
			}

			content, err := squareCli.Rest().ReadApplicationFile(id, args[0])
			if err != nil {
				return err
			}

			if output != "" {
				if err := os.WriteFile(output, content.Data, 0644); err != nil {
					return err
				}

				fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.app.file.read.saved", map[string]any{"File": output}))
				return nil
			}

			_, err = squareCli.Out().Write(content.Data)
			return err
		},
	}

	appFlag(cmd, &appID)
	cmd.Flags().StringVarP(&output, "output", "o", "", "Write the content to a local file")
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
			id, err := resolveApp(squareCli, appID)
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

			result, err := squareCli.Rest().PutApplicationFile(id, args[0], content)
			if err != nil {
				return err
			}
			if !result.Written {
				return fmt.Errorf("%s", squareCli.I18n().T("commands.app.file.write.failed"))
			}

			fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.app.file.write.success", map[string]any{"File": args[0]}))
			return nil
		},
	}

	appFlag(cmd, &appID)
	cmd.Flags().StringVar(&from, "from", "", "Read the content from a local file (default: stdin)")
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
			id, err := resolveApp(squareCli, appID)
			if err != nil {
				return err
			}

			if err := squareCli.Rest().MoveApplicationFile(id, args[0], args[1]); err != nil {
				return err
			}

			fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.app.file.move.success", map[string]any{"From": args[0], "To": args[1]}))
			return nil
		},
	}

	appFlag(cmd, &appID)
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
			id, err := resolveApp(squareCli, appID)
			if err != nil {
				return err
			}

			message := squareCli.I18n().T("commands.app.file.delete.confirm", map[string]any{"File": args[0]})
			if !cmdutil.Confirm(squareCli, message, yes) {
				return nil
			}

			if err := squareCli.Rest().DeleteApplicationFile(id, args[0]); err != nil {
				return err
			}

			fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.app.file.delete.success", map[string]any{"File": args[0]}))
			return nil
		},
	}

	appFlag(cmd, &appID)
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip the confirmation prompt")
	return cmd
}
