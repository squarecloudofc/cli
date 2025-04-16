package upload

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
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

func NewCommand(squareCli *cli.SquareCli) *cobra.Command {
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
	m, err := NewModel(squareCli, options.ConfigFile)
	if err != nil {
		fmt.Fprint(squareCli.Out(), "Unable to send application to Square Cloud due to: ", err.Error())
		return nil
	}

	if _, err := tea.NewProgram(m).Run(); err != nil {
		return err
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

	if err := file.Close(); err != nil {
		return nil, err
	}

	file, err = os.Open(file.Name())
	if err != nil {
		return nil, err
	}

	return file, nil
}
