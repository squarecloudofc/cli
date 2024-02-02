package zipper

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var (
	DefaultIgnoredFiles = []string{"node_modules/", ".git/", ".vscode/", ".github/", ".cache/", "package-lock.json"}
)

func shouldIgnoreFile(filedir string, file os.FileInfo) bool {
	var ignoreEntries []string
	ignoreEntries = append(ignoreEntries, DefaultIgnoredFiles...)

	for _, entry := range ignoreEntries {
		if strings.HasSuffix(entry, "/") && !file.IsDir() {
			if strings.Contains(filepath.Dir(filedir)+"/", entry) {
				return true
			}
		}

		if strings.Contains(filedir, entry) {
			return true
		}
	}

	return false
}

func ZipFolder(folder string, file *os.File) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	w := zip.NewWriter(file)
	defer w.Close()

	return filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		absolutepath, _ := strings.CutPrefix(path, fmt.Sprintf("%s/", wd))
		if shouldIgnoreFile(absolutepath, info) {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		fileinfo, err := file.Stat()
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(fileinfo)
		if err != nil {
			return err
		}

		header.Name = path[len(folder)+1:]
		writer, err := w.CreateHeader(header)
		if err != nil {
			return err
		}

		if _, err = io.Copy(writer, file); err != nil {
			return err
		}

		return nil
	})
}
