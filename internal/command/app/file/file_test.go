package file

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/squarecloudofc/cli/i18n"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/config"
	"github.com/squarecloudofc/sdk-api-go/v3"
)

type fakeCLI struct{ out, err bytes.Buffer }

func (f *fakeCLI) Config() *config.Config   { return &config.Config{} }
func (f *fakeCLI) API() *squarecloud.Client { return cli.NewClient("test-key") }
func (f *fakeCLI) I18n() i18n.Localizer     { return i18n.NewLocalizer("en") }
func (f *fakeCLI) Err() io.Writer           { return &f.err }
func (f *fakeCLI) In() io.ReadCloser        { return io.NopCloser(strings.NewReader("")) } // not a terminal
func (f *fakeCLI) Out() io.Writer           { return &f.out }

// fakeAPI serves "remote" as the content of every file and counts the reads.
func fakeAPI(t *testing.T) *int {
	t.Helper()
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Path != "/apps/app1/files/content" {
			t.Errorf("unexpected API call %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"status":"success","response":{"data":"cmVtb3Rl"}}`) // base64 of "remote"
	}))
	t.Cleanup(srv.Close)

	old := cli.BaseURL
	cli.BaseURL = srv.URL
	t.Cleanup(func() { cli.BaseURL = old })
	return &calls
}

func runRead(args ...string) error {
	cmd := newReadCommand(&fakeCLI{})
	cmd.SetArgs(append([]string{"/config.json", "--app", "app1"}, args...))
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	return cmd.Execute()
}

func TestReadOutput(t *testing.T) {
	calls := fakeAPI(t)
	dir := t.TempDir()

	// A new file is written without a question, readable by its owner only.
	created := filepath.Join(dir, "new.json")
	if err := runRead("-o", created); err != nil {
		t.Fatalf("read -o new file = %v", err)
	}
	assertFile(t, created, "remote", 0600)

	// An existing file needs a yes; with no terminal to ask, nothing is
	// downloaded or written.
	existing := filepath.Join(dir, "existing.json")
	if err := os.WriteFile(existing, []byte("local"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(existing, 0644); err != nil { // whatever the umask
		t.Fatal(err)
	}
	err := runRead("-o", existing)
	if want := i18n.NewLocalizer("en").T("errors.confirm.no_tty"); err == nil || err.Error() != want {
		t.Fatalf("read -o existing file = %v, want %q", err, want)
	}
	assertFile(t, existing, "local", 0644)
	if *calls != 1 {
		t.Fatalf("API calls = %d, want 1 (no download before the answer)", *calls)
	}

	// --yes overwrites and keeps the file's mode.
	if err := runRead("-o", existing, "--yes"); err != nil {
		t.Fatalf("read -o existing --yes = %v", err)
	}
	assertFile(t, existing, "remote", 0644)
}

func assertFile(t *testing.T, path, content string, perm os.FileMode) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil || string(data) != content {
		t.Fatalf("%s = %q, %v; want %q", path, data, err, content)
	}
	if runtime.GOOS == "windows" { // no Unix permission bits there
		return
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != perm {
		t.Fatalf("%s mode = %o, want %o", path, got, perm)
	}
}
