package app

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/cmdutil"
	"github.com/squarecloudofc/cli/internal/ui"
)

func NewDeleteCommand(squareCli cli.SquareCLI) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete [app id]",
		Short: squareCli.I18n().T("metadata.commands.app.delete.short"),
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			appID, err := cmdutil.ResolveAppID(cmd.Context(), squareCli, args)
			if err != nil {
				return err
			}

			message := squareCli.I18n().T("commands.app.delete.confirm", map[string]any{"Appid": appID})
			if err := cmdutil.Confirm(squareCli, message, yes); err != nil {
				return err
			}

			if err := squareCli.API().Apps.Delete(cmd.Context(), appID); err != nil {
				fmt.Fprintf(squareCli.Err(), "%s %s\n", ui.XMark, squareCli.I18n().T("commands.app.delete.failed"))
				return err
			}

			fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.app.delete.success"))
			return nil
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, squareCli.I18n().T("metadata.flags.yes"))
	return cmd
}
