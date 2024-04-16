package squareignore

import (
	"errors"
	"os"
	"strings"
)

var SquareIgnoreFiles = []string{".squarecloudignore", ".squareignore", "square.ignore"}

func Load() ([]string, error) {
	var fileContent []byte

	for _, filename := range SquareIgnoreFiles {
		_, err := os.Lstat(filename)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, errors.New(".squareignore file does not exists")
			}

			return nil, err
		}

		fileContent, err = os.ReadFile(filename)
		if err != nil {
			return nil, err
		}
	}

	var filtered []string
	for _, e := range strings.Split(string(fileContent), "\n") {
		entry := strings.TrimSpace(e)
		if entry != "" {
			filtered = append(filtered, entry)
		}
	}

	return filtered, nil
}
