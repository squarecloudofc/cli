package app

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/cmdutil"
)

func NewLogsCommand(squareCli cli.SquareCLI) *cobra.Command {
	return &cobra.Command{
		Use:         "logs [app id]",
		Short:       squareCli.I18n().T("metadata.commands.app.logs.short"),
		Annotations: map[string]string{"json": "true"},
		Args:        cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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
			return nil
		},
	}
}
