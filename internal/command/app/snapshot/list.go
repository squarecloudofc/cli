package snapshot

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/sdk-api-go/squarecloud"
)

func NewListCommand(squareCli cli.SquareCLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: squareCli.I18n().T("metadata.commands.app.list.short"),
		RunE:  runSnapshotListCommand(squareCli),
	}

	return cmd
}

type snapshotSummary struct {
	ID         string
	Count      int
	TotalBytes int64
	LastUpdate time.Time
	Snapshots  []squarecloud.Snapshot
}

func runSnapshotListCommand(squareCli cli.SquareCLI) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) (err error) {
		rest := squareCli.Rest()
		raw, err := rest.UserSnapshots("applications")
		if err != nil {
			return err
		}

		if len(raw) < 1 {
			fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.app.list.empty"))
			return err
		}

		if len(args) > 0 {
			id := args[0]

			w := tabwriter.NewWriter(squareCli.Out(), 0, 0, 2, ' ', tabwriter.TabIndent)
			defer w.Flush()

			headers := []string{
				"ID",
				"Version ID",
				squareCli.I18n().T("commands.app.snapshot.list.table.size"),
				squareCli.I18n().T("commands.app.snapshot.list.table.last_update"),
			}
			fmt.Fprintln(w, strings.Join(headers, " \t "))

			for _, snap := range raw {
				if snap.Name == id {

					var versionID string
					values, err := url.ParseQuery(snap.Key)
					if err == nil {
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
			}

			return nil
		}

		grouped := groupSnapshotsByName(raw)
		summaries := summarizeGroups(grouped)
		sort.Slice(summaries, func(i, j int) bool { return summaries[i].TotalBytes > summaries[j].TotalBytes })

		writeSummaryTable(squareCli, summaries)

		return nil
	}
}

func groupSnapshotsByName(list []squarecloud.Snapshot) map[string][]squarecloud.Snapshot {
	groups := make(map[string][]squarecloud.Snapshot, len(list))
	for _, s := range list {
		groups[s.Name] = append(groups[s.Name], s)
	}
	return groups
}

func summarizeGroups(groups map[string][]squarecloud.Snapshot) []snapshotSummary {
	summaries := make([]snapshotSummary, 0, len(groups))
	for id, snaps := range groups {
		var total int64
		var last time.Time

		for _, s := range snaps {
			total += int64(s.Size)
			if s.Modified.After(last) {
				last = s.Modified
			}
		}

		summaries = append(summaries, snapshotSummary{
			ID:         id,
			Count:      len(snaps),
			TotalBytes: total,
			LastUpdate: last,
			Snapshots:  snaps,
		})
	}
	return summaries
}

func writeSummaryTable(squareCli cli.SquareCLI, data []snapshotSummary) {
	w := tabwriter.NewWriter(squareCli.Out(), 0, 0, 2, ' ', tabwriter.TabIndent)
	defer w.Flush()

	headers := []string{
		"ID",
		squareCli.I18n().T("commands.app.snapshot.list.table.quantity"),
		squareCli.I18n().T("commands.app.snapshot.list.table.size"),
		squareCli.I18n().T("commands.app.snapshot.list.table.last_update"),
	}
	fmt.Fprintln(w, strings.Join(headers, " \t "))

	for _, s := range data {
		row := []string{
			s.ID,
			fmt.Sprintf("%d", s.Count),
			FormatBytes(s.TotalBytes),
			s.LastUpdate.Format(time.DateTime),
		}
		fmt.Fprintln(w, strings.Join(row, " \t "))
	}
}

func FormatBytes(size int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/float64(GB))
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/float64(MB))
	case size >= KB:
		return fmt.Sprintf("%d KB", size/KB)
	default:
		return fmt.Sprintf("%d Bytes", size)
	}
}
