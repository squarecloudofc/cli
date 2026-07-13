package snapshot

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/cmdutil"
	"github.com/squarecloudofc/cli/internal/ui"
	"github.com/squarecloudofc/sdk-api-go/v2/squarecloud"
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
		newRestoreCommand(squareCli),
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
			dbID, err := cmdutil.ResolveDatabaseID(squareCli, args)
			if err != nil {
				return err
			}

			result, err := squareCli.Rest().CreateDatabaseSnapshot(dbID)
			if err != nil {
				return err
			}

			if !download {
				fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.snapshot.create.success.link"))
				fmt.Fprintf(squareCli.Out(), "  %s\n", result.URL)
				return nil
			}

			fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.snapshot.create.downloading"))

			filename := fmt.Sprintf("Square Cloud - DB Snapshot %s.zip", time.Now().Format("2006-01-02 15-04-05"))
			if err := cmdutil.DownloadFile(filename, result.URL); err != nil {
				fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.XMark, squareCli.I18n().T("commands.snapshot.create.error"))
				return err
			}

			fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.snapshot.create.success.downloaded", map[string]any{"File": filename}))
			return nil
		},
	}

	cmd.Flags().BoolVar(&download, "download", false, squareCli.I18n().T("metadata.commands.snapshot.create.flags.download"))
	return cmd
}

func newRestoreCommand(squareCli cli.SquareCLI) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "restore <db id> <snapshot id> <version id>",
		Short: squareCli.I18n().T("metadata.commands.database.snapshot.restore.short"),
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			dbID, snapshotID, versionID := args[0], args[1], args[2]

			message := squareCli.I18n().T("commands.snapshot.restore.confirm", map[string]any{"Id": dbID})
			if !cmdutil.Confirm(squareCli, message, yes) {
				return nil
			}

			if err := squareCli.Rest().RestoreDatabaseSnapshot(dbID, snapshotID, versionID); err != nil {
				return err
			}

			fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.snapshot.restore.success"))
			return nil
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip the confirmation prompt")
	return cmd
}
