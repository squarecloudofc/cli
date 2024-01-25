package config

import (
	"os"
	"path"
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

	return nil, err
}
