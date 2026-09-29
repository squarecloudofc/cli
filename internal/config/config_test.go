package config

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestSaveReplacesAtomically(t *testing.T) {
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

	c, err := Load() // must not panic nor read ./config.json
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
	dir := t.TempDir()
	t.Setenv("APPDATA", dir)
	t.Setenv("XDG_CONFIG_HOME", dir)
	path := filepath.Join(ConfigDir(), ConfigFileName)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	old := `{"locale": "en", "auth_token": "k"}`
	if err := os.WriteFile(path, []byte(old), 0600); err != nil {
		t.Fatal(err)
	}

	c, err := Load()
	if err != nil {
		t.Fatal(err)
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
