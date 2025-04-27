package app

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/ui"
	"github.com/squarecloudofc/cli/internal/ui/application_selector"
	"github.com/squarecloudofc/sdk-api-go/squarecloud"
)

func NewStartCommand(squareCli cli.SquareCLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start your application",
		RunE:  runSendSignal(squareCli, squarecloud.ApplicationSignalStart),
	}

	return cmd
}

func NewRestartCommand(squareCli cli.SquareCLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restart",
		Short: "Restart your application",
		RunE:  runSendSignal(squareCli, squarecloud.ApplicationSignalRestart),
	}

	return cmd
}

func NewStopCommand(squareCli cli.SquareCLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop your application",
		RunE:  runSendSignal(squareCli, squarecloud.ApplicationSignalStop),
	}

	return cmd
}

func runSendSignal(squareCli cli.SquareCLI, signal squarecloud.ApplicationSignal) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) (err error) {
		var appId string
		rest := squareCli.Rest()

		if len(args) > 0 {
			appId = args[0]
		}

		if len(args) < 1 {
			m, err := application_selector.RunSelector(squareCli)
			if err != nil {
				return err
			}

			appId = m.ID
		}

		err = rest.PostApplicationSignal(appId, signal)
		if err != nil {
			fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.XMark, squareCli.I18n().T("commands.app.signal.failed"))
			return
		}

		fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.app.signal.success", map[string]any{
			"Signal": string(signal),
		}))
		return nil
	}
}
