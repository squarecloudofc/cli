// Package blob holds the `squarecloud blob` commands for Square Cloud Blob.
package blob

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/blob"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/cmdutil"
	"github.com/squarecloudofc/cli/internal/ui"
	"golang.org/x/term"
)

func NewCommand(squareCli cli.SquareCLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "blob",
		Short: squareCli.I18n().T("metadata.commands.blob.root.short"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(
		newListCommand(squareCli),
		newUploadCommand(squareCli),
		newDeleteCommand(squareCli),
		newURLCommand(squareCli),
		newStatsCommand(squareCli),
	)
	return cmd
}

func client(squareCli cli.SquareCLI) *blob.Client {
	return &blob.Client{Key: cli.APIKey(squareCli.Config()), HTTPClient: cli.HTTPClient(), UserAgent: cli.UserAgent()}
}

func newListCommand(squareCli cli.SquareCLI) *cobra.Command {
	return &cobra.Command{
		Use:         "list [prefix]",
		Short:       squareCli.I18n().T("metadata.commands.blob.list.short"),
		Annotations: map[string]string{"json": "true"},
		Args:        cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var prefix string
			if len(args) > 0 {
				prefix = args[0]
			}
			objects, err := client(squareCli).List(cmd.Context(), prefix)
			if err != nil {
				return err
			}

			if cmdutil.WantsJSON(cmd) {
				if objects == nil {
					objects = []blob.Object{} // [] rather than null for scripts
				}
				return cmdutil.PrintJSON(squareCli.Out(), objects)
			}

			t := squareCli.I18n()
			if len(objects) == 0 {
				fmt.Fprintln(squareCli.Out(), t.T("commands.blob.list.empty"))
				return nil
			}

			w := cmdutil.NewTableWriter(squareCli.Out())
			defer w.Flush()
			fmt.Fprintln(w, strings.Join([]string{"ID", t.T("ui.table.size"), t.T("ui.table.created"), t.T("ui.table.url")}, " \t "))
			for _, o := range objects {
				link := t.T("commands.blob.private")
				if o.URL != nil {
					link = *o.URL
				}
				fmt.Fprintln(w, strings.Join([]string{o.ID, cmdutil.FormatBytes(o.Size), shortDate(o.CreatedAt), link}, " \t "))
			}
			return nil
		},
	}
}

func newUploadCommand(squareCli cli.SquareCLI) *cobra.Command {
	var name string
	var private bool

	cmd := &cobra.Command{
		Use:         "upload <file>",
		Short:       squareCli.I18n().T("metadata.commands.blob.upload.short"),
		Annotations: map[string]string{"json": "true"},
		Args:        cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			t := squareCli.I18n()
			info, err := os.Stat(args[0])
			if err != nil {
				return err
			}
			if info.IsDir() {
				return fmt.Errorf("%s", t.T("commands.blob.upload.is_dir", map[string]any{"Path": args[0]}))
			}

			// Progress goes to stderr: stdout carries only the result.
			fmt.Fprintln(squareCli.Err(), t.T("commands.blob.upload.uploading", map[string]any{"File": filepath.Base(args[0]), "Size": cmdutil.FormatBytes(info.Size())}))
			var progress func(done, total int)
			if f, ok := squareCli.Err().(*os.File); ok && term.IsTerminal(int(f.Fd())) {
				progress = func(done, total int) {
					fmt.Fprintf(squareCli.Err(), "\r%s", t.T("commands.blob.upload.parts", map[string]any{"Done": done, "Total": total}))
					if done == total {
						fmt.Fprintln(squareCli.Err())
					}
				}
			}

			res, err := client(squareCli).Upload(cmd.Context(), args[0], name, private, progress)
			if err != nil {
				return err
			}

			if cmdutil.WantsJSON(cmd) {
				return cmdutil.PrintJSON(squareCli.Out(), res)
			}
			fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, t.T("commands.blob.upload.success", map[string]any{"ID": res.ID}))
			if res.URL != nil {
				fmt.Fprintf(squareCli.Out(), "  %s\n", *res.URL)
			} else {
				fmt.Fprintf(squareCli.Out(), "  %s\n", t.T("commands.blob.upload.private_hint", map[string]any{"ID": res.ID}))
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", squareCli.I18n().T("metadata.flags.blob_name"))
	cmd.Flags().BoolVar(&private, "private", false, squareCli.I18n().T("metadata.flags.blob_private"))
	return cmd
}

func newDeleteCommand(squareCli cli.SquareCLI) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: squareCli.I18n().T("metadata.commands.blob.delete.short"),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			t := squareCli.I18n()
			if err := cmdutil.Confirm(squareCli, t.T("commands.blob.delete.confirm", map[string]any{"ID": args[0]}), yes); err != nil {
				return err
			}
			if err := client(squareCli).Delete(cmd.Context(), args[0]); err != nil {
				return err
			}
			fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, t.T("commands.blob.delete.success", map[string]any{"ID": args[0]}))
			return nil
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, squareCli.I18n().T("metadata.flags.yes"))
	return cmd
}

func newURLCommand(squareCli cli.SquareCLI) *cobra.Command {
	return &cobra.Command{
		Use:         "url <id>",
		Short:       squareCli.I18n().T("metadata.commands.blob.url.short"),
		Annotations: map[string]string{"json": "true"},
		Args:        cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			link, err := client(squareCli).DownloadURL(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			if cmdutil.WantsJSON(cmd) {
				return cmdutil.PrintJSON(squareCli.Out(), link)
			}
			// Only the link on stdout, so $(squarecloud blob url <id>) works.
			fmt.Fprintln(squareCli.Out(), link.URL)
			if link.ExpiresAt != nil {
				fmt.Fprintln(squareCli.Err(), squareCli.I18n().T("commands.blob.url.expires", map[string]any{"Date": shortDate(*link.ExpiresAt)}))
			}
			return nil
		},
	}
}

func newStatsCommand(squareCli cli.SquareCLI) *cobra.Command {
	return &cobra.Command{
		Use:         "stats",
		Short:       squareCli.I18n().T("metadata.commands.blob.stats.short"),
		Annotations: map[string]string{"json": "true"},
		Args:        cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			stats, err := client(squareCli).Stats(cmd.Context())
			if err != nil {
				return err
			}
			if cmdutil.WantsJSON(cmd) {
				return cmdutil.PrintJSON(squareCli.Out(), stats)
			}
			t := squareCli.I18n()
			fmt.Fprintln(squareCli.Out(), t.T("commands.blob.stats.objects", map[string]any{"Count": stats.Usage.Objects}))
			fmt.Fprintln(squareCli.Out(), t.T("commands.blob.stats.storage", map[string]any{
				"Used":     cmdutil.FormatBytes(stats.Usage.Storage),
				"Included": cmdutil.FormatBytes(stats.Plan.Included),
			}))
			return nil
		},
	}
}

// shortDate renders an RFC 3339 time as local "2006-01-02 15:04", or as is.
func shortDate(s string) string {
	if parsed, err := time.Parse(time.RFC3339, s); err == nil {
		return parsed.Local().Format("2006-01-02 15:04")
	}
	return s
}
