package cmdutil

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/sdk-api-go/v2/squarecloud"
)

// NewSnapshotListCommand builds the shared `snapshot list` command used by
// both `app snapshot list` and `db snapshot list` — only the scope differs.
func NewSnapshotListCommand(squareCli cli.SquareCLI, scope squarecloud.SnapshotScope) *cobra.Command {
	return &cobra.Command{
		Use:   "list [id]",
		Short: squareCli.I18n().T("metadata.commands.snapshot.list.short"),
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			snapshots, err := squareCli.Rest().UserSnapshots(scope)
			if err != nil {
				return err
			}

			if WantsJSON(cmd) {
				return PrintJSON(squareCli.Out(), snapshots)
			}

			if len(snapshots) < 1 {
				fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.snapshot.list.empty"))
				return nil
			}

			w := tabwriter.NewWriter(squareCli.Out(), 0, 0, 2, ' ', tabwriter.TabIndent)
			defer w.Flush()

			// With an ID filter: one row per snapshot version.
			if len(args) > 0 {
				id := args[0]

				headers := []string{
					"ID",
					"VERSION ID",
					squareCli.I18n().T("commands.snapshot.list.table.size"),
					squareCli.I18n().T("commands.snapshot.list.table.last_update"),
				}
				fmt.Fprintln(w, strings.Join(headers, " \t "))

				for _, snap := range snapshots {
					if snap.Name != id {
						continue
					}

					var versionID string
					if values, err := url.ParseQuery(snap.Key); err == nil {
						versionID = values.Get("versionId")
					}

					row := []string{
						snap.Name,
						versionID,
						FormatBytes(int64(snap.Size)),
						snap.Modified.Format(time.DateTime),
					}
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
				s, ok := groups[snap.Name]
				if !ok {
					s = &summary{id: snap.Name}
					groups[snap.Name] = s
				}

				s.count++
				s.totalBytes += int64(snap.Size)
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
				squareCli.I18n().T("commands.snapshot.list.table.quantity"),
				squareCli.I18n().T("commands.snapshot.list.table.size"),
				squareCli.I18n().T("commands.snapshot.list.table.last_update"),
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

// DownloadFile fetches fileURL into destination — used by `snapshot create
// --download`.
func DownloadFile(destination, fileURL string) error {
	resp, err := http.Get(fileURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with HTTP %d", resp.StatusCode)
	}

	out, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
