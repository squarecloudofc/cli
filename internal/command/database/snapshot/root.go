package snapshot

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/cmdutil"
	"github.com/squarecloudofc/cli/internal/ui"
	"github.com/squarecloudofc/sdk-api-go/v4"
)

func NewCommand(squareCli cli.SquareCLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snapshot",
		Short: squareCli.I18n().T("metadata.commands.database.snapshot.root.short"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(
		newCreateCommand(squareCli),
		cmdutil.NewSnapshotRestoreCommand(squareCli, squarecloud.SnapshotScopeDatabases),
		cmdutil.NewSnapshotListCommand(squareCli, squarecloud.SnapshotScopeDatabases),
	)

	return cmd
}

func newCreateCommand(squareCli cli.SquareCLI) *cobra.Command {
	var download bool

	cmd := &cobra.Command{
		Use:   "create [db id]",
		Short: squareCli.I18n().T("metadata.commands.database.snapshot.create.short"),
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dbID, err := cmdutil.ResolveDatabaseID(cmd.Context(), squareCli, args)
			if err != nil {
				return err
			}

			result, err := squareCli.API().Databases.Snapshots.Create(cmd.Context(), dbID)
			if err != nil {
				return err
			}

			if result.Pending {
				fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.snapshot.create.pending"))
				return nil
			}

			if !download {
				fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.snapshot.create.success.link"))
				fmt.Fprintf(squareCli.Out(), "  %s\n", result.URL)
				return nil
			}

			fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.snapshot.create.downloading"))

			filename := fmt.Sprintf("Square Cloud - DB Snapshot %s.zip", time.Now().Format("2006-01-02 15-04-05"))
			if err := cmdutil.DownloadSnapshot(cmd.Context(), squareCli.API(), filename, result.URL); err != nil {
				fmt.Fprintf(squareCli.Err(), "%s %s\n", ui.XMark, squareCli.I18n().T("commands.snapshot.create.error"))
				return err
			}

			fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.snapshot.create.success.downloaded", map[string]any{"File": filename}))
			return nil
		},
	}

	cmd.Flags().BoolVar(&download, "download", false, squareCli.I18n().T("metadata.flags.download"))
	return cmd
}
