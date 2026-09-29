package command

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/cmdutil"
	"github.com/squarecloudofc/cli/internal/ui"
)

func NewZipCommand(squareCli cli.SquareCLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "zip",
		Short:       squareCli.I18n().T("metadata.commands.zip.short"),
		Annotations: map[string]string{"skipAuthCheck": "true"},
		RunE:        runZipCommand(squareCli),
	}

	return cmd
}

func runZipCommand(squareCli cli.SquareCLI) func(cmd *cobra.Command, args []string) error {
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
				return errors.New(squareCli.I18n().T("commands.zip.exists", map[string]any{"File": workDirName + ".zip", "Error": err.Error()}))
			}
		}

		file, err := cmdutil.ZipProject(squareCli, workDir)
		if err != nil {
			return err
		}
		defer os.Remove(file.Name())
		defer file.Close()

		// Copied, not renamed: the temp dir is often another filesystem
		// (tmpfs /tmp, %TEMP% on another drive) and rename fails across them.
		err = copyFile(file, zipfilename)
		if err != nil {
			fmt.Fprintf(squareCli.Err(), "%s %s\n", ui.XMark, squareCli.I18n().T("commands.zip.failed", map[string]any{"Zip": workDirName}))
			return err
		}

		fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.zip.success", map[string]any{"Zip": workDirName}))
		return nil
	}
}

// copyFile writes src (from its start) to dst, removing dst on failure.
func copyFile(src *os.File, dst string) error {
	if _, err := src.Seek(0, io.SeekStart); err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}

	_, err = io.Copy(out, src)
	if closeErr := out.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		os.Remove(dst)
	}
	return err
}
