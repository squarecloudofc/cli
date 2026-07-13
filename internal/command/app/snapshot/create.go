package snapshot

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/cmdutil"
	"github.com/squarecloudofc/cli/internal/ui"
)

func NewCreateCommand(squareCli cli.SquareCLI) *cobra.Command {
	var download bool

	cmd := &cobra.Command{
		Use:   "create [app id]",
		Short: squareCli.I18n().T("metadata.commands.app.snapshot.create.short"),
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			appID, err := cmdutil.ResolveAppID(squareCli, args)
			if err != nil {
				return err
			}

			result, err := squareCli.Rest().CreateApplicationSnapshot(appID)
			if err != nil {
				return err
			}

			if !download {
				fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.snapshot.create.success.link"))
				fmt.Fprintf(squareCli.Out(), "  %s\n", result.URL)
				return nil
			}

			fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.snapshot.create.downloading"))

			filename := fmt.Sprintf("Square Cloud - Snapshot %s.zip", time.Now().Format("2006-01-02 15-04-05"))
			if err := cmdutil.DownloadFile(filename, result.URL); err != nil {
				fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.XMark, squareCli.I18n().T("commands.snapshot.create.error"))
				return err
			}

			fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.snapshot.create.success.downloaded", map[string]any{"File": filename}))
			return nil
		},
	}

	cmd.Flags().BoolVar(&download, "download", false, squareCli.I18n().T("metadata.commands.snapshot.create.flags.download"))
	return cmd
}
