package zipper

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var DefaultIgnoredFiles = []string{"node_modules", ".git/", ".gitignore", ".squareignore", ".squarecloudignore", "square.ignore", ".vscode", ".github"}

func shouldIgnoreFile(info os.FileInfo, ignoreEntries []string) bool {
	ignoreEntries = append(ignoreEntries, DefaultIgnoredFiles...)

	for _, pattern := range ignoreEntries {
		match, _ := filepath.Match(pattern, info.Name())
		if match {
			return true
		}

		if strings.HasSuffix(info.Name(), pattern) {
			return true
		}

		if strings.Contains(pattern, "/") && info.IsDir() && strings.HasSuffix(info.Name(), strings.ReplaceAll(pattern, "/", "")) {
			return true
		}
	}

	return false
}

func ZipFolder(folder string, destination *os.File, ignoreFiles []string) error {
	w := zip.NewWriter(destination)
	defer w.Close()

	return filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if shouldIgnoreFile(info, ignoreFiles) {
			if info.IsDir() {
				return filepath.SkipDir
			}

			return nil
		}

		if !info.IsDir() {
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

			header.Method = zip.Deflate
			header.Name = path[len(folder)+1:]
			writer, err := w.CreateHeader(header)
			if err != nil {
				return err
			}

			if _, err = io.Copy(writer, file); err != nil {
				return err
			}

		}
		return nil
	})
}
