package zipper

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func ZipFolder(folder string, file *os.File) error {
	w := zip.NewWriter(file)
	defer w.Close()

	err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.Contains(path, ".git") {
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
	if err != nil {
		return err
	}

	return nil
}
