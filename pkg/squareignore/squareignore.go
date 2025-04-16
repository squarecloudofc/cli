package squareignore

import (
	"os"
	"path/filepath"
	"strings"
)

var SquareIgnoreFiles = []string{".squarecloudignore", ".squareignore", "square.ignore", "squarecloud.ignore"}

func Load() ([]string, error) {
	var fileContent []byte

	path, err := os.Getwd()
	if err != nil {
		return []string{}, err
	}

	for _, filename := range SquareIgnoreFiles {
		ignorefilepath := filepath.Join(path, filename)
		_, err := os.Lstat(ignorefilepath)
		if err != nil {
			continue
		}

		fileContent, err = os.ReadFile(ignorefilepath)
		if err != nil {
			return []string{}, err
		}

		break
	}

	var filtered []string
	for e := range strings.SplitSeq(string(fileContent), "\n") {
		entry := strings.TrimSpace(e)
		if entry != "" {
			filtered = append(filtered, entry)
		}
	}

	return filtered, nil
}
