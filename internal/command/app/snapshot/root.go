package snapshot

import (
	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/cmdutil"
	"github.com/squarecloudofc/sdk-api-go/v2/squarecloud"
)

func NewCommand(squareCli cli.SquareCLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snapshot",
		Short: squareCli.I18n().T("metadata.commands.app.snapshot.root.short"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(
		NewCreateCommand(squareCli),
		NewRestoreCommand(squareCli),
		cmdutil.NewSnapshotListCommand(squareCli, squarecloud.SnapshotScopeApplications),
	)

	return cmd
}
