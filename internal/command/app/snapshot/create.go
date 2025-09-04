package snapshot

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/ui"
	"github.com/squarecloudofc/cli/internal/ui/application_selector"
)

type snapshotOptions struct {
	download bool
}

func NewCreateCommand(squareCli cli.SquareCLI) *cobra.Command {
	options := &snapshotOptions{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: squareCli.I18n().T("metadata.commands.snapshot.create.short"),
		RunE:  runBackupCreateCommand(squareCli, options),
	}

	cmd.Flags().BoolVar(&options.download, "download", false, "Download the backup file")

	return cmd
}

func runBackupCreateCommand(squareCli cli.SquareCLI, options *snapshotOptions) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) (err error) {
		var appId string
		rest := squareCli.Rest()

		if len(args) > 0 {
			appId = args[0]
		}

		if len(args) < 1 {
			m, err := application_selector.RunSelector(squareCli)
			if err != nil {
				return err
			}

			appId = m.ID
		}

		result, err := rest.CreateApplicationBackup(appId)
		if err != nil {
			return err
		}

		if options.download {
			fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.app.snapshot.downloading"))

			timestamp := time.Now().Format("2006-01-02 15:04:05")
			filename := fmt.Sprintf("Square Cloud - Backup %s.zip", timestamp)

			err = downloadBackup(filename, result.URL)
			if err != nil {
				fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.XMark, squareCli.I18n().T("commands.app.snapshot.error"))
				return nil
			}

			fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.app.snapshot.success.downloaded", map[string]any{"File": filename}))
		} else {
			fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.app.snapshot.success.link", map[string]any{"Link": result.URL}))
			fmt.Fprintf(squareCli.Out(), "  %s\n", result.URL)
		}

		return nil
	}
}

func downloadBackup(destination string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
