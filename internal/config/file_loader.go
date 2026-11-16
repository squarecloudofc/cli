package config

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
)

// DefaultProfile is the login used without --profile or SQUARECLOUD_PROFILE.
const DefaultProfile = "default"

var profileName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{0,31}$`)

// ErrInvalidProfile is returned by Load for a profile name that is not 1-32
// letters, digits, - or _.
var ErrInvalidProfile = errors.New("invalid profile name")

// Load reads the saved config of a profile ("" is DefaultProfile): config.json
// for the default one, profiles/<name>.json for the others. The API key is
// read from the keychain on first use (Key). The language is not part of it:
// the CLI detects it on every run (i18n.DetectSystemLanguage), and a "locale"
// left by older versions is ignored and dropped on the next Save.
func Load(profile string) (*Config, error) {
	if profile == "" {
		profile = DefaultProfile
	}
	if !profileName.MatchString(profile) {
		return nil, ErrInvalidProfile
	}

	dir := ConfigDir()
	if dir == "" {
		return &Config{profile: profile}, nil // Save reports ErrNoConfigDir
	}

	configPath := filepath.Join(dir, ConfigFileName)
	if profile != DefaultProfile {
		configPath = filepath.Join(dir, "profiles", profile+".json")
	}
	configFile := New(configPath)
	configFile.profile = profile

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
