package app

import (
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/cmdutil"
	"github.com/squarecloudofc/cli/internal/ui"
)

func NewRealtimeCommand(squareCli cli.SquareCLI) *cobra.Command {
	return &cobra.Command{
		Use:   "realtime [app id]",
		Short: squareCli.I18n().T("metadata.commands.app.realtime.short"),
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			appID, err := cmdutil.ResolveAppID(cmd.Context(), squareCli, args)
			if err != nil {
				return err
			}

			stream, err := squareCli.API().Apps.Realtime(cmd.Context(), appID)
			if err != nil {
				return err
			}
			defer stream.Close()

			fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.app.realtime.connected", map[string]any{"Appid": appID}))

			var lastStatus string
			for {
				event, err := stream.Next()
				if err != nil {
					if errors.Is(err, io.EOF) || cmd.Context().Err() != nil {
						return nil
					}
					return err
				}

				switch event.Event {
				case "logs":
					out := squareCli.Out()
					if event.Stream == "stderr" {
						out = squareCli.Err()
					}
					fmt.Fprintln(out, event.Line)
				case "status":
					// Status frames arrive every few seconds; show only changes.
					if event.Status.Status != lastStatus {
						lastStatus = event.Status.Status
						fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.TextYellow.SetString("[status]"), lastStatus)
					}
				default:
					fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.TextYellow.SetString("["+event.Event+"]"), event.Data)
				}
			}
		},
	}
}
