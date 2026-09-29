package app

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/cmdutil"
	"github.com/squarecloudofc/cli/internal/ui"
	"github.com/squarecloudofc/cli/pkg/squareconfig"
)

type UploadOptions struct {
	ConfigFile *squareconfig.SquareConfig
	File       string
}

func NewUploadCommand(squareCli cli.SquareCLI) *cobra.Command {
	options := UploadOptions{}

	cmd := &cobra.Command{
		Use:         "upload",
		Short:       squareCli.I18n().T("metadata.commands.app.upload.short"),
		Annotations: map[string]string{"json": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			config, er := squareconfig.Load()
			if er != nil {
				return er
			}

			options.ConfigFile = config

			return runUploadCommand(cmd.Context(), squareCli, &options, cmdutil.WantsJSON(cmd))
		},
	}

	cmd.Flags().StringVar(&options.File, "file", "", squareCli.I18n().T("metadata.flags.file"))
	return cmd
}

func runUploadCommand(ctx context.Context, squareCli cli.SquareCLI, options *UploadOptions, asJSON bool) error {
	workDir, err := os.Getwd()
	if err != nil {
		return err
	}

	// With --json, stdout carries only the result.
	progress := squareCli.Out()
	if asJSON {
		progress = squareCli.Err()
	}

	var file *os.File
	if options.File == "" {
		fmt.Fprintln(progress, squareCli.I18n().T("commands.app.upload.states.compressing"))

		file, err = cmdutil.ZipProject(squareCli, workDir)
		if err != nil {
			return err
		}

		defer os.Remove(file.Name())
	}

	if options.File != "" {
		fmt.Fprintln(progress, squareCli.I18n().T("commands.app.upload.states.loading_file"))

		file, err = os.Open(options.File) // relative to the cwd; absolute paths too
		if err != nil {
			return err
		}
	}

	defer file.Close()

	fmt.Fprintln(progress, squareCli.I18n().T("commands.app.upload.states.uploading"))
	uploaded, err := squareCli.API().Apps.Create(ctx, file)
	if err == nil && uploaded.ID != "" {
		if !options.ConfigFile.IsCreated() {
			options.ConfigFile.ID = uploaded.ID
			options.ConfigFile.Save()
		}
	}

	if err != nil {
		if cli.Hint(squareCli, err) != "" {
			return err // expired key or rate limit: the central handler explains it
		}
		return errors.New(squareCli.I18n().T("commands.app.upload.error", map[string]any{
			"Error": err.Error(),
		}))
	}

	if asJSON {
		return cmdutil.PrintJSON(squareCli.Out(), uploaded)
	}

	fmt.Fprintf(squareCli.Out(), "\n%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.app.upload.success"))
	fmt.Fprintf(squareCli.Out(), "  %s\n", squareCli.I18n().T("commands.app.upload.access", map[string]any{
		"Link": ui.TextLink.Render(fmt.Sprintf("https://squarecloud.app/dashboard/applications/%s", uploaded.ID)),
	}))

	return nil
}
