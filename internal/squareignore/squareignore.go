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

	trimmed := strings.TrimSpace(string(file))
	var filtered []string
	for _, e := range strings.Split(trimmed, "\n") {
		if e != "" {
			filtered = append(filtered, e)
		}
	}

	return filtered, nil
}
