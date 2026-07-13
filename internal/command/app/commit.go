package app

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/pkg/squareconfig"
	"github.com/squarecloudofc/cli/pkg/squareignore"
	"github.com/squarecloudofc/cli/pkg/zipper"
	"github.com/squarecloudofc/sdk-api-go/v2/rest"
	"github.com/squarecloudofc/sdk-api-go/v2/squarecloud"
)

type CommitOptions struct {
	File          *os.File
	FileName      string
	ApplicationID string
	Path          string
	Restart       bool
}

func NewCommitCommand(squareCli cli.SquareCLI) *cobra.Command {
	options := CommitOptions{}

	cmd := &cobra.Command{
		Use:     "commit [app id]",
		Short:   squareCli.I18n().T("metadata.commands.app.commit.short"),
		Aliases: []string{"push"},
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := squareconfig.Load()
			if err != nil {
				return err
			}

			if len(args) > 0 {
				options.ApplicationID = args[0]
			} else {
				if config.ID == "" {
					fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.app.commit.arguments.missing"))
					fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.app.commit.arguments.missing_2"))
					return nil
				}

				options.ApplicationID = config.ID
			}

			return runCommitCommand(squareCli, &options)
		},
	}

	cmd.Flags().BoolVarP(&options.Restart, "restart", "r", false, "Restart your application when commit")
	cmd.Flags().StringVar(&options.FileName, "file", "", "File you want to upload to square cloud")
	cmd.Flags().StringVar(&options.Path, "path", "", "Destination directory inside the application")
	return cmd
}

func runCommitCommand(squareCli cli.SquareCLI, options *CommitOptions) error {
	var err error

	if options.FileName != "" {
		fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.app.commit.states.loading_file", map[string]any{
			"Filename": filepath.Base(options.FileName),
		}))

		options.File, err = openCommitFile(options.FileName)
	} else {
		fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.app.commit.states.compressing"))

		options.File, err = zipCommitWorkingDirectory()
	}
	if err != nil {
		return err
	}

	defer options.File.Close()
	if isTemporaryFile(options.File) {
		defer os.Remove(options.File.Name())
	}

	fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.app.commit.states.uploading", map[string]any{
		"Appid": options.ApplicationID,
	}))

	var requestOpts []rest.RequestOpt
	if options.Path != "" {
		requestOpts = append(requestOpts, rest.WithQueryParam("path", options.Path))
	}

	if err := squareCli.Rest().PostApplicationCommit(options.ApplicationID, options.File, requestOpts...); err != nil {
		return err
	}

	if options.Restart {
		if err := squareCli.Rest().PostApplicationSignal(options.ApplicationID, squarecloud.ApplicationSignalRestart); err != nil {
			return err
		}
	}

	fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.app.commit.success"))
	return nil
}

func openCommitFile(filename string) (*os.File, error) {
	workDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	return os.Open(filepath.Join(workDir, filename))
}

func zipCommitWorkingDirectory() (*os.File, error) {
	workDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	destination, err := os.CreateTemp("", "sc-commit-*.zip")
	if err != nil {
		return nil, err
	}

	ignoreFiles, _ := squareignore.Load()
	if err := zipper.ZipFolderW(destination, workDir, ignoreFiles); err != nil {
		return nil, err
	}

	if _, err := destination.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	return destination, nil
}

func isTemporaryFile(file *os.File) bool {
	return strings.HasPrefix(file.Name(), os.TempDir())
}
