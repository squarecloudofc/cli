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

func NewZipCommand(squareCli *cli.SquareCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "zip",
		Short: "Zip the current folder",
		RunE:  runZipCommand(squareCli),
	}

	return cmd
}

func runZipCommand(squareCli *cli.SquareCli) RunEFunc {
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
				fmt.Fprintln(squareCli.Err(), "source.zip already exists and its not possible to delete it")
			}
		}

		file, err := os.CreateTemp("", "*.zip")
		if err != nil {
			return err
		}
		defer file.Close()

		ignoreFiles, err := squareignore.Load()
		if err != nil {
			ignoreFiles = []string{}
		}

		err = zipper.ZipFolder(workDir, file, ignoreFiles)
		if err != nil {
			return err
		}

		os.Rename(file.Name(), zipfilename)

		fmt.Fprintf(squareCli.Out(), "%s Your source has successfuly zipped to %s.zip\n", ui.CheckMark, workDirName)
		return nil
	}
}
