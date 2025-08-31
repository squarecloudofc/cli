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

func NewCreateCommand(squareCli cli.SquareCLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: squareCli.I18n().T("metadata.commands.snapshot.create.short"),
		RunE:  runBackupCreateCommand(squareCli),
	}

	return cmd
}

func runBackupCreateCommand(squareCli cli.SquareCLI) func(cmd *cobra.Command, args []string) error {
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

		fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.app.snapshot.downloading"))

		time := time.Now().Format("2006-01-02 15:04:05")
		filename := fmt.Sprintf("Square Cloud - Backup %s.zip", time)

		err = downloadBackup(filename, result.URL)
		if err != nil {
			fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.XMark, squareCli.I18n().T("commands.app.snapshot.error"))
			return
		}

		fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.app.snapshot.success", map[string]any{"File": filename}))
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
