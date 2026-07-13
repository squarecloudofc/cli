// Package cmdutil holds small helpers shared by every CLI command: resource
// ID resolution, JSON output, confirmation prompts and byte formatting.
package cmdutil

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/ui/selector"
	"github.com/squarecloudofc/cli/pkg/squareconfig"
)

// ResolveAppID resolves the target application: positional argument first,
// then the ID stored in the local squarecloud.app config, then an
// interactive picker.
func ResolveAppID(squareCli cli.SquareCLI, args []string) (string, error) {
	if len(args) > 0 {
		return args[0], nil
	}

	if config, err := squareconfig.Load(); err == nil && config.ID != "" {
		return config.ID, nil
	}

	apps, err := squareCli.Rest().GetApplications()
	if err != nil {
		return "", err
	}
	if len(apps) == 0 {
		return "", fmt.Errorf("%s", squareCli.I18n().T("commands.app.list.empty"))
	}

	items := make([]selector.Item, len(apps))
	for i, app := range apps {
		items[i] = selector.Item{
			ID:    app.ID,
			Label: app.Name,
			Desc:  fmt.Sprintf("(%s - %s)", app.ID, app.Cluster),
		}
	}

	item, err := selector.Run(squareCli.I18n().T("ui.select.application"), items)
	if err != nil {
		return "", err
	}

	return item.ID, nil
}

// ResolveDatabaseID resolves the target database: positional argument first,
// then an interactive picker.
func ResolveDatabaseID(squareCli cli.SquareCLI, args []string) (string, error) {
	if len(args) > 0 {
		return args[0], nil
	}

	databases, err := squareCli.Rest().GetDatabases()
	if err != nil {
		return "", err
	}
	if len(databases) == 0 {
		return "", fmt.Errorf("%s", squareCli.I18n().T("commands.database.list.empty"))
	}

	items := make([]selector.Item, len(databases))
	for i, db := range databases {
		items[i] = selector.Item{
			ID:    db.ID,
			Label: db.Name,
			Desc:  fmt.Sprintf("(%s - %s - %s)", db.ID, db.Type, db.Cluster),
		}
	}

	item, err := selector.Run(squareCli.I18n().T("ui.select.database"), items)
	if err != nil {
		return "", err
	}

	return item.ID, nil
}

// WantsJSON reports whether the global --json flag is set.
func WantsJSON(cmd *cobra.Command) bool {
	b, _ := cmd.Flags().GetBool("json")
	return b
}

// PrintJSON writes v as indented JSON.
func PrintJSON(w io.Writer, v any) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}

// Confirm asks a y/N question unless assumeYes is set.
func Confirm(squareCli cli.SquareCLI, message string, assumeYes bool) bool {
	if assumeYes {
		return true
	}

	fmt.Fprintf(squareCli.Out(), "%s [y/N]: ", message)

	var response string
	_, _ = fmt.Fscanln(squareCli.In(), &response)

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

// FormatBytes renders a byte count in a compact human form.
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
