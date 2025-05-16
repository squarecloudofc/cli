package zipper

import (
	"archive/zip"
	"fmt"
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

func ZipFolder(folder string, ignoreFiles []string) (*os.File, error) {
	destination, err := os.CreateTemp("", "*.zip")
	if err != nil {
		return nil, err
	}

	err = ZipFolderW(destination, folder, ignoreFiles)
	if err != nil {
		os.Remove(destination.Name())
		return nil, err
	}
	return destination, err
}

func ZipFolderW(writer io.Writer, folder string, ignoreFiles []string) error {
	w := zip.NewWriter(writer)
	defer w.Close()

	baseLen := len(folder) + 1

	return filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error walking path %s: %w", path, err)
		}

		if shouldIgnoreFile(info, ignoreFiles) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if !info.IsDir() {
			if err := addFileToZip(w, path, info, baseLen); err != nil {
				return fmt.Errorf("error adding file %s to zip: %w", path, err)
			}
		}
		return nil
	})
}

func addFileToZip(w *zip.Writer, path string, info os.FileInfo, baseLen int) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("error opening file %s: %w", path, err)
	}
	defer file.Close()

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return fmt.Errorf("error creating file header for %s: %w", path, err)
	}

	header.Method = zip.Deflate
	header.Name = path[baseLen:]

	writer, err := w.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("error creating header for %s: %w", path, err)
	}

	if _, err = io.Copy(writer, file); err != nil {
		return fmt.Errorf("error copying file %s to zip: %w", path, err)
	}

	return nil
}
