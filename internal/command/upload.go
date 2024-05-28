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

type UploadOptions struct {
	ConfigFile *squareconfig.SquareConfig
	File       string
}

func NewUploadCommand(squareCli *cli.SquareCli) *cobra.Command {
	options := UploadOptions{}

	cmd := &cobra.Command{
		Use:   "upload",
		Short: "Upload your application to Square Cloud",
		RunE: func(cmd *cobra.Command, args []string) error {
			config, er := squareconfig.Load()
			if er != nil {
				return er
			}

			options.ConfigFile = config

			return runUploadCommand(squareCli, &options)
		},
	}

	cmd.Flags().StringVar(&options.File, "file", "", "File you want to upload to square cloud")
	return cmd
}

func runUploadCommand(squareCli *cli.SquareCli, options *UploadOptions) error {
	rest := squareCli.Rest()

	var appId string

	workDir, err := os.Getwd()
	if err != nil {
		return err
	}

	var file *os.File
	if options.File != "" {
		file, err = os.Open(filepath.Join(workDir, options.File))
		if err != nil {
			fmt.Fprintln(squareCli.Out(), "Unable to open the zip file")
			return err
		}
	} else {
		file, err = zipWorkdir(workDir)
		if err != nil {
			fmt.Fprintln(squareCli.Out(), "Unable to zip the working directory")
			return err
		}

		defer file.Close()
		defer os.Remove(file.Name())
	}

	success, err := rest.ApplicationUpload(appId, file.Name())
	if err != nil {
		return err
	}

	if success.ID != "" {
		if !options.ConfigFile.IsCreated() {
			options.ConfigFile.ID = success.ID
			err = options.ConfigFile.Save()
		}

		fmt.Fprintf(squareCli.Out(), "%s Your application has been uploaded\n", ui.CheckMark)
		if err != nil {
			fmt.Fprint(squareCli.Out(), "Unable to save your application id into squarecloud.app config file\n")
		}
	} else {
		fmt.Fprintf(squareCli.Out(), "%s Unable to upload your application\n", ui.XMark)
	}
	return nil
}

func zipWorkdir(wd string) (*os.File, error) {
	zipfilename := path.Join(wd, "*.zip")
	file, err := os.CreateTemp("", filepath.Base(zipfilename))
	if err != nil {
		return nil, err
	}

	ignoreFiles, err := squareignore.Load()
	if err != nil {
		ignoreFiles = []string{}
	}

	err = zipper.ZipFolder(wd, file, ignoreFiles)
	if err != nil {
		return nil, err
	}

	return file, nil
}
