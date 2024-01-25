package config

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

const (
	ConfigFileName = "config.json"
)

var (
	// we need load the config file once
	initConfigDir = new(sync.Once)
	configDir     = ""
)

func ConfigDir() string {
	initConfigDir.Do(func() {
		if a := os.Getenv("AppData"); runtime.GOOS == "windows" && a != "" {
			configDir = filepath.Join(a, "Square Cloud CLI")
		} else if b := os.Getenv("XDG_CONFIG_HOME"); b != "" {
			configDir = filepath.Join(b, "squarecloud")
		} else {
			dir, err := os.UserHomeDir()
			if err != nil {
				panic("cannot get a valid dir for square cloud config")
			}
			configDir = filepath.Join(dir, ".config", "squarecloud")
		}
	})

	return configDir
}
