package app

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/pkg/squareconfig"
	"github.com/squarecloudofc/cli/pkg/squareignore"
	"github.com/squarecloudofc/cli/pkg/zipper"
)

type UploadOptions struct {
	ConfigFile *squareconfig.SquareConfig
	File       string
}

func NewUploadCommand(squareCli cli.SquareCLI) *cobra.Command {
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

func runUploadCommand(squareCli cli.SquareCLI, options *UploadOptions) error {
	workDir, err := os.Getwd()
	if err != nil {
		return err
	}

	var file *os.File
	if options.File == "" {
		fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.app.upload.states.loading_file"))

		ignoreFiles, _ := squareignore.Load()
		file, err := zipper.ZipFolder(workDir, ignoreFiles)
		if err != nil {
			return err
		}

		defer os.Remove(file.Name())
	}

	if options.File != "" {
		fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.app.upload.states.compressing"))

		file, err = os.Open(filepath.Join(workDir, options.File))
		if err != nil {
			return err
		}
	}
	defer file.Close()

	fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.app.upload.states.uploading"))
	uploaded, err := squareCli.Rest().PostApplications(file)
	if err == nil && uploaded.ID != "" {
		if !options.ConfigFile.IsCreated() {
			options.ConfigFile.ID = uploaded.ID
			options.ConfigFile.Save()
		}
	}

	return nil
}
