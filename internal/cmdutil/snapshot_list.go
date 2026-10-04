package cmdutil

import (
	"context"
	"fmt"
	"os"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/sdk-api-go/v3"
)

// NewSnapshotListCommand builds the shared `snapshot list` command used by
// both `app snapshot list` and `db snapshot list` — only the scope differs.
func NewSnapshotListCommand(squareCli cli.SquareCLI, scope squarecloud.SnapshotScope) *cobra.Command {
	return &cobra.Command{
		Use:         "list [id]",
		Short:       squareCli.I18n().T("metadata.commands.snapshot.list.short"),
		Annotations: map[string]string{"json": "true"},
		Args:        cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			snapshots, err := squareCli.API().Account.Snapshots(cmd.Context(), scope)
			if err != nil {
				return err
			}
			if len(args) > 0 { // the JSON output honours the ID too
				snapshots = slices.DeleteFunc(snapshots, func(s squarecloud.Snapshot) bool { return snapshotOwner(s.Name) != args[0] })
			}

			if WantsJSON(cmd) {
				return PrintJSON(squareCli.Out(), snapshots)
			}

			if len(snapshots) < 1 {
				fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.snapshot.list.empty"))
				return nil
			}

			w := NewTableWriter(squareCli.Out())
			defer w.Flush()

			// With an ID filter: one row per snapshot version.
			if len(args) > 0 {
				id := args[0]

				headers := []string{
					"ID",
					squareCli.I18n().T("ui.table.version_id"),
					squareCli.I18n().T("ui.table.size"),
					squareCli.I18n().T("ui.table.last_update"),
				}
				fmt.Fprintln(w, strings.Join(headers, " \t "))

				for _, row := range versionRows(snapshots, id) {
					fmt.Fprintln(w, strings.Join(row, " \t "))
				}

				return nil
			}

			// Without a filter: one summary row per resource.
			type summary struct {
				id         string
				count      int
				totalBytes int64
				lastUpdate time.Time
			}

			groups := map[string]*summary{}
			for _, snap := range snapshots {
				id := snapshotOwner(snap.Name)
				s, ok := groups[id]
				if !ok {
					s = &summary{id: id}
					groups[id] = s
				}

				s.count++
				s.totalBytes += snap.Size
				if snap.Modified.After(s.lastUpdate) {
					s.lastUpdate = snap.Modified
				}
			}

			summaries := make([]*summary, 0, len(groups))
			for _, s := range groups {
				summaries = append(summaries, s)
			}
			sort.Slice(summaries, func(i, j int) bool { return summaries[i].totalBytes > summaries[j].totalBytes })

			headers := []string{
				"ID",
				squareCli.I18n().T("ui.table.quantity"),
				squareCli.I18n().T("ui.table.size"),
				squareCli.I18n().T("ui.table.last_update"),
			}
			fmt.Fprintln(w, strings.Join(headers, " \t "))

			for _, s := range summaries {
				row := []string{
					s.id,
					fmt.Sprintf("%d", s.count),
					FormatBytes(s.totalBytes),
					s.lastUpdate.Format(time.DateTime),
				}
				fmt.Fprintln(w, strings.Join(row, " \t "))
			}

			return nil
		},
	}
}

// snapshotOwner returns the app/database ID a snapshot name belongs to:
// "<id>_<runtime>_<origin>", or just "<id>" for legacy snapshots. IDs are hex,
// so they never contain "_".
func snapshotOwner(name string) string {
	id, _, _ := strings.Cut(name, "_")
	return id
}

// versionRows renders one row per version of the snapshots of resource id. The
// snapshot name and version ID are the ones `snapshot restore` takes.
func versionRows(snapshots []squarecloud.Snapshot, id string) [][]string {
	var rows [][]string
	for _, snap := range snapshots {
		if snapshotOwner(snap.Name) != id {
			continue
		}
		rows = append(rows, []string{
			snap.Name,
			snap.VersionID,
			FormatBytes(snap.Size),
			snap.Modified.Format(time.DateTime),
		})
	}
	return rows
}

// DownloadSnapshot saves a snapshot archive (a signed URL) to destination,
// used by `snapshot create --download`. A partial file is removed.
func DownloadSnapshot(ctx context.Context, api *squarecloud.Client, destination, link string) error {
	// 0600: snapshots carry an app's environment variables or a database's data.
	out, err := os.OpenFile(destination, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}

	err = api.DownloadSnapshot(ctx, link, out)
	if closeErr := out.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		os.Remove(destination)
	}
	return err
}
