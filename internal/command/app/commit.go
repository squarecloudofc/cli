package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/cmdutil"
	"github.com/squarecloudofc/cli/pkg/squareconfig"
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
					return errors.New(squareCli.I18n().T("commands.app.commit.arguments.missing") + "\n" +
						squareCli.I18n().T("commands.app.commit.arguments.missing_2"))
				}

				options.ApplicationID = config.ID
			}

			return runCommitCommand(cmd.Context(), squareCli, &options)
		},
	}

	cmd.Flags().BoolVarP(&options.Restart, "restart", "r", false, squareCli.I18n().T("metadata.flags.restart"))
	cmd.Flags().StringVar(&options.FileName, "file", "", squareCli.I18n().T("metadata.flags.file"))
	cmd.Flags().StringVar(&options.Path, "path", "", squareCli.I18n().T("metadata.flags.path"))
	return cmd
}

func runCommitCommand(ctx context.Context, squareCli cli.SquareCLI, options *CommitOptions) error {
	var err error

	if options.FileName != "" {
		fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.app.commit.states.loading_file", map[string]any{
			"Filename": filepath.Base(options.FileName),
		}))

		options.File, err = openCommitFile(options.FileName)
	} else {
		fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.app.commit.states.compressing"))

		options.File, err = cmdutil.ZipProject(squareCli, ".")
	}
	if err != nil {
		return err
	}

	if options.FileName == "" { // our temp zip, never the user's --file
		defer os.Remove(options.File.Name())
	}
	defer options.File.Close() // runs first: Windows cannot remove an open file

	fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.app.commit.states.uploading", map[string]any{
		"Appid": options.ApplicationID,
	}))

	if err := squareCli.API().Apps.Commit(ctx, options.ApplicationID, options.File, options.Path, ""); err != nil {
		return err
	}

	if options.Restart {
		if err := squareCli.API().Apps.Restart(ctx, options.ApplicationID); err != nil {
			return err
		}
	}

	fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.app.commit.success"))
	return nil
}

func openCommitFile(filename string) (*os.File, error) {
	return os.Open(filename) // relative to the cwd; absolute paths too
}
