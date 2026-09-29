package zipper

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/squarecloudofc/cli/pkg/squareignore"
)

// ZipFolder zips folder into a temp file, rewound for reading. The caller
// closes and removes it.
func ZipFolder(folder string, ignore *squareignore.Matcher, skipped func(rel string)) (*os.File, error) {
	destination, err := os.CreateTemp("", "squarecloud-*.zip")
	if err != nil {
		return nil, err
	}

	err = ZipFolderW(destination, folder, ignore, skipped)
	if err == nil {
		_, err = destination.Seek(0, io.SeekStart)
	}
	if err != nil {
		destination.Close() // Windows cannot remove an open file
		os.Remove(destination.Name())
		return nil, err
	}
	return destination, nil
}

// ZipFolderW writes folder as a zip to writer, leaving out what ignore
// matches. A symbolic link to a file is stored as that file; a link to a
// folder or a broken link is left out and reported to skipped (optional)
// with its slash-separated path.
func ZipFolderW(writer io.Writer, folder string, ignore *squareignore.Matcher, skipped func(rel string)) error {
	w := zip.NewWriter(writer)

	err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error walking path %s: %w", path, err)
		}

		// The folder itself is never matched: "dist" would otherwise skip a
		// project folder named "api-dist" and produce an empty zip.
		if path == folder {
			return nil
		}

		rel, err := filepath.Rel(folder, path)
		if err != nil {
			return fmt.Errorf("error resolving %s: %w", path, err)
		}
		rel = filepath.ToSlash(rel)

		// Symlinks, and Windows junctions (irregular, not a dir, for Lstat).
		link := !info.IsDir() && info.Mode()&(os.ModeSymlink|os.ModeIrregular) != 0
		isDir, broken := info.IsDir(), false
		if link {
			target, err := os.Stat(path)
			broken = err != nil
			isDir = err == nil && target.IsDir()
		}

		if ignore.Match(rel, isDir) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		switch {
		case link && (broken || isDir):
			if skipped != nil {
				skipped(rel)
			}
		case !info.IsDir():
			if err := addFileToZip(w, path, rel); err != nil {
				return fmt.Errorf("error adding file %s to zip: %w", path, err)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Close writes the central directory; a failure leaves an unreadable zip.
	return w.Close()
}

func addFileToZip(w *zip.Writer, path, name string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("error opening file %s: %w", path, err)
	}
	defer file.Close()

	// Stat the opened file, not the walked entry: a symlink is stored as the
	// file it points to, not as a link entry whose "target" is that content.
	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("error reading file info for %s: %w", path, err)
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return fmt.Errorf("error creating file header for %s: %w", path, err)
	}

	header.Method = zip.Deflate
	header.Name = name

	writer, err := w.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("error creating header for %s: %w", path, err)
	}

	if _, err = io.Copy(writer, file); err != nil {
		return fmt.Errorf("error copying file %s to zip: %w", path, err)
	}

	return nil
}
