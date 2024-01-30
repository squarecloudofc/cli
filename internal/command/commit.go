package command

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/squareconfig"
	"github.com/squarecloudofc/cli/pkg/zipper"
)

func NewCommitCommand(squareCli *cli.SquareCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commit",
		Short: "Commit your application to Square Cloud",
		RunE:  runCommitCommand(squareCli),
	}

	cmd.PersistentFlags().StringP("search", "s", "", "Search for an application")
	return cmd
}

func runCommitCommand(squareCli *cli.SquareCli) RunEFunc {
	return func(cmd *cobra.Command, args []string) (err error) {
		config, err := squareconfig.Load()
		if err != nil {
			return err
		}

		if config.IsCreated() {
			fmt.Fprintln(squareCli.Out(), "seems you don't have a squarecloud.config file, please create one")
			return
		}

		if config.ID == "" {
			fmt.Fprintln(squareCli.Out(), "your squarecloud.config file don't have ID property")
		}

		fmt.Fprintln(squareCli.Out(), "zipping your aplication")
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

		err = zipper.ZipFolder(workDir, file)
		if err != nil {
			return err
		}

		fmt.Fprintf(squareCli.Out(), "sending to square cloud")
		success, err := squareCli.Rest.ApplicationCommit(config.ID, file.Name())
		if err != nil {
			return err
		}

		if success {
			fmt.Fprintln(squareCli.Out(), "Your application has been commited")
		} else {
			fmt.Fprintln(squareCli.Out(), "Unable to commit your application")
		}
		return nil
	}
}
