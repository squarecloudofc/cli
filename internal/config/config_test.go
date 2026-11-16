package config

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/zalando/go-keyring"
)

// TestMain swaps the OS keychain for go-keyring's in-memory fake: no test
// may read or write the real one.
func TestMain(m *testing.M) {
	keyring.MockInit()
	os.Exit(m.Run())
}

// noKeychain makes every keychain call fail, as in a container, for one test.
func noKeychain(t *testing.T) {
	keyring.MockInitWithError(errors.New("no keychain"))
	t.Cleanup(keyring.MockInit)
}

func tempConfigDir(t *testing.T) string {
	dir := t.TempDir()
	t.Setenv("APPDATA", dir)
	t.Setenv("XDG_CONFIG_HOME", dir)
	return ConfigDir()
}

func TestSaveReplacesAtomically(t *testing.T) {
	noKeychain(t) // the key goes to the file
	dir := filepath.Join(t.TempDir(), "Square Cloud CLI")
	path := filepath.Join(dir, ConfigFileName)

	c := New(path)
	c.SetKey("first-key", "u1", nil, nil)
	if err := c.Save(); err != nil {
		t.Fatal(err)
	}
	c.SetKey("second-key", "u2", nil, nil)
	if err := c.Save(); err != nil {
		t.Fatal(err)
	}
	if !c.KeyInFile() {
		t.Fatal("KeyInFile = false without a keychain")
	}

	loaded := New(path)
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := loaded.LoadFromReader(f); err != nil {
		t.Fatal(err)
	}
	if loaded.AuthToken != "second-key" || loaded.AccountID != "u2" {
		t.Fatalf("loaded %+v, want the second save", loaded)
	}

	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Fatalf("folder holds %d entries, want only %s (temp files left behind?)", len(entries), ConfigFileName)
	}

	if runtime.GOOS != "windows" { // no Unix permission bits there
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if perm := info.Mode().Perm(); perm != 0600 {
			t.Fatalf("permissions = %o, want 600", perm)
		}
	}
}

func TestNoConfigDir(t *testing.T) {
	for _, env := range []string{"APPDATA", "XDG_CONFIG_HOME", "HOME", "USERPROFILE"} {
		t.Setenv(env, "")
	}

	if dir := ConfigDir(); dir != "" {
		t.Fatalf("ConfigDir = %q, want \"\"", dir)
	}

	c, err := Load("") // must not panic nor read ./config.json
	if err != nil {
		t.Fatalf("Load = %v", err)
	}
	if err := c.Save(); !errors.Is(err, ErrNoConfigDir) {
		t.Fatalf("Save = %v, want ErrNoConfigDir", err)
	}
}

// The language is detected on every run: Load must not write the config to
// store one, and a "locale" saved by older versions is dropped.
func TestLocaleIsNotStored(t *testing.T) {
	noKeychain(t)
	path := filepath.Join(tempConfigDir(t), ConfigFileName)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	old := `{"locale": "en", "auth_token": "k"}`
	if err := os.WriteFile(path, []byte(old), 0600); err != nil {
		t.Fatal(err)
	}

	c, err := Load("")
	if err != nil {
		t.Fatal(err)
	}
	if c.Key() != "k" {
		t.Fatalf("Key = %q, want k", c.Key())
	}
	if data, _ := os.ReadFile(path); string(data) != old {
		t.Fatalf("Load rewrote the config: %s", data)
	}

	if err := c.Save(); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	if strings.Contains(string(data), "locale") || !strings.Contains(string(data), `"auth_token": "k"`) {
		t.Fatalf("saved %s, want the key and no locale", data)
	}
}

func TestKeyGoesToTheKeychain(t *testing.T) {
	keyring.MockInit()
	path := filepath.Join(tempConfigDir(t), ConfigFileName)

	c, _ := Load("")
	c.SetKey("secret-key", "acc", nil, []string{"apps:read"})
	if err := c.Save(); err != nil {
		t.Fatal(err)
	}
	if c.KeyInFile() {
		t.Fatal("KeyInFile = true with a keychain")
	}
	data, _ := os.ReadFile(path)
	if strings.Contains(string(data), "secret-key") || !strings.Contains(string(data), `"account_id": "acc"`) {
		t.Fatalf("file = %s, want the account and no key", data)
	}
	if got, _ := keyring.Get(KeychainService, DefaultProfile); got != "secret-key" {
		t.Fatalf("keychain = %q, want secret-key", got)
	}

	reloaded, _ := Load("")
	if reloaded.Key() != "secret-key" {
		t.Fatalf("Key after reload = %q", reloaded.Key())
	}

	// Logout clears the keychain and the file.
	reloaded.SetKey("", "", nil, nil)
	if err := reloaded.Save(); err != nil {
		t.Fatal(err)
	}
	if _, err := keyring.Get(KeychainService, DefaultProfile); !errors.Is(err, keyring.ErrNotFound) {
		t.Fatalf("keychain after logout: %v, want ErrNotFound", err)
	}
	if again, _ := Load(""); again.Key() != "" {
		t.Fatalf("Key after logout = %q", again.Key())
	}
}

// A key left in the file (older versions, or no keychain at login) moves to
// the keychain the first time it is read.
func TestFileKeyMigrates(t *testing.T) {
	keyring.MockInit()
	path := filepath.Join(tempConfigDir(t), ConfigFileName)
	_ = os.MkdirAll(filepath.Dir(path), 0700)
	if err := os.WriteFile(path, []byte(`{"auth_token": "legacy", "account_id": "acc"}`), 0600); err != nil {
		t.Fatal(err)
	}

	c, _ := Load("")
	if c.Key() != "legacy" {
		t.Fatalf("Key = %q, want legacy", c.Key())
	}
	if got, _ := keyring.Get(KeychainService, DefaultProfile); got != "legacy" {
		t.Fatalf("keychain = %q, want legacy", got)
	}
	data, _ := os.ReadFile(path)
	if strings.Contains(string(data), "legacy") || !strings.Contains(string(data), `"account_id": "acc"`) {
		t.Fatalf("file after migration = %s", data)
	}
}

func TestProfiles(t *testing.T) {
	keyring.MockInit()
	dir := tempConfigDir(t)

	work, err := Load("work")
	if err != nil {
		t.Fatal(err)
	}
	if work.Path() != filepath.Join(dir, "profiles", "work.json") || work.Profile() != "work" {
		t.Fatalf("work profile at %s (%s)", work.Path(), work.Profile())
	}
	work.SetKey("work-key", "acc-w", nil, nil)
	if err := work.Save(); err != nil {
		t.Fatal(err)
	}
	if got, _ := keyring.Get(KeychainService, "work"); got != "work-key" {
		t.Fatalf("keychain[work] = %q", got)
	}

	if def, _ := Load(""); def.Key() != "" || def.Path() != filepath.Join(dir, ConfigFileName) {
		t.Fatalf("default profile sees %q at %s", def.Key(), def.Path())
	}
	if def, _ := Load(DefaultProfile); def.Path() != filepath.Join(dir, ConfigFileName) {
		t.Fatalf("explicit default profile at %s", def.Path())
	}

	for _, bad := range []string{"../x", "a/b", "-x", strings.Repeat("a", 33), "a b"} {
		if _, err := Load(bad); !errors.Is(err, ErrInvalidProfile) {
			t.Errorf("Load(%q) = %v, want ErrInvalidProfile", bad, err)
		}
	}
}
