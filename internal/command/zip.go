package command

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/ui"
	"github.com/squarecloudofc/cli/pkg/squareignore"
	"github.com/squarecloudofc/cli/pkg/zipper"
)

func NewZipCommand(squareCli cli.SquareCLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "zip",
		Short:       "Zip the current folder",
		Annotations: map[string]string{"skipAuthCheck": "true"},
		RunE:        runZipCommand(squareCli),
	}

	return cmd
}

func runZipCommand(squareCli cli.SquareCLI) RunEFunc {
	return func(cmd *cobra.Command, args []string) (err error) {
		workDir, err := os.Getwd()
		if err != nil {
			return err
		}

		workDirName := filepath.Base(workDir)

		zipfilename := path.Join(workDir, workDirName+".zip")
		if _, err := os.Lstat(zipfilename); err == nil {
			err := os.Remove(zipfilename)
			if err != nil {
				fmt.Fprintf(squareCli.Err(), "%s.zip already exists and its not possible to delete it\n", workDirName)
				return nil
			}
		}

		ignoreFiles, _ := squareignore.Load()
		file, err := zipper.ZipFolder(workDir, ignoreFiles)
		if err != nil {
			return err
		}
		defer file.Close()

		os.Rename(file.Name(), zipfilename)

		fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.zip.success", map[string]any{"Zip": workDirName}))
		return nil
	}
}
