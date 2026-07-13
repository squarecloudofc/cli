package squareconfig

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// SquareConfig refers to the "squarecloud.app"/"squarecloud.config" file.
// The CLI only consumes ID; every other line is preserved verbatim on Save
// since the file is real platform configuration owned by the user.
type SquareConfig struct {
	ID string

	filename string
	// lines holds the raw file content so Save never destroys keys the CLI
	// doesn't know about.
	lines []string
	// created marks that the file didn't exist on Load.
	created bool
}

func New(path string) *SquareConfig {
	return &SquareConfig{filename: path}
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

	return squareconfig, squareconfig.LoadFromReader(file)
}

func (c *SquareConfig) IsCreated() bool {
	return c.created
}

func (c *SquareConfig) LoadFromReader(reader io.Reader) error {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		c.lines = append(c.lines, line)

		if key, value, found := strings.Cut(line, "="); found && strings.TrimSpace(key) == "ID" {
			c.ID = strings.TrimSpace(value)
		}
	}

	return scanner.Err()
}

func (c *SquareConfig) Save() error {
	if err := os.MkdirAll(filepath.Dir(c.filename), 0771); err != nil {
		return err
	}

	buf := &bytes.Buffer{}
	idWritten := false

	for _, line := range c.lines {
		if key, _, found := strings.Cut(line, "="); found && strings.TrimSpace(key) == "ID" {
			fmt.Fprintf(buf, "ID=%s\n", c.ID)
			idWritten = true
			continue
		}

		fmt.Fprintln(buf, line)
	}

	if !idWritten && c.ID != "" {
		fmt.Fprintf(buf, "ID=%s\n", c.ID)
	}

	return os.WriteFile(c.filename, buf.Bytes(), 0600)
}
