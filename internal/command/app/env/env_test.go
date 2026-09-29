package env

import (
	"bytes"
	"io"
	"maps"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
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

// TestNothingDeletesEveryVariableByAccident: replace with an empty set is
// refused, and remove clears everything only with an explicit --all.
func TestNothingDeletesEveryVariableByAccident(t *testing.T) {
	var requests []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		requests = append(requests, r.Method+" "+r.URL.Path+" "+strings.TrimSpace(string(body)))
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":"success","response":{}}`)
	}))
	defer srv.Close()
	old := cli.BaseURL
	cli.BaseURL = srv.URL
	defer func() { cli.BaseURL = old }()

	emptyFile := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(emptyFile, []byte("# nothing yet\n"), 0600); err != nil {
		t.Fatal(err)
	}

	en := i18n.NewLocalizer("en")
	replaceEmpty, usage, noTTY := en.T("commands.app.env.replace.empty"), en.T("commands.app.env.remove.usage"), en.T("errors.confirm.no_tty")
	for _, c := range []struct {
		args    []string
		wantErr string
	}{
		{[]string{"replace", "--app", "app1"}, replaceEmpty},
		{[]string{"replace", "--app", "app1", "--yes"}, replaceEmpty},
		{[]string{"replace", "--app", "app1", "--from-file", emptyFile, "--yes"}, replaceEmpty},
		{[]string{"remove", "--app", "app1"}, usage},
		{[]string{"remove", "--app", "app1", "--all", "KEY"}, usage},
		{[]string{"remove", "--app", "app1", "--all"}, noTTY},
	} {
		cmd := NewCommand(&fakeCLI{})
		cmd.SetArgs(c.args)
		cmd.SetOut(io.Discard)
		cmd.SetErr(io.Discard)
		if err := cmd.Execute(); err == nil || err.Error() != c.wantErr {
			t.Errorf("env %v = %v, want %q", c.args, err, c.wantErr)
		}
	}
	if len(requests) != 0 {
		t.Fatalf("API requests %q, want none", requests)
	}

	f := &fakeCLI{}
	cmd := NewCommand(f)
	cmd.SetArgs([]string{"remove", "--app", "app1", "--all", "--yes"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("env remove --all --yes = %v", err)
	}
	if want := []string{`PUT /apps/app1/envs {"envs":{}}`}; !slices.Equal(requests, want) {
		t.Fatalf("API requests %q, want %q", requests, want)
	}
	if !strings.Contains(f.out.String(), en.T("commands.app.env.remove.success_all")) {
		t.Fatalf("stdout = %q", f.out.String())
	}
}

func TestParseDotEnv(t *testing.T) {
	content := "\xef\xbb\xbf" + `FIRST=1
# a comment
export EXPORTED=yes
SPACED = padded value
INLINE=value # a comment
TABBED=value	# a comment
HASH=http://x/#frag
DOUBLE="quoted value"
SINGLE='it''s'
EMPTY_QUOTES=""
QUOTED_COMMENT="a # b" # trailing comment
UNPAIRED="open
MIXED="a'
INNER="a"b"
EMPTY=
no equals sign
` + "CRLF=1\r\n"

	envs := squarecloud.EnvVars{}
	parseDotEnv(content, envs)

	want := squarecloud.EnvVars{
		"FIRST":          "1",
		"EXPORTED":       "yes",
		"SPACED":         "padded value",
		"INLINE":         "value",
		"TABBED":         "value",
		"HASH":           "http://x/#frag",
		"DOUBLE":         "quoted value",
		"SINGLE":         `'it''s'`,
		"EMPTY_QUOTES":   "",
		"QUOTED_COMMENT": "a # b",
		"UNPAIRED":       `"open`,
		"MIXED":          `"a'`,
		"INNER":          `"a"b"`,
		"EMPTY":          "",
		"CRLF":           "1",
	}
	if !maps.Equal(envs, want) {
		for k, v := range want {
			if envs[k] != v {
				t.Errorf("%s = %q, want %q", k, envs[k], v)
			}
		}
		for k := range envs {
			if _, ok := want[k]; !ok {
				t.Errorf("unexpected key %q", k)
			}
		}
	}
}
