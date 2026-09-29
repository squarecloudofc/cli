package config

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	filename string `json:"-"`

	AuthToken string `json:"auth_token,omitempty"`
	// AccountID owns AuthToken; a login for another account asks first.
	AccountID string `json:"account_id,omitempty"`

	// Known only for keys issued by the browser login; empty for pasted keys.
	KeyExpiresAt *time.Time `json:"key_expires_at,omitempty"`
	KeyScopes    []string   `json:"key_scopes,omitempty"`
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

// Save writes the config atomically: a temp file in the same folder, synced,
// then renamed over the old one. A crash mid-write leaves the previous file,
// not an empty one that would log the person out.
func (c *Config) Save() error {
	if c.filename == "" {
		return ErrNoConfigDir
	}

	dir := filepath.Dir(c.filename)
	if err := os.MkdirAll(dir, 0700); err != nil { // holds the API key
		return err
	}

	data, err := json.MarshalIndent(c, "", "\t")
	if err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, ConfigFileName+".*.tmp") // 0600
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name()) // no-op once renamed

	_, err = tmp.Write(data)
	if err == nil {
		err = tmp.Sync()
	}
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}

	if os.Rename(tmp.Name(), c.filename) != nil {
		// Windows refuses to replace a file another process holds open
		// (antivirus, an editor): write in place, as before, over failing.
		return os.WriteFile(c.filename, data, 0600)
	}
	return nil
}

// SetKey stores the API key with what is known about it: its account, and
// (browser login only) expiry and scopes. All empty for a logout.
func (c *Config) SetKey(key, accountID string, expiresAt *time.Time, scopes []string) {
	c.AuthToken = key
	c.AccountID = accountID
	c.KeyExpiresAt = expiresAt
	c.KeyScopes = scopes
}
