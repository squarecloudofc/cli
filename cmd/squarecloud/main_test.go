package main

import (
	"bytes"
	"cmp"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/squarecloudofc/cli/i18n"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/config"
	"github.com/squarecloudofc/sdk-api-go/v3"
)

type fakeCLI struct {
	cfg      *config.Config
	lang     string // "" = en
	out, err bytes.Buffer
}

func (f *fakeCLI) Config() *config.Config   { return f.cfg }
func (f *fakeCLI) API() *squarecloud.Client { return cli.NewClient(cli.APIKey(f.cfg)) }
func (f *fakeCLI) I18n() i18n.Localizer     { return i18n.NewLocalizer(cmp.Or(f.lang, "en")) }
func (f *fakeCLI) Err() io.Writer           { return &f.err }
func (f *fakeCLI) In() io.ReadCloser        { return io.NopCloser(strings.NewReader("")) }
func (f *fakeCLI) Out() io.Writer           { return &f.out }

// run executes the real command tree logged out, with any API call failing
// the test.
func run(t *testing.T, args ...string) (*fakeCLI, error) {
	t.Helper()
	return runIn(t, "en", args...)
}

func runIn(t *testing.T, lang string, args ...string) (*fakeCLI, error) {
	t.Helper()
	t.Setenv(cli.APIKeyEnv, "")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("unexpected API call %s %s", r.Method, r.URL.Path)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)
	old := cli.BaseURL
	cli.BaseURL = srv.URL
	t.Cleanup(func() { cli.BaseURL = old })

	f := &fakeCLI{cfg: config.New(filepath.Join(t.TempDir(), config.ConfigFileName)), lang: lang}
	cmd := newSquareCloudCommand(f)
	cmd.SetArgs(args)
	return f, cmd.Execute()
}

func TestBareGroupsShowHelpLoggedOut(t *testing.T) {
	for _, group := range []string{"app", "auth", "db", "workspace"} {
		f, err := run(t, group)
		if err != nil {
			t.Errorf("squarecloud %s = %v, want its help", group, err)
		}
		if !strings.Contains(f.out.String(), "Usage:") {
			t.Errorf("squarecloud %s printed %q, want its help", group, f.out.String())
		}
	}

	// A subcommand still needs a login.
	if _, err := run(t, "app", "list"); !errors.As(err, new(*cli.AuthError)) {
		t.Errorf("app list logged out = %v, want an AuthError", err)
	}
}

func TestLogoutLoggedOut(t *testing.T) {
	f, err := run(t, "auth", "logout")
	if err != nil || !strings.Contains(f.out.String(), f.I18n().T("commands.auth.logout.none")) {
		t.Fatalf("auth logout = %v, stdout %q", err, f.out.String())
	}
}

func TestWantsUpdateCheck(t *testing.T) {
	for _, args := range [][]string{
		{"__complete", "app", ""}, {"__completeNoDesc", "a"}, {"completion", "bash"},
		{"--version"}, {"-v"}, {"help"}, {"app", "--help"}, {"app", "list", "-h"},
	} {
		if wantsUpdateCheck(args) {
			t.Errorf("wantsUpdateCheck(%q) = true, want false", args)
		}
	}
	if !wantsUpdateCheck([]string{"app", "list"}) || !wantsUpdateCheck(nil) {
		t.Error("a normal run should check for updates")
	}
}

func TestJSONRejectedWhereUnsupported(t *testing.T) {
	en := i18n.NewLocalizer("en")
	for path, args := range map[string][]string{
		"squarecloud app start":    {"app", "start", "--json"},
		"squarecloud zip":          {"--json", "zip"},
		"squarecloud app realtime": {"app", "realtime", "--json"},
	} {
		_, err := run(t, args...)
		want := en.T("errors.common.json_unsupported", map[string]any{"Command": path})
		if err == nil || err.Error() != want {
			t.Errorf("%v = %v, want %q", args, err, want)
		}
	}

	// A command with JSON output gets past the check (to the login check).
	if _, err := run(t, "app", "list", "--json"); !errors.As(err, new(*cli.AuthError)) {
		t.Errorf("app list --json = %v, want the login check", err)
	}
}

// TestHelpIsTranslated: cobra's own help text (headings, -h, the help
// command) follows the language like the rest of the CLI.
func TestHelpIsTranslated(t *testing.T) {
	pt := i18n.NewLocalizer("pt")
	for args, keys := range map[string][]string{
		"--help": {"metadata.commands.root.short", "metadata.help.usage", "metadata.help.available_commands",
			"metadata.help.command", "metadata.help.flags", "metadata.help.flag", "metadata.flags.json", "metadata.flags.version"},
		"app env set --help": {"metadata.commands.app.env.set.short", "metadata.help.flags", "metadata.help.global_flags",
			"metadata.help.flag", "metadata.flags.app", "metadata.flags.from_file"},
		"help app": {"metadata.commands.app.root.short", "metadata.help.available_commands"},
	} {
		f, err := runIn(t, "pt", strings.Fields(args)...)
		if err != nil {
			t.Fatalf("%s = %v", args, err)
		}
		out := f.out.String()
		for _, key := range keys {
			if !strings.Contains(out, pt.T(key)) {
				t.Errorf("%s: missing %s (%q) in\n%s", args, key, pt.T(key), out)
			}
		}
		for _, english := range []string{"Usage:", "Flags:", "Available Commands:", "help for", "for more information", "Help about"} {
			if strings.Contains(out, english) {
				t.Errorf("%s: English %q left in\n%s", args, english, out)
			}
		}
	}

	f, _ := runIn(t, "pt", "app", "--help")
	if want := pt.T("metadata.help.more_info", map[string]any{"Command": "squarecloud app"}); !strings.Contains(f.out.String(), want) {
		t.Errorf("app --help: missing %q in\n%s", want, f.out.String())
	}
}
