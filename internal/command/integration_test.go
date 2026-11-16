package command

import (
	"strings"
	"testing"

	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/integration"
)

// fakeRegistry stands in for HKEY_CURRENT_USER: key -> name -> data.
type fakeRegistry map[string]map[string]string

func (f fakeRegistry) SetString(key, name, data string) error {
	if f[key] == nil {
		f[key] = map[string]string{}
	}
	f[key][name] = data
	return nil
}

func (f fakeRegistry) DeleteKey(key string) error { delete(f, key); return nil }

func runIntegration(t *testing.T, reg integration.Registry, args ...string) (*fakeCLI, error) {
	t.Helper()
	old := integration.System
	integration.System = reg // never the real registry
	t.Cleanup(func() { integration.System = old })

	f := &fakeCLI{}
	cmd := NewIntegrationCommand(f)
	cmd.SetArgs(args)
	return f, cmd.Execute()
}

func TestIntegrationInstallUninstall(t *testing.T) {
	reg := fakeRegistry{}

	f, err := runIntegration(t, reg, "install", "--dry-run")
	if err != nil || len(reg) != 0 || !strings.Contains(f.out.String(), `integration send "%1"`) {
		t.Fatalf("install --dry-run: err %v, wrote %v, printed %q", err, reg, f.out.String())
	}

	if _, err := runIntegration(t, reg, "install"); err != nil {
		t.Fatal(err)
	}
	if len(reg) != len(integration.Keys()) {
		t.Fatalf("install wrote %d keys, want %d: %v", len(reg), len(integration.Keys()), reg)
	}

	f, err = runIntegration(t, reg, "uninstall", "--dry-run")
	if err != nil || len(reg) != len(integration.Keys()) || !strings.Contains(f.out.String(), `HKCU\Software\Classes\Directory\shell\SquareCloud`) {
		t.Fatalf("uninstall --dry-run: err %v, registry %v, printed %q", err, reg, f.out.String())
	}

	if _, err := runIntegration(t, reg, "uninstall"); err != nil || len(reg) != 0 {
		t.Fatalf("uninstall: err %v, left %v", err, reg)
	}
}

func TestIntegrationElsewhere(t *testing.T) {
	want := (&fakeCLI{}).I18n().T("commands.integration.unsupported")
	for _, sub := range []string{"install", "uninstall"} {
		if _, err := runIntegration(t, nil, sub); err == nil || err.Error() != want {
			t.Errorf("%s without a registry = %v, want %q", sub, err, want)
		}
	}
}

func TestSendNeedsLogin(t *testing.T) {
	t.Setenv(cli.APIKeyEnv, "")
	f := &fakeCLI{}
	err := sendToSquareCloud(t.Context(), f, t.TempDir())
	if want := f.I18n().T("errors.common.not_logged"); err == nil || err.Error() != want {
		t.Fatalf("send logged out = %v, want %q", err, want)
	}
}
