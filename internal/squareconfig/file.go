package squareconfig

import (
	"bytes"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/squarecloudofc/cli/pkg/properties"
)

// SquareConfig refers to "squarecloud.app" file
type SquareConfig struct {
	ID          string `properties:"ID"`
	DisplayName string `properties:"DISPLAY_NAME"`
	Description string `properties:"DESCRIPTION"`
	Main        string `properties:"MAIN"`
	Version     string `properties:"VERSION"`
	Memory      string `properties:"MEMORY"`
	Subdomain   string `properties:"SUBDOMAIN"`
	Start       string `properties:"START"`
	AutoRestart string `properties:"AUTO_RESTART"`

	filename string `properties:"-"`

	// we will use this created property to check if the file exists or not (internal use only)
	created bool `properties:"-"`
}

func New(path string) *SquareConfig {
	return &SquareConfig{
		filename: path,
	}
}

func Load() (*SquareConfig, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	filepath := path.Join(cwd, GetConfigFile())
	squareconfig := New(filepath)

	file, err := os.Open(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			squareconfig.created = true
			return squareconfig, nil
		}

		return nil, err
	}
	defer file.Close()
	err = squareconfig.LoadFromReader(file)

	squareconfig.created = false
	return squareconfig, err
}

func (c *SquareConfig) IsCreated() bool {
	return c.created
}

func (c *SquareConfig) LoadFromReader(reader io.Reader) error {
	buf := new(bytes.Buffer)
	_, err := io.Copy(buf, reader)
	if err != nil {
		return err
	}

	err = properties.Unmarshal(buf.Bytes(), c)
	return err
}

func (c *SquareConfig) Save() (tmpErr error) {
	err := os.MkdirAll(filepath.Dir(c.filename), 0771)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(c.filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}

	defer file.Close()

	data, err := properties.Marshal(c)
	if err != nil {
		return err
	}

	_, err = file.Write(data)
	return err
}
