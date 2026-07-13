package app

import (
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/cmdutil"
	"github.com/squarecloudofc/cli/internal/ui"
	"github.com/squarecloudofc/sdk-api-go/v2/rest"
)

func NewRealtimeCommand(squareCli cli.SquareCLI) *cobra.Command {
	return &cobra.Command{
		Use:   "realtime [app id]",
		Short: squareCli.I18n().T("metadata.commands.app.realtime.short"),
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			appID, err := cmdutil.ResolveAppID(squareCli, args)
			if err != nil {
				return err
			}

			stream, err := squareCli.Rest().ApplicationRealtime(appID, rest.WithContext(cmd.Context()))
			if err != nil {
				return err
			}
			defer stream.Close()

			fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.app.realtime.connected", map[string]any{"Appid": appID}))

			for {
				event, err := stream.Next()
				if err != nil {
					if errors.Is(err, io.EOF) || cmd.Context().Err() != nil {
						return nil
					}
					return err
				}

				if event.Event == "message" {
					fmt.Fprintln(squareCli.Out(), event.Data)
					continue
				}

				fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.TextYellow.SetString("["+event.Event+"]"), event.Data)
			}
		},
	}
}
