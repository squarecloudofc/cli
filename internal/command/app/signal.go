package app

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/cmdutil"
	"github.com/squarecloudofc/cli/internal/ui"
)

func NewStartCommand(squareCli cli.SquareCLI) *cobra.Command {
	return &cobra.Command{
		Use:   "start [app id]",
		Short: squareCli.I18n().T("metadata.commands.app.signal.start.short"),
		Args:  cobra.MaximumNArgs(1),
		RunE: runSendSignal(squareCli, "START", func(ctx context.Context, appID string) error {
			return squareCli.API().Apps.Start(ctx, appID)
		}),
	}
}

func NewRestartCommand(squareCli cli.SquareCLI) *cobra.Command {
	return &cobra.Command{
		Use:   "restart [app id]",
		Short: squareCli.I18n().T("metadata.commands.app.signal.restart.short"),
		Args:  cobra.MaximumNArgs(1),
		RunE: runSendSignal(squareCli, "RESTART", func(ctx context.Context, appID string) error {
			return squareCli.API().Apps.Restart(ctx, appID)
		}),
	}
}

func NewStopCommand(squareCli cli.SquareCLI) *cobra.Command {
	return &cobra.Command{
		Use:   "stop [app id]",
		Short: squareCli.I18n().T("metadata.commands.app.signal.stop.short"),
		Args:  cobra.MaximumNArgs(1),
		RunE: runSendSignal(squareCli, "STOP", func(ctx context.Context, appID string) error {
			return squareCli.API().Apps.Stop(ctx, appID)
		}),
	}
}

func runSendSignal(squareCli cli.SquareCLI, signal string, send func(ctx context.Context, appID string) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		appID, err := cmdutil.ResolveAppID(cmd.Context(), squareCli, args)
		if err != nil {
			return err
		}

		if err := send(cmd.Context(), appID); err != nil {
			if cmdutil.AlreadyInState(squareCli, err) {
				return nil
			}
			fmt.Fprintf(squareCli.Err(), "%s %s\n", ui.XMark, squareCli.I18n().T("commands.app.signal.failed"))
			return err
		}

		fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.app.signal.success", map[string]any{
			"Signal": signal,
		}))
		return nil
	}
}
