package auth

import (
	"bytes"
	"context"
	"encoding/json"
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
	out, err bytes.Buffer
}

func (f *fakeCLI) Config() *config.Config   { return f.cfg }
func (f *fakeCLI) API() *squarecloud.Client { return cli.NewClient(cli.APIKey(f.cfg)) }
func (f *fakeCLI) I18n() i18n.Localizer     { return i18n.NewLocalizer("en") }
func (f *fakeCLI) Err() io.Writer           { return &f.err }
func (f *fakeCLI) In() io.ReadCloser        { return io.NopCloser(strings.NewReader("")) } // not a terminal
func (f *fakeCLI) Out() io.Writer           { return &f.out }

// fakeAPI answers GET /users/me for two keys of two accounts and counts
// requests. cli.BaseURL points at it for the test.
func fakeAPI(t *testing.T) *int {
	t.Helper()
	t.Setenv(cli.APIKeyEnv, "")
	accounts := map[string]string{"key-a": "acc-a", "key-b": "acc-b"}

	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		id, ok := accounts[r.Header.Get("Authorization")]
		if r.URL.Path != "/users/me" || !ok {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "error", "code": "ACCESS_DENIED"})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "response": map[string]any{
			"user": map[string]any{"id": id, "name": id, "email": id + "@example.com", "plan": map[string]any{"name": "standard"}},
		}})
	}))
	t.Cleanup(srv.Close)

	old := cli.BaseURL
	cli.BaseURL = srv.URL
	t.Cleanup(func() { cli.BaseURL = old })
	return &calls
}

func runLogin(t *testing.T, cfg *config.Config, args ...string) (*fakeCLI, error) {
	t.Helper()
	f := &fakeCLI{cfg: cfg}
	cmd := NewLoginCommand(f)
	cmd.SetArgs(args)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	return f, cmd.ExecuteContext(context.Background())
}

func savedConfig(t *testing.T, key, accountID string) *config.Config {
	t.Helper()
	cfg := config.New(filepath.Join(t.TempDir(), config.ConfigFileName))
	cfg.SetKey(key, accountID, nil, nil)
	return cfg
}

func TestLoginAccountSwitch(t *testing.T) {
	en := i18n.NewLocalizer("en")

	cases := []struct {
		name      string
		savedKey  string
		savedID   string // "" = a config saved before account IDs were stored
		args      []string
		wantErr   string
		wantKey   string
		wantCalls int
	}{
		{name: "first login", args: []string{"--token", "key-b"}, wantKey: "key-b", wantCalls: 1},
		{name: "same account", savedKey: "key-b", savedID: "acc-b", args: []string{"--token", "key-b"}, wantKey: "key-b", wantCalls: 1},
		{name: "other account, no terminal", savedKey: "key-a", savedID: "acc-a", args: []string{"--token", "key-b"}, wantErr: en.T("errors.confirm.no_tty"), wantKey: "key-a", wantCalls: 1},
		{name: "other account, --yes", savedKey: "key-a", savedID: "acc-a", args: []string{"--token", "key-b", "--yes"}, wantKey: "key-b", wantCalls: 1},
		{name: "legacy config, other account", savedKey: "key-a", args: []string{"--token", "key-b"}, wantErr: en.T("errors.confirm.no_tty"), wantKey: "key-a", wantCalls: 2},
		{name: "legacy config, same account", savedKey: "key-b", args: []string{"--token", "key-b"}, wantKey: "key-b", wantCalls: 2},
		{name: "legacy config, dead key", savedKey: "revoked", args: []string{"--token", "key-b"}, wantKey: "key-b", wantCalls: 2},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			calls := fakeAPI(t)
			cfg := savedConfig(t, c.savedKey, c.savedID)

			_, err := runLogin(t, cfg, c.args...)
			switch {
			case c.wantErr == "" && err != nil:
				t.Fatalf("login = %v, want success", err)
			case c.wantErr != "" && (err == nil || err.Error() != c.wantErr):
				t.Fatalf("login = %v, want %q", err, c.wantErr)
			}
			if cfg.AuthToken != c.wantKey {
				t.Fatalf("saved key = %q, want %q", cfg.AuthToken, c.wantKey)
			}
			if c.wantErr == "" && cfg.AccountID != "acc-b" {
				t.Fatalf("saved account = %q, want acc-b", cfg.AccountID)
			}
			if *calls != c.wantCalls {
				t.Fatalf("API calls = %d, want %d", *calls, c.wantCalls)
			}
		})
	}
}

func TestLoginRefusesBrowserLoginInCI(t *testing.T) {
	calls := fakeAPI(t)
	t.Setenv("CI", "true")

	_, err := runLogin(t, savedConfig(t, "", ""))
	want := i18n.NewLocalizer("en").T("commands.auth.login.ci", map[string]any{"Link": "https://squarecloud.app/account/security"})
	if err == nil || err.Error() != want {
		t.Fatalf("login = %v, want %q", err, want)
	}
	if *calls != 0 {
		t.Fatalf("API calls = %d, want none (no key may be created)", *calls)
	}

	// A dashboard key passed explicitly is still accepted in CI.
	if _, err := runLogin(t, savedConfig(t, "", ""), "--token", "key-a"); err != nil {
		t.Fatalf("login --token in CI = %v", err)
	}
}

func TestLogoutWhenLoggedOut(t *testing.T) {
	f := &fakeCLI{cfg: savedConfig(t, "", "")}
	cmd := NewLogoutCommand(f)
	cmd.SetArgs(nil)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("logout = %v, want nil", err)
	}
	if want := f.I18n().T("commands.auth.logout.none"); !strings.Contains(f.out.String(), want) {
		t.Fatalf("stdout = %q, want %q", f.out.String(), want)
	}
}
