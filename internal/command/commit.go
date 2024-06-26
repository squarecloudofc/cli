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

type CommitOptions struct {
	ConfigFile    *squareconfig.SquareConfig
	ApplicationID string
	Restart       bool
}

func NewCommitCommand(squareCli *cli.SquareCli) *cobra.Command {
	options := CommitOptions{}

	cmd := &cobra.Command{
		Use:   "commit",
		Short: "Commit your application to Square Cloud",
		RunE: func(cmd *cobra.Command, args []string) error {
			config, er := squareconfig.Load()
			if er != nil {
				return er
			}

			if len(args) > 0 {
				options.ApplicationID = args[0]
			} else {
				if config.ID == "" {
					fmt.Fprintln(squareCli.Out(), "You not specified your application ID in command arguments")
					fmt.Fprintln(squareCli.Out(), "You can also specify an ID parameter in your squarecloud.app")
					return nil
				}

				options.ApplicationID = config.ID
			}

			options.ConfigFile = config
			return runCommitCommand(squareCli, &options)
		},
	}

	cmd.Flags().BoolVarP(&options.Restart, "restart", "r", false, "Restart your application when commit")
	return cmd
}

func runCommitCommand(squareCli *cli.SquareCli, options *CommitOptions) error {
	rest := squareCli.Rest()

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

	success, err := rest.ApplicationCommit(options.ApplicationID, options.Restart, file.Name())
	if err != nil {
		return err
	}

	if success {
		fmt.Fprintf(squareCli.Out(), "%s Your source has successfuly commited to Square Cloud\n", ui.CheckMark)
	} else {
		fmt.Fprintf(squareCli.Out(), "%s Unable to commit your application\n", ui.XMark)
	}
	return nil
}
