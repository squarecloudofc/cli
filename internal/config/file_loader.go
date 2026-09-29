package config

import (
	"os"
	"path/filepath"
)

// Load reads the saved config. The language is not part of it: the CLI
// detects it on every run (i18n.DetectSystemLanguage), and a "locale" left by
// older versions is ignored and dropped on the next Save.
func Load() (*Config, error) {
	dir := ConfigDir()
	if dir == "" {
		return &Config{}, nil // Save reports ErrNoConfigDir
	}

	configPath := filepath.Join(dir, ConfigFileName)
	configFile := New(configPath)

	file, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return configFile, nil
		}

		return nil, err
	}
	defer file.Close()

	return configFile, configFile.LoadFromReader(file)
}
