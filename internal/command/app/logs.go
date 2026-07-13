package app

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/cmdutil"
)

func NewLogsCommand(squareCli cli.SquareCLI) *cobra.Command {
	return &cobra.Command{
		Use:   "logs [app id]",
		Short: squareCli.I18n().T("metadata.commands.app.logs.short"),
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			appID, err := cmdutil.ResolveAppID(squareCli, args)
			if err != nil {
				return err
			}

			result, err := squareCli.Rest().GetApplicationLogs(appID)
			if err != nil {
				return err
			}

			fmt.Fprintln(squareCli.Out(), result.Logs)
			return nil
		},
	}
}
