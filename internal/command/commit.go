package command

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/ui"
	"github.com/squarecloudofc/cli/pkg/squareconfig"
	"github.com/squarecloudofc/cli/pkg/squareignore"
	"github.com/squarecloudofc/cli/pkg/zipper"
)

func NewCommitCommand(squareCli *cli.SquareCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commit",
		Short: "Commit your application to Square Cloud",
		RunE:  runCommitCommand(squareCli),
	}

	cmd.PersistentFlags().BoolP("restart", "r", false, "Restart your application when commit")
	// TODO: Future
	//  cmd.PersistentFlags().StringP("file", "f", "", "File name you want to commit")
	return cmd
}

func runCommitCommand(squareCli *cli.SquareCli) RunEFunc {
	return func(cmd *cobra.Command, args []string) (err error) {
		rest := squareCli.Rest()

		var appId string

		if len(args) < 1 {
			config, er := squareconfig.Load()
			if er != nil {
				return er
			}

			if !config.IsCreated() && config.ID == "" {
				fmt.Fprintln(squareCli.Out(), "You not specified your application ID in command arguments")
				fmt.Fprintln(squareCli.Out(), "You can also specify an ID parameter in your squarecloud.app")
				return
			}
		} else {
			appId = args[0]
		}

		workDir, err := os.Getwd()
		if err != nil {
			return err
		}

		zipfilename := path.Join(workDir, "*.zip")
		file, err := os.CreateTemp("", filepath.Base(zipfilename))
		if err != nil {
			return err
		}
		defer file.Close()
		defer os.Remove(file.Name())

		ignoreFiles, err := squareignore.Load()
		if err != nil {
			ignoreFiles = []string{}
		}

		err = zipper.ZipFolder(workDir, file, ignoreFiles)
		if err != nil {
			return err
		}

		success, err := rest.ApplicationCommit(appId, file.Name())
		if err != nil {
			return err
		}

		if success {
			fmt.Fprintf(squareCli.Out(), "%s Your application has been commited\n", ui.CheckMark)
		} else {
			fmt.Fprintf(squareCli.Out(), "%s Unable to commit your application\n", ui.XMark)
		}
		return nil
	}
}
