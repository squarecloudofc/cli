package app

import (
	"fmt"
	"io"
	"os"
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

func NewUploadCommand(squareCli cli.SquareCLI) *cobra.Command {
	options := UploadOptions{}

	cmd := &cobra.Command{
		Use:   "upload",
		Short: squareCli.I18n().T("metadata.commands.app.upload.short"),
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
		fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.app.upload.states.compressing"))

		ignoreFiles, _ := squareignore.Load()
		file, err = zipper.ZipFolder(workDir, ignoreFiles)
		if err != nil {
			return err
		}

		_, err = file.Seek(0, io.SeekStart)
		if err != nil {
			return err
		}

		defer os.Remove(file.Name())
	}

	if options.File != "" {
		fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.app.upload.states.loading_file"))

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

	if err != nil {
		fmt.Fprintf(squareCli.Out(), "\n%s %s\n", ui.XMark, squareCli.I18n().T("commands.app.upload.error", map[string]any{
			"Error": err.Error(),
		}))
		return nil
	}

	fmt.Fprintf(squareCli.Out(), "\n%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.app.upload.success"))
	fmt.Fprintf(squareCli.Out(), "  %s\n", squareCli.I18n().T("commands.app.upload.access", map[string]any{
		"Link": ui.TextLink.Render(fmt.Sprintf("https://squarecloud.app/dashboard/app/%s", uploaded.ID)),
	}))

	return nil
}
