package command

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/squareconfig"
	"github.com/squarecloudofc/cli/internal/ui"
	"github.com/squarecloudofc/cli/pkg/zipper"
)

func NewCommitCommand(squareCli *cli.SquareCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commit",
		Short: "Commit your application to Square Cloud",
		RunE:  runCommitCommand(squareCli),
	}

	return cmd
}

func runCommitCommand(squareCli *cli.SquareCli) RunEFunc {
	return func(cmd *cobra.Command, args []string) (err error) {
		self, err := squareCli.Rest.SelfUser()
		if err != nil {
			return err
		}

		if self == nil || self.User.Tag == "" {
			fmt.Fprintf(squareCli.Out(), "No user associated with current Square Cloud Token\n")
			return
		}

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

		success, err := squareCli.Rest.ApplicationCommit(config.ID, file.Name())
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
