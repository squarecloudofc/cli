package cmdutil

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/ui"
	"github.com/squarecloudofc/sdk-api-go/v4"
)

// restorePoll is how often a restore still running is checked: the API
// allows 30 state reads a minute.
const restorePoll = 5 * time.Second

// sleep waits d unless ctx ends first. A variable so tests skip the wait.
var sleep = func(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// NewSnapshotRestoreCommand builds `snapshot restore` for `app snapshot` and
// `db snapshot`; only the scope differs.
func NewSnapshotRestoreCommand(squareCli cli.SquareCLI, scope squarecloud.SnapshotScope) *cobra.Command {
	var yes, noWait bool

	use, short := "restore <app id> <snapshot id> <version id>", "metadata.commands.app.snapshot.restore.short"
	if scope == squarecloud.SnapshotScopeDatabases {
		use, short = "restore <db id> <snapshot id> <version id>", "metadata.commands.database.snapshot.restore.short"
	}

	cmd := &cobra.Command{
		Use:   use,
		Short: squareCli.I18n().T(short),
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, snapshotID, versionID := args[0], args[1], args[2]

			message := squareCli.I18n().T("commands.snapshot.restore.confirm", map[string]any{"Id": id})
			if err := Confirm(squareCli, message, yes); err != nil {
				return err
			}

			snapshots := squareCli.API().Apps.Snapshots
			if scope == squarecloud.SnapshotScopeDatabases {
				snapshots = squareCli.API().Databases.Snapshots
			}

			result, err := snapshots.Restore(cmd.Context(), id, snapshotID, versionID)
			if err != nil {
				return err
			}

			if !result.Pending {
				fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.snapshot.restore.success"))
				return nil
			}
			if noWait {
				fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.TextBlue.Render("i"), squareCli.I18n().T("commands.snapshot.restore.background"))
				return nil
			}

			return followRestore(cmd.Context(), squareCli, func(ctx context.Context) (squarecloud.SnapshotRestoreState, error) {
				return snapshots.RestoreState(ctx, id)
			})
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, squareCli.I18n().T("metadata.flags.yes"))
	cmd.Flags().BoolVar(&noWait, "no-wait", false, squareCli.I18n().T("metadata.flags.no_wait"))
	return cmd
}

// followRestore reads the state of a restore that outlived the API's
// response deadline every restorePoll, printing each new step, until it
// succeeds or fails. Ctrl+C stops watching; the restore keeps running.
func followRestore(ctx context.Context, squareCli cli.SquareCLI, state func(context.Context) (squarecloud.SnapshotRestoreState, error)) error {
	t := squareCli.I18n()
	out := squareCli.Out()
	background := func() {
		fmt.Fprintf(out, "%s %s\n", ui.TextBlue.Render("i"), t.T("commands.snapshot.restore.background"))
	}

	fmt.Fprintln(out, t.T("commands.snapshot.restore.following"))

	var step string
	for {
		if sleep(ctx, restorePoll) != nil {
			background()
			return nil
		}

		st, err := state(ctx)
		if err != nil {
			background()
			if ctx.Err() != nil {
				return nil
			}
			return err
		}

		switch st.State {
		case "success":
			fmt.Fprintf(out, "%s %s\n", ui.CheckMark, t.T("commands.snapshot.restore.success"))
			return nil
		case "error":
			return errors.New(t.T("commands.snapshot.restore.failed", map[string]any{"Code": cmp.Or(st.Code, squarecloud.CodeSnapshotRestoreFailed)}))
		case "queued", "running":
		default:
			// "none" (no record of the restore: expired, or the node lost
			// it) or a state this version doesn't know: polling on might
			// never end.
			return errors.New(t.T("commands.snapshot.restore.lost"))
		}

		if st.Step != "" && st.Step != step {
			step = st.Step
			fmt.Fprintf(out, "  %s\n", t.T("commands.snapshot.restore.step", map[string]any{"Step": step}))
		}
	}
}
