// Package cmdutil holds small helpers shared by every CLI command: resource
// ID resolution, JSON output, confirmation prompts, project zips and byte
// formatting.
package cmdutil

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/ui"
	"github.com/squarecloudofc/cli/internal/ui/selector"
	"github.com/squarecloudofc/cli/pkg/squareconfig"
	"github.com/squarecloudofc/cli/pkg/squareignore"
	"github.com/squarecloudofc/cli/pkg/zipper"
	"github.com/squarecloudofc/sdk-api-go/v3"
	"golang.org/x/term"
)

// ResolveAppID resolves the target application: positional argument first,
// then the ID stored in the local squarecloud.app config, then an
// interactive picker.
func ResolveAppID(ctx context.Context, squareCli cli.SquareCLI, args []string) (string, error) {
	if len(args) > 0 {
		return args[0], nil
	}

	if config, err := squareconfig.Load(); err == nil && config.ID != "" {
		return config.ID, nil
	}

	me, err := squareCli.API().Account.Me(ctx)
	if err != nil {
		return "", err
	}
	apps := me.Applications
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
func ResolveDatabaseID(ctx context.Context, squareCli cli.SquareCLI, args []string) (string, error) {
	if len(args) > 0 {
		return args[0], nil
	}

	me, err := squareCli.API().Account.Me(ctx)
	if err != nil {
		return "", err
	}
	databases := me.Databases
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

// TableWriter lines up tab-separated cells like text/tabwriter with two
// spaces of padding (the text after a line's last tab is not a column), but
// measures cells in terminal columns: tabwriter counts runes, which pushes
// CJK headers and names out of line.
type TableWriter struct {
	out io.Writer
	buf strings.Builder
}

func NewTableWriter(out io.Writer) *TableWriter { return &TableWriter{out: out} }

func (t *TableWriter) Write(p []byte) (int, error) { return t.buf.Write(p) }

// Flush writes the buffered lines, aligned.
func (t *TableWriter) Flush() error {
	if t.buf.Len() == 0 {
		return nil
	}
	lines := strings.Split(strings.TrimSuffix(t.buf.String(), "\n"), "\n")
	t.buf.Reset()

	var widths []int
	rows := make([][]string, len(lines))
	for i, line := range lines {
		rows[i] = strings.Split(line, "\t")
		for j, cell := range rows[i][:len(rows[i])-1] {
			if j == len(widths) {
				widths = append(widths, 0)
			}
			widths[j] = max(widths[j], lipgloss.Width(cell))
		}
	}

	var b strings.Builder
	for _, cells := range rows {
		for j, cell := range cells {
			b.WriteString(cell)
			if j < len(cells)-1 {
				b.WriteString(strings.Repeat(" ", widths[j]-lipgloss.Width(cell)+2))
			}
		}
		b.WriteByte('\n')
	}
	_, err := io.WriteString(t.out, b.String())
	return err
}

// PrintJSON writes v as indented JSON.
func PrintJSON(w io.Writer, v any) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}

// isTerminal reports whether r is an interactive terminal. A variable so
// tests can stand in for one.
var isTerminal = func(r io.Reader) bool {
	f, ok := r.(*os.File)
	return ok && term.IsTerminal(int(f.Fd()))
}

// Confirm asks a y/N question on stderr unless assumeYes is set. Anything
// but a yes is an error, so a skipped action never exits 0: an answered no
// says nothing changed, and a question that cannot be asked (stdin is not a
// terminal, or it hit EOF) says to pass --yes.
func Confirm(squareCli cli.SquareCLI, message string, assumeYes bool) error {
	if assumeYes {
		return nil
	}

	noTerminal := errors.New(squareCli.I18n().T("errors.confirm.no_tty"))
	if !isTerminal(squareCli.In()) {
		return noTerminal
	}

	fmt.Fprintf(squareCli.Err(), "%s [y/N]: ", message)

	response, err := bufio.NewReader(squareCli.In()).ReadString('\n')
	if err != nil && response == "" {
		fmt.Fprintln(squareCli.Err()) // Ctrl+D leaves the cursor after the prompt
		return noTerminal
	}

	switch strings.ToLower(strings.TrimSpace(response)) {
	case "y", "yes":
		return nil
	}
	return errors.New(squareCli.I18n().T("errors.confirm.declined"))
}

// ZipProject zips the project in dir into a temp file (see zipper.ZipFolder),
// leaving out what squarecloud.ignore lists. Notes about a legacy ignore
// file name and skipped links go to stderr.
func ZipProject(squareCli cli.SquareCLI, dir string) (*os.File, error) {
	ignore, legacy, err := squareignore.Load(dir)
	if err != nil {
		return nil, err
	}
	if legacy != "" {
		fmt.Fprintln(squareCli.Err(), ui.TextYellow.Render(squareCli.I18n().T("commands.zip.legacy_ignore", map[string]any{"File": legacy})))
	}

	return zipper.ZipFolder(dir, ignore, func(rel string) {
		fmt.Fprintln(squareCli.Err(), squareCli.I18n().T("commands.zip.skipped_link", map[string]any{"Path": rel}))
	})
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

// AlreadyInState handles the API's 409 answer to starting a running
// container or stopping a stopped one: it prints an informational line and
// reports true, so the caller can treat it as success.
func AlreadyInState(squareCli cli.SquareCLI, err error) bool {
	var apiErr *squarecloud.APIError
	if !errors.As(err, &apiErr) {
		return false
	}

	info := ui.TextBlue.Render("i")
	switch apiErr.Code {
	case squarecloud.CodeContainerAlreadyStarted:
		fmt.Fprintf(squareCli.Out(), "%s %s\n", info, squareCli.I18n().T("commands.signal.already_running"))
	case squarecloud.CodeContainerAlreadyStopped:
		fmt.Fprintf(squareCli.Out(), "%s %s\n", info, squareCli.I18n().T("commands.signal.already_stopped"))
	default:
		return false
	}
	return true
}
