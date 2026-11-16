package command

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/command/app"
	"github.com/squarecloudofc/cli/internal/integration"
	"github.com/squarecloudofc/cli/internal/ui"
	"github.com/squarecloudofc/cli/pkg/squareconfig"
)

// NewIntegrationCommand manages "Send to Square Cloud" in the Windows
// Explorer menu; other systems get an explanation.
func NewIntegrationCommand(squareCli cli.SquareCLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "integration",
		Short: squareCli.I18n().T("metadata.commands.integration.root.short"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newIntegrationInstall(squareCli), newIntegrationUninstall(squareCli), newIntegrationSend(squareCli))
	return cmd
}

func newIntegrationInstall(squareCli cli.SquareCLI) *cobra.Command {
	var dryRun bool
	cmd := &cobra.Command{
		Use:         "install",
		Short:       squareCli.I18n().T("metadata.commands.integration.install.short"),
		Annotations: map[string]string{"skipAuthCheck": "true"},
		Args:        cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			t := squareCli.I18n()
			if integration.System == nil {
				return errors.New(t.T("commands.integration.unsupported"))
			}
			exe, err := os.Executable()
			if err == nil {
				exe, err = filepath.EvalSymlinks(exe)
			}
			if err != nil {
				return err
			}

			values := integration.Values(exe, t.T("commands.integration.menu_label"))
			if dryRun {
				fmt.Fprintln(squareCli.Out(), t.T("commands.integration.dry_run.install"))
				for _, v := range values {
					name := v.Name
					if name == "" {
						name = "(default)"
					}
					fmt.Fprintf(squareCli.Out(), "  HKCU\\%s  %s = %s\n", v.Key, name, v.Data)
				}
				return nil
			}

			if err := integration.Install(integration.System, values); err != nil {
				return err
			}
			fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, t.T("commands.integration.installed"))
			return nil
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, squareCli.I18n().T("metadata.flags.dry_run"))
	return cmd
}

func newIntegrationUninstall(squareCli cli.SquareCLI) *cobra.Command {
	var dryRun bool
	cmd := &cobra.Command{
		Use:         "uninstall",
		Short:       squareCli.I18n().T("metadata.commands.integration.uninstall.short"),
		Annotations: map[string]string{"skipAuthCheck": "true"},
		Args:        cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			t := squareCli.I18n()
			if integration.System == nil {
				return errors.New(t.T("commands.integration.unsupported"))
			}
			if dryRun {
				fmt.Fprintln(squareCli.Out(), t.T("commands.integration.dry_run.uninstall"))
				for _, key := range integration.Keys() {
					fmt.Fprintf(squareCli.Out(), "  HKCU\\%s\n", key)
				}
				return nil
			}

			if err := integration.Uninstall(integration.System); err != nil {
				return err
			}
			fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, t.T("commands.integration.uninstalled"))
			return nil
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, squareCli.I18n().T("metadata.flags.dry_run"))
	return cmd
}

// newIntegrationSend is what the Explorer menu runs. The console window
// closes when it ends, so the result is also shown as a notification.
func newIntegrationSend(squareCli cli.SquareCLI) *cobra.Command {
	return &cobra.Command{
		Use:    "send <path>",
		Hidden: true,
		// Checked in sendToSquareCloud, so a logged-out run still notifies.
		Annotations: map[string]string{"skipAuthCheck": "true"},
		Args:        cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			t := squareCli.I18n()
			err := sendToSquareCloud(cmd.Context(), squareCli, args[0])
			if err == nil {
				_ = integration.Notify("Square Cloud", t.T("commands.integration.sent", map[string]any{"Name": filepath.Base(args[0])}))
				return nil
			}
			message := cli.Hint(squareCli, err)
			if message == "" {
				message = err.Error()
			}
			_ = integration.Notify(t.T("commands.integration.send_failed", map[string]any{"Name": filepath.Base(args[0])}), message)
			return err
		},
	}
}

// sendToSquareCloud commits a folder whose squarecloud.app has an ID to that
// application, and uploads any other folder or .zip as a new one: what
// commit or upload would do run from there.
func sendToSquareCloud(ctx context.Context, squareCli cli.SquareCLI, path string) error {
	if !cli.CheckAuth(squareCli) {
		return errors.New(squareCli.I18n().T("errors.common.not_logged"))
	}
	path, err := filepath.Abs(path) // before changing folders
	if err != nil {
		return err
	}
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if !info.IsDir() {
		if err := os.Chdir(filepath.Dir(path)); err != nil {
			return err
		}
		return app.Upload(ctx, squareCli, path)
	}

	if err := os.Chdir(path); err != nil {
		return err
	}
	if config, err := squareconfig.Load(); err == nil && config.ID != "" {
		return app.Commit(ctx, squareCli, config.ID)
	}
	return app.Upload(ctx, squareCli, "")
}
