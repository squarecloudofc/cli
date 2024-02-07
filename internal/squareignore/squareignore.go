package squareignore

import (
	"errors"
	"os"
	"strings"
)

const SquareIgnoreFile = ".squareignore"

func Load() ([]string, error) {
	_, err := os.Lstat(SquareIgnoreFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New(".squareignore file does not exists")
		}

		return nil, err
	}

	file, err := os.ReadFile(SquareIgnoreFile)
	if err != nil {
		return nil, err
	}

	var filtered []string
	for _, e := range strings.Split(string(file), "\n") {
		entry := strings.TrimSpace(e)
		if entry != "" {
			filtered = append(filtered, entry)
		}
	}

	return filtered, nil
}
