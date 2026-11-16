package config

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/zalando/go-keyring"
)

// KeychainService names the CLI's entries in the OS keychain (Windows
// Credential Manager, macOS Keychain, Secret Service on Linux): one per
// profile, the profile name being the account.
const KeychainService = "Square Cloud CLI"

type Config struct {
	filename string `json:"-"`
	profile  string
	// keyRead: the keychain was consulted, or the key set, during this run.
	keyRead bool
	// keyInFile: the last Save found no keychain and wrote the key here.
	keyInFile bool

	// AuthToken is the API key. On disk only when no keychain was available
	// (or in a file written by a version before the keychain); use Key.
	AuthToken string `json:"auth_token,omitempty"`
	// AccountID owns AuthToken; a login for another account asks first.
	AccountID string `json:"account_id,omitempty"`

	// Known only for keys issued by the browser login; empty for pasted keys.
	KeyExpiresAt *time.Time `json:"key_expires_at,omitempty"`
	KeyScopes    []string   `json:"key_scopes,omitempty"`
}

// New is the default profile's config kept at filename.
func New(filename string) *Config {
	return &Config{
		filename: filename,
		profile:  DefaultProfile,
	}
}

// Profile is the name this config was loaded for.
func (c *Config) Profile() string { return c.profile }

// Path is the config file, "" without a config folder.
func (c *Config) Path() string { return c.filename }

// KeyInFile reports that the last Save kept the key in the config file
// because no keychain was available (containers, WSL without D-Bus).
func (c *Config) KeyInFile() bool { return c.keyInFile }

func (c *Config) LoadFromReader(reader io.Reader) error {
	err := json.NewDecoder(reader).Decode(c)
	if err != nil {
		return err
	}

	return nil
}

// Key is the saved API key. The first call reads the keychain, or, when the
// file holds the key (an older version, or no keychain at the last login),
// moves it into the keychain if one is available now.
func (c *Config) Key() string {
	if c.keyRead || c.filename == "" {
		return c.AuthToken
	}
	c.keyRead = true

	if c.AuthToken == "" {
		c.AuthToken, _ = keyring.Get(KeychainService, c.profile)
		return c.AuthToken
	}
	if keyring.Set(KeychainService, c.profile, c.AuthToken) == nil {
		stored := *c
		stored.AuthToken = ""
		_ = stored.write() // best effort: the next run tries again
	}
	return c.AuthToken
}

// Save stores the key in the keychain, or in the file when there is none,
// and the rest of the config in the file. An empty key (logout) is removed
// from both.
func (c *Config) Save() error {
	if c.filename == "" {
		return ErrNoConfigDir
	}

	stored := *c
	if c.AuthToken == "" {
		_ = keyring.Delete(KeychainService, c.profile) // not found, or no keychain
	} else if keyring.Set(KeychainService, c.profile, c.AuthToken) == nil {
		stored.AuthToken = ""
	}
	c.keyRead = true
	c.keyInFile = stored.AuthToken != ""
	return stored.write()
}

// write saves the config atomically: a temp file in the same folder, synced,
// then renamed over the old one. A crash mid-write leaves the previous file,
// not an empty one that would log the person out.
func (c *Config) write() error {
	dir := filepath.Dir(c.filename)
	if err := os.MkdirAll(dir, 0700); err != nil { // may hold the API key
		return err
	}

	data, err := json.MarshalIndent(c, "", "\t")
	if err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, filepath.Base(c.filename)+".*.tmp") // 0600
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
	c.keyRead = true
}
