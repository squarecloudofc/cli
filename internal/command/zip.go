package command

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
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
		fmt.Fprintln(squareCli.Out(), "zipping your aplication")
		workDir, err := os.Getwd()
		if err != nil {
			return err
		}

		zipfilename := path.Join(workDir, "source.zip")
		file, err := os.CreateTemp("", "*.zip")
		if err != nil {
			return err
		}
		defer file.Close()

		err = zipper.ZipFolder(workDir, file)
		if err != nil {
			return err
		}

		os.Rename(file.Name(), zipfilename)

		fmt.Fprintf(squareCli.Out(), "your source has successfuly zipped")
		return nil
	}
}
