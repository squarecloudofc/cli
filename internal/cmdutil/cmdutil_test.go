package cmdutil

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"text/tabwriter"

	"github.com/charmbracelet/lipgloss"
	"github.com/squarecloudofc/cli/i18n"
	"github.com/squarecloudofc/cli/internal/config"
	"github.com/squarecloudofc/sdk-api-go/v3"
)

type fakeCLI struct {
	in       io.Reader
	api      *squarecloud.Client
	out, err bytes.Buffer
}

func (f *fakeCLI) Config() *config.Config   { return &config.Config{} }
func (f *fakeCLI) API() *squarecloud.Client { return f.api }
func (f *fakeCLI) I18n() i18n.Localizer     { return i18n.NewLocalizer("en") }
func (f *fakeCLI) Err() io.Writer           { return &f.err }
func (f *fakeCLI) In() io.ReadCloser        { return io.NopCloser(f.in) }
func (f *fakeCLI) Out() io.Writer           { return &f.out }

func TestConfirm(t *testing.T) {
	en := i18n.NewLocalizer("en")
	noTTY, declined := en.T("errors.confirm.no_tty"), en.T("errors.confirm.declined")

	cases := []struct {
		name       string
		tty, yes   bool
		input      string
		wantErr    string // "" = confirmed
		wantPrompt bool
	}{
		{name: "yes flag skips the question", yes: true, input: "n\n"},
		{name: "yes flag without a terminal", yes: true},
		{name: "no terminal", input: "y\n", wantErr: noTTY},
		{name: "answered y", tty: true, input: "y\n", wantPrompt: true},
		{name: "answered YES", tty: true, input: " YES \r\n", wantPrompt: true},
		{name: "answered n", tty: true, input: "n\n", wantErr: declined, wantPrompt: true},
		{name: "enter defaults to no", tty: true, input: "\n", wantErr: declined, wantPrompt: true},
		{name: "EOF", tty: true, input: "", wantErr: noTTY, wantPrompt: true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			isTerminal = func(io.Reader) bool { return c.tty }
			t.Cleanup(func() { isTerminal = realIsTerminal })

			f := &fakeCLI{in: strings.NewReader(c.input)}
			err := Confirm(f, "Delete it?", c.yes)

			switch {
			case c.wantErr == "" && err != nil:
				t.Fatalf("Confirm = %v, want nil", err)
			case c.wantErr != "" && (err == nil || err.Error() != c.wantErr):
				t.Fatalf("Confirm = %v, want %q", err, c.wantErr)
			}
			if got := strings.Contains(f.err.String(), "Delete it? [y/N]"); got != c.wantPrompt {
				t.Fatalf("prompt on stderr = %v, want %v (stderr %q)", got, c.wantPrompt, f.err.String())
			}
			if f.out.Len() != 0 {
				t.Fatalf("stdout = %q, want nothing", f.out.String())
			}
		})
	}
}

var realIsTerminal = isTerminal

func TestZipProjectWarnsAboutLegacyIgnoreFile(t *testing.T) {
	dir := t.TempDir()
	for name, content := range map[string]string{".squareignore": "secret.txt\n", "secret.txt": "x", "index.js": "x"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	f := &fakeCLI{}
	file, err := ZipProject(f, dir)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())
	defer file.Close()

	info, _ := file.Stat()
	r, err := zip.NewReader(file, info.Size())
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range r.File {
		names = append(names, e.Name)
	}
	slices.Sort(names)
	if want := []string{".squareignore", "index.js"}; !slices.Equal(names, want) {
		t.Errorf("entries = %q, want %q", names, want)
	}

	want := f.I18n().T("commands.zip.legacy_ignore", map[string]any{"File": ".squareignore"})
	if !strings.Contains(f.err.String(), want) || f.out.Len() != 0 {
		t.Errorf("stderr = %q, stdout = %q; want the rename warning on stderr only", f.err.String(), f.out.String())
	}
}

func TestIsTerminalRejectsFilesAndReaders(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "stdin")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	if realIsTerminal(f) || realIsTerminal(strings.NewReader("y\n")) {
		t.Fatal("a regular file or a reader is not a terminal")
	}
}

// TableWriter must print what text/tabwriter printed for the CLI's tables,
// and keep CJK columns aligned where tabwriter (counting runes) does not.
func TestTableWriter(t *testing.T) {
	write := func(w io.Writer) {
		fmt.Fprintln(w, strings.Join([]string{"  NAME", "APP ID", "MEMORY"}, " \t "))
		fmt.Fprintf(w, "  %s \t %s \t %s \t\n", "api", "a1b2c3", "512MB")
		fmt.Fprintf(w, "  %s \t %s \t %s \t\n", "a longer name", "", "1024MB")
	}

	var want, got bytes.Buffer
	tw := tabwriter.NewWriter(&want, 0, 0, 2, ' ', tabwriter.TabIndent)
	write(tw)
	tw.Flush()
	w := NewTableWriter(&got)
	write(w)
	if err := w.Flush(); err != nil || got.String() != want.String() {
		t.Fatalf("TableWriter printed\n%s\ntabwriter printed\n%s", got.String(), want.String())
	}

	got.Reset()
	w = NewTableWriter(&got)
	fmt.Fprintln(w, "名称 \t ID")
	fmt.Fprintln(w, "api \t a1")
	w.Flush()
	lines := strings.Split(got.String(), "\n")
	column := func(line, s string) int { return lipgloss.Width(line[:strings.Index(line, s)]) }
	if column(lines[0], "ID") != column(lines[1], "a1") {
		t.Fatalf("CJK header out of line:\n%s", got.String())
	}
}
