package config

import (
	"os"
	"path"

	"github.com/squarecloudofc/cli/internal/i18n"
)

func Load() (*Config, error) {
	configPath := path.Join(ConfigDir(), ConfigFileName)
	configFile := New(configPath)

	file, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return configFile, nil
		}

		return nil, err
	}
	defer file.Close()
	err = configFile.LoadFromReader(file)

	if configFile.Locale == "" {
		lang := i18n.DetectSystemLanguage()
		configFile.Locale = lang
		configFile.Save()
	}

	return configFile, err
}
