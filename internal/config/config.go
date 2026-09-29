package config

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
)

const (
	ConfigFileName = "config.json"
)

// ErrNoConfigDir is returned by Save when ConfigDir found no folder.
var ErrNoConfigDir = errors.New("no folder for the CLI config: set HOME (APPDATA on Windows)")

// ConfigDir is the folder of the CLI config: %APPDATA%\Square Cloud CLI on
// Windows, else $XDG_CONFIG_HOME/squarecloud or ~/.config/squarecloud. It is
// "" when none of those variables is set (minimal containers): the CLI still
// runs on SQUARECLOUD_API_KEY, only saving the config fails.
func ConfigDir() string {
	if a := os.Getenv("APPDATA"); runtime.GOOS == "windows" && a != "" {
		return filepath.Join(a, "Square Cloud CLI")
	}
	if b := os.Getenv("XDG_CONFIG_HOME"); b != "" {
		return filepath.Join(b, "squarecloud")
	}
	if dir, err := os.UserHomeDir(); err == nil {
		return filepath.Join(dir, ".config", "squarecloud")
	}
	return ""
}
