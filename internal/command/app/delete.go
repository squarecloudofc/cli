package app

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/ui"
	"github.com/squarecloudofc/cli/internal/ui/application_selector"
)

func NewDeleteCommand(squareCli cli.SquareCLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: squareCli.I18n().T("metadata.commands.app.delete.short"),
		RunE:  runDeleteCommand(squareCli),
	}

	return cmd
}

func runDeleteCommand(squareCli cli.SquareCLI) func(cmd *cobra.Command, args []string) error {
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

		err = rest.DeleteApplication(appId)
		if err != nil {
			fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.XMark, squareCli.I18n().T("commands.app.delete.failed"))
			return err
		}

		fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.app.delete.success"))
		return nil
	}
}
