package app

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/cmdutil"
	"github.com/squarecloudofc/sdk-api-go/v4"
)

func NewLogsCommand(squareCli cli.SquareCLI) *cobra.Command {
	var follow bool

	cmd := &cobra.Command{
		Use:         "logs [app id]",
		Short:       squareCli.I18n().T("metadata.commands.app.logs.short"),
		Annotations: map[string]string{"json": "true"},
		Args:        cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if follow && cmdutil.WantsJSON(cmd) {
				return errors.New(squareCli.I18n().T("errors.common.json_unsupported", map[string]any{"Command": cmd.CommandPath() + " --follow"}))
			}

			appID, err := cmdutil.ResolveAppID(cmd.Context(), squareCli, args)
			if err != nil {
				return err
			}

			result, err := squareCli.API().Apps.Logs(cmd.Context(), appID)
			if err != nil {
				return err
			}

			if cmdutil.WantsJSON(cmd) {
				return cmdutil.PrintJSON(squareCli.Out(), map[string]string{"logs": result})
			}

			fmt.Fprintln(squareCli.Out(), result)
			if !follow {
				return nil
			}

			// Like tail -f: from here on, only new lines.
			stream, err := squareCli.API().Apps.Realtime(cmd.Context(), appID)
			if err != nil {
				return err
			}
			return followRealtime(cmd.Context(), squareCli, appID, stream, func(event squarecloud.RealtimeEvent) {
				if event.Event == "logs" {
					printLogLine(squareCli, event)
				}
			})
		},
	}

	cmd.Flags().BoolVarP(&follow, "follow", "f", false, squareCli.I18n().T("metadata.flags.follow"))
	return cmd
}
