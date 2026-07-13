package snapshot

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/cmdutil"
	"github.com/squarecloudofc/cli/internal/ui"
)

func NewRestoreCommand(squareCli cli.SquareCLI) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "restore <app id> <snapshot id> <version id>",
		Short: squareCli.I18n().T("metadata.commands.app.snapshot.restore.short"),
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			appID, snapshotID, versionID := args[0], args[1], args[2]

			message := squareCli.I18n().T("commands.snapshot.restore.confirm", map[string]any{"Id": appID})
			if !cmdutil.Confirm(squareCli, message, yes) {
				return nil
			}

			if err := squareCli.Rest().RestoreApplicationSnapshot(appID, snapshotID, versionID); err != nil {
				return err
			}

			fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.snapshot.restore.success"))
			return nil
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip the confirmation prompt")
	return cmd
}
