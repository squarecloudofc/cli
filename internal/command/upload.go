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

func NewUploadCommand(squareCli *cli.SquareCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upload",
		Short: "Upload your application to Square Cloud",
		RunE:  runUploadCommand(squareCli),
	}

	cmd.PersistentFlags().BoolP("restart", "r", false, "Restart your application when commit")
	// TODO: Future
	//  cmd.PersistentFlags().StringP("file", "f", "", "Zip of the application you want to commit")
	return cmd
}

func runUploadCommand(squareCli *cli.SquareCli) RunEFunc {
	return func(cmd *cobra.Command, args []string) (err error) {
		rest := squareCli.Rest()

		var appId string
		config, er := squareconfig.Load()
		if er != nil {
			return er
		}

		if config.IsCreated() {
			fmt.Fprintln(squareCli.Out(), "Seems you don't have a squarecloud.app config file, create one.")
			return
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

		success, err := rest.ApplicationUpload(appId, file.Name())
		if err != nil {
			return err
		}

		if success.ID != "" {
			config.ID = success.ID
			err := config.Save()

			fmt.Fprintf(squareCli.Out(), "%s Your application has been uploaded\n", ui.CheckMark)
			if err != nil {
				fmt.Fprint(squareCli.Out(), "Unable to save your application id into squarecloud.app config file\n")
			}
		} else {
			fmt.Fprintf(squareCli.Out(), "%s Unable to commit your application\n", ui.XMark)
		}
		return nil
	}
}
