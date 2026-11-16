package cli

import (
	"os"
	"testing"

	"github.com/squarecloudofc/cli/i18n"
	"github.com/squarecloudofc/cli/internal/config"
	"github.com/zalando/go-keyring"
)

// TestMain swaps the OS keychain for go-keyring's in-memory fake: no test
// may read or write the real one.
func TestMain(m *testing.M) {
	keyring.MockInit()
	os.Exit(m.Run())
}

func TestUseProfile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("APPDATA", dir)
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv(APIKeyEnv, "")
	_ = keyring.Set(config.KeychainService, config.DefaultProfile, "default-key")
	_ = keyring.Set(config.KeychainService, "work", "work-key")

	s := &squarecliImpl{i18n: i18n.NewLocalizer("en")}
	if err := s.UseProfile(""); err != nil || APIKey(s.Config()) != "default-key" {
		t.Fatalf("default profile: err %v, key %q", err, APIKey(s.Config()))
	}
	if err := s.UseProfile("work"); err != nil || APIKey(s.Config()) != "work-key" {
		t.Fatalf("work profile: err %v, key %q", err, APIKey(s.Config()))
	}

	// SQUARECLOUD_API_KEY wins over every profile.
	t.Setenv(APIKeyEnv, "env-key")
	if got := APIKey(s.Config()); got != "env-key" {
		t.Fatalf("APIKey with the env set = %q", got)
	}

	want := s.i18n.T("errors.config.invalid_profile", map[string]any{"Profile": "../etc"})
	if err := s.UseProfile("../etc"); err == nil || err.Error() != want {
		t.Fatalf("UseProfile(../etc) = %v, want %q", err, want)
	}
}
