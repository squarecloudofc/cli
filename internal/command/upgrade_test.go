package command

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/squarecloudofc/cli/i18n"
	"github.com/squarecloudofc/cli/internal/build"
	"github.com/squarecloudofc/cli/internal/updater"
)

// TestUpgradeCheck runs `upgrade` against a fake GitHub; none of these cases
// downloads or replaces anything.
func TestUpgradeCheck(t *testing.T) {
	en := i18n.NewLocalizer("en")
	status := http.StatusOK
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/releases/latest" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(`{"tag_name":"v3.2.0"}`))
	}))
	defer srv.Close()
	oldURL, oldVersion := updater.LatestReleaseURL, build.Version
	updater.LatestReleaseURL = srv.URL + "/releases/latest"
	t.Cleanup(func() { updater.LatestReleaseURL, build.Version = oldURL, oldVersion })

	run := func(version string, args ...string) (*fakeCLI, error) {
		build.Version = version
		f := &fakeCLI{}
		cmd := NewUpgradeCommand(f)
		cmd.SetArgs(args)
		cmd.SetOut(io.Discard)
		cmd.SetErr(io.Discard)
		return f, cmd.Execute()
	}

	for _, version := range []string{"v3.2.0", "v3.3.0", "development"} {
		f, err := run(version)
		want := en.T("commands.upgrade.up_to_date", map[string]any{"Version": version, "Latest": "v3.2.0"})
		if err != nil || !strings.Contains(f.out.String(), want) {
			t.Errorf("upgrade on %s = %v, %q; want %q", version, err, f.out.String(), want)
		}
	}

	f, err := run("v3.1.0", "--check")
	if err != nil {
		t.Fatalf("upgrade --check = %v", err)
	}
	for _, want := range []string{en.T("commands.upgrade.available", map[string]any{"Version": "v3.1.0", "Latest": "v3.2.0"}), en.T("commands.upgrade.run")} {
		if !strings.Contains(f.out.String(), want) {
			t.Errorf("upgrade --check: missing %q in %q", want, f.out.String())
		}
	}

	status = http.StatusForbidden // GitHub's rate limit
	if _, err := run("v3.1.0"); err == nil || !strings.Contains(err.Error(), "403") {
		t.Errorf("upgrade with GitHub refusing = %v, want the 403", err)
	}
}
