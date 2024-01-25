package config

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
)

type Config struct {
	filename string `json:"-"`

	AuthToken       string `json:"auth_token,omitempty"`
	LastUpdateCheck string `json:"last_update_check,omitempty"`
}

func New(filename string) *Config {
	return &Config{
		filename: filename,
	}
}

func (c *Config) LoadFromReader(reader io.Reader) error {
	err := json.NewDecoder(reader).Decode(c)
	if err != nil {
		return err
	}

	return nil
}

func (c *Config) Save() (tmpErr error) {
	err := os.MkdirAll(filepath.Dir(c.filename), 0771)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(c.filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}

	defer file.Close()

	data, err := json.MarshalIndent(c, "", "\t")
	if err != nil {
		return err
	}

	_, err = file.Write(data)
	return err
}
