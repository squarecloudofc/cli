package blob

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/squarecloudofc/cli/i18n"
	"github.com/squarecloudofc/cli/internal/blob"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/config"
	"github.com/squarecloudofc/sdk-api-go/v4"
)

type fakeCLI struct{ out, err bytes.Buffer }

func (f *fakeCLI) Config() *config.Config   { return &config.Config{} }
func (f *fakeCLI) API() *squarecloud.Client { return nil }
func (f *fakeCLI) I18n() i18n.Localizer     { return i18n.NewLocalizer("en") }
func (f *fakeCLI) Err() io.Writer           { return &f.err }
func (f *fakeCLI) In() io.ReadCloser        { return io.NopCloser(strings.NewReader("")) } // not a terminal
func (f *fakeCLI) Out() io.Writer           { return &f.out }

// run executes `blob args...` against a fake Blob service answering body
// to every request; it returns the request count.
func run(t *testing.T, body string, args ...string) (*fakeCLI, int, error) {
	t.Helper()
	t.Setenv(cli.APIKeyEnv, "test-key")
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, body)
	}))
	t.Cleanup(srv.Close)
	old := blob.BaseURL
	blob.BaseURL = srv.URL + "/v1"
	t.Cleanup(func() { blob.BaseURL = old })

	f := &fakeCLI{}
	cmd := NewCommand(f)
	cmd.PersistentFlags().Bool("json", false, "") // global on the real root
	cmd.SetArgs(args)
	cmd.SilenceErrors, cmd.SilenceUsage = true, true
	err := cmd.Execute()
	return f, calls, err
}

// Only the link goes to stdout, so $(squarecloud blob url <id>) works.
func TestURLPrintsOnlyTheLink(t *testing.T) {
	f, _, err := run(t, `{"status":"success","response":{"url":"https://blob.example/d/abc","expires_at":"2026-11-16T10:00:00Z","private":true}}`, "url", "u/a_h.png")
	if err != nil {
		t.Fatal(err)
	}
	if f.out.String() != "https://blob.example/d/abc\n" {
		t.Fatalf("stdout = %q", f.out.String())
	}
	if !strings.Contains(f.err.String(), "2026-11-16") {
		t.Fatalf("stderr = %q, want the expiry", f.err.String())
	}
}

func TestListJSONIsAnArray(t *testing.T) {
	f, _, err := run(t, `{"status":"success","response":{"objects":[]}}`, "list", "--json")
	if err != nil || strings.TrimSpace(f.out.String()) != "[]" {
		t.Fatalf("list --json: err %v, stdout %q", err, f.out.String())
	}
}

func TestDeleteAsksFirst(t *testing.T) {
	_, calls, err := run(t, `{"status":"success"}`, "delete", "u/a.png")
	if want := i18n.NewLocalizer("en").T("errors.confirm.no_tty"); err == nil || err.Error() != want {
		t.Fatalf("delete without -y = %v, want %q", err, want)
	}
	if calls != 0 {
		t.Fatalf("%d requests before the confirmation", calls)
	}
	if _, calls, err := run(t, `{"status":"success"}`, "delete", "u/a.png", "-y"); err != nil || calls != 1 {
		t.Fatalf("delete -y: err %v, %d requests", err, calls)
	}
}
