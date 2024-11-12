package app

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/ui"
)

func NewBackupCommand(squareCli *cli.SquareCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Create a backup of you application",
		RunE:  runBackupCommand(squareCli),
	}

	return cmd
}

func runBackupCommand(squareCli *cli.SquareCli) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) (err error) {
		var appId string
		rest := squareCli.Rest()

		if len(args) > 0 {
			appId = args[0]
		}

		if len(args) < 1 {
			id, err := CreateApplicationSelection(squareCli)
			if err != nil {
				return err
			}

			appId = id
		}

		result, err := rest.CreateApplicationBackup(appId)
		if err != nil {
			return err
		}

		if result.URL == "" {
			fmt.Fprintf(squareCli.Out(), "%s It's not possible to download your backup, please try again later...\n", ui.XMark)
		}

		fmt.Fprintln(squareCli.Out(), "Downloading your backup...")

		time := time.Now().Format("2006-01-02 15:04:05")
		filename := fmt.Sprintf("Square Cloud - Backup %s.zip", time)

		err = downloadBackup(filename, result.URL)
		if err != nil {
			fmt.Fprintf(squareCli.Out(), "%s It's not possible to download your backup, please try again later...\n", ui.XMark)
			return
		}

		fmt.Fprintf(squareCli.Out(), "%s Your backup is successfuly downloaded to %s\n", ui.CheckMark, filename)
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
