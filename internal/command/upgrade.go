package command

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/build"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/ui"
	"github.com/squarecloudofc/cli/internal/updater"
)

// NewUpgradeCommand replaces the running binary with the latest release.
func NewUpgradeCommand(squareCli cli.SquareCLI) *cobra.Command {
	var check bool

	cmd := &cobra.Command{
		Use:         "upgrade",
		Short:       squareCli.I18n().T("metadata.commands.upgrade.short"),
		Annotations: map[string]string{"skipAuthCheck": "true"},
		Args:        cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			t := squareCli.I18n()
			out := squareCli.Out()

			latest, err := updater.FetchLatest(cmd.Context(), updater.LatestReleaseURL)
			if err != nil {
				return errors.New(t.T("commands.upgrade.check_failed", map[string]any{"Error": err.Error()}))
			}

			if !updater.IsOutdated(build.Version, latest) {
				fmt.Fprintf(out, "%s %s\n", ui.CheckMark, t.T("commands.upgrade.up_to_date", map[string]any{"Version": build.Version, "Latest": latest}))
				return nil
			}

			exe, err := os.Executable()
			if err == nil {
				exe, err = filepath.EvalSymlinks(exe)
			}
			if err != nil {
				return err
			}
			// npm and the system package manager own their copy: say how to
			// update it instead of replacing a file they track.
			var managed string
			const releases = "https://github.com/squarecloudofc/cli/releases/latest"
			switch {
			case updater.InstalledWithNPM(exe):
				managed = t.T("commands.upgrade.npm")
			case updater.PackagedBy(exe) == "deb":
				managed = t.T("commands.upgrade.package_deb", map[string]any{"Link": releases})
			case updater.PackagedBy(exe) == "rpm":
				managed = t.T("commands.upgrade.package_rpm", map[string]any{"Link": releases})
			}

			if check {
				fmt.Fprintf(out, "%s %s\n", ui.TextBlue.Render("i"), t.T("commands.upgrade.available", map[string]any{"Version": build.Version, "Latest": latest}))
				if managed != "" {
					fmt.Fprintln(out, managed)
				} else {
					fmt.Fprintln(out, t.T("commands.upgrade.run"))
				}
				return nil
			}
			if managed != "" {
				return errors.New(managed)
			}

			fmt.Fprintln(out, t.T("commands.upgrade.downloading", map[string]any{"Latest": latest}))
			bin, err := updater.Download(cmd.Context(), latest, runtime.GOOS, runtime.GOARCH)
			if err != nil {
				return err
			}
			if err := updater.Replace(exe, bin, runtime.GOOS); err != nil {
				return errors.New(t.T("commands.upgrade.replace_failed", map[string]any{"Path": exe, "Error": err.Error()}))
			}

			fmt.Fprintf(out, "%s %s\n", ui.CheckMark, t.T("commands.upgrade.success", map[string]any{"Latest": latest}))
			return nil
		},
	}

	cmd.Flags().BoolVar(&check, "check", false, squareCli.I18n().T("metadata.flags.check"))
	return cmd
}
