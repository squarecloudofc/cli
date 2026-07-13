package app

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/cmdutil"
	"github.com/squarecloudofc/cli/internal/ui"
	"github.com/squarecloudofc/sdk-api-go/v2/squarecloud"
)

func NewStartCommand(squareCli cli.SquareCLI) *cobra.Command {
	return &cobra.Command{
		Use:   "start [app id]",
		Short: squareCli.I18n().T("metadata.commands.app.signal.start.short"),
		Args:  cobra.MaximumNArgs(1),
		RunE:  runSendSignal(squareCli, squarecloud.ApplicationSignalStart),
	}
}

func NewRestartCommand(squareCli cli.SquareCLI) *cobra.Command {
	return &cobra.Command{
		Use:   "restart [app id]",
		Short: squareCli.I18n().T("metadata.commands.app.signal.restart.short"),
		Args:  cobra.MaximumNArgs(1),
		RunE:  runSendSignal(squareCli, squarecloud.ApplicationSignalRestart),
	}
}

func NewStopCommand(squareCli cli.SquareCLI) *cobra.Command {
	return &cobra.Command{
		Use:   "stop [app id]",
		Short: squareCli.I18n().T("metadata.commands.app.signal.stop.short"),
		Args:  cobra.MaximumNArgs(1),
		RunE:  runSendSignal(squareCli, squarecloud.ApplicationSignalStop),
	}
}

func runSendSignal(squareCli cli.SquareCLI, signal squarecloud.ApplicationSignal) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		appID, err := cmdutil.ResolveAppID(squareCli, args)
		if err != nil {
			return err
		}

		if err := squareCli.Rest().PostApplicationSignal(appID, signal); err != nil {
			fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.XMark, squareCli.I18n().T("commands.app.signal.failed"))
			return err
		}

		fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.app.signal.success", map[string]any{
			"Signal": string(signal),
		}))
		return nil
	}
}
