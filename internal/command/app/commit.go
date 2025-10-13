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
	"github.com/squarecloudofc/sdk-api-go/squarecloud"
)

type CommitOptions struct {
	ConfigFile *squareconfig.SquareConfig

	File          *os.File
	FileName      string
	ApplicationID string
	Restart       bool
}

func NewCommitCommand(squareCli cli.SquareCLI) *cobra.Command {
	options := CommitOptions{}

	cmd := &cobra.Command{
		Use:     "commit",
		Short:   squareCli.I18n().T("metadata.commands.app.commit.short"),
		Aliases: []string{"push"},
		RunE: func(cmd *cobra.Command, args []string) error {
			config, er := squareconfig.Load()
			if er != nil {
				return er
			}

			if len(args) > 0 {
				options.ApplicationID = args[0]
			}

			if len(args) == 0 {
				if config.ID == "" {
					fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.app.commit.arguments.missing"))
					fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.app.commit.arguments.missing_2"))
					return nil
				}

				options.ApplicationID = config.ID
			}

			options.ConfigFile = config
			return runCommitCommand(squareCli, &options)
		},
	}

	cmd.Flags().BoolVarP(&options.Restart, "restart", "r", false, "Restart your application when commit")
	cmd.Flags().StringVar(&options.FileName, "file", "", "File you want to upload to square cloud")
	return cmd
}

func runCommitCommand(squareCli cli.SquareCLI, options *CommitOptions) error {
	var err error

	if options.FileName != "" {
		fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.app.commit.states.loading_file", map[string]any{
			"Filename": filepath.Base(options.FileName),
		}))

		var fileErr error
		options.File, fileErr = handleCommitFile(squareCli, options)
		if fileErr != nil {
			return fileErr
		}
	}

	if options.File == nil {
		fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.app.commit.states.compressing"))

		options.File, err = handleCommitWorkingDirectory()
		if err != nil {
			return err
		}
	}

	if options.File != nil {
		defer options.File.Close()
		if isTemporaryFile(options.File) {
			defer os.Remove(options.File.Name())
		}
	}

	fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.app.commit.states.uploading", map[string]any{
		"Appid": options.ApplicationID,
	}))
	err = squareCli.Rest().PostApplicationCommit(options.ApplicationID, options.File)
	if err != nil {
		return err
	}

	if options.Restart {
		signalErr := squareCli.Rest().PostApplicationSignal(options.ApplicationID, squarecloud.ApplicationSignalRestart)
		if signalErr != nil {
			return signalErr
		}
	}

	fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.app.commit.success"))
	return nil
}

func handleCommitFile(_ cli.SquareCLI, options *CommitOptions) (*os.File, error) {
	if options.FileName == "" {
		return nil, fmt.Errorf("file name is empty")
	}
	workDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	file, err := os.Open(filepath.Join(workDir, options.FileName))
	if err != nil {
		return nil, err
	}

	return file, nil
}

func handleCommitWorkingDirectory() (*os.File, error) {
	workDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	destination, err := os.CreateTemp("", "sc-commit-*.zip")
	if err != nil {
		return nil, err
	}

	ignoreFiles, _ := squareignore.Load()
	err = zipper.ZipFolderW(destination, workDir, ignoreFiles)
	if err != nil {
		return nil, err
	}

	_, err = destination.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}

	return destination, nil
}

func isTemporaryFile(file *os.File) bool {
	tempDir := os.TempDir()
	filePath := file.Name()

	isTempLocation := strings.HasPrefix(filePath, tempDir)

	return isTempLocation
}
