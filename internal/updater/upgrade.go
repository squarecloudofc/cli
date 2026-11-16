package updater

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// ReleaseDownloadURL is where release files are downloaded from: a direct
// download, so no GitHub API rate limit. A variable so tests use a fake.
var ReleaseDownloadURL = "https://github.com/squarecloudofc/cli/releases/download"

// maxDownload caps a release file; archives are around 10 MB.
const maxDownload = 200 << 20

// AssetName is the release archive for goos/goarch, as goreleaser.yaml names
// it; 32-bit ARM is built as armv7 only.
func AssetName(goos, goarch string) string {
	if goarch == "arm" {
		goarch = "armv7"
	}
	ext := ".tar.gz"
	if goos == "windows" {
		ext = ".zip"
	}
	return "squarecloud_" + goos + "_" + goarch + ext
}

// InstalledWithNPM reports whether exe is the binary @squarecloud/cli keeps
// in node_modules; npm, not upgrade, has to update that one.
func InstalledWithNPM(exe string) bool {
	return strings.Contains(strings.ReplaceAll(exe, `\`, "/"), "/node_modules/")
}

// dpkgList is the file list dpkg keeps for an installed squarecloud package.
var dpkgList = "/var/lib/dpkg/info/squarecloud.list"

// PackagedBy reports which package installed exe: "deb", "rpm", or "" for
// any other install. ponytail: the .deb and .rpm put the binary in /usr/bin
// and no other installer does (install.sh uses ~/.local/bin, older versions
// /usr/local/bin); dpkg's file list tells the two apart. Ask dpkg -S and
// rpm -qf if a copy placed there by hand ever needs telling apart.
func PackagedBy(exe string) string {
	if filepath.ToSlash(exe) != "/usr/bin/squarecloud" {
		return ""
	}
	if _, err := os.Stat(dpkgList); err == nil {
		return "deb"
	}
	return "rpm"
}

// checksumFor returns the SHA-256 of file in a goreleaser checksums.txt
// ("<hex>  <name>" lines).
func checksumFor(sums, file string) (string, error) {
	for _, line := range strings.Split(sums, "\n") {
		if fields := strings.Fields(line); len(fields) == 2 && fields[1] == file && len(fields[0]) == 64 {
			if _, err := hex.DecodeString(fields[0]); err == nil {
				return strings.ToLower(fields[0]), nil
			}
		}
	}
	return "", fmt.Errorf("checksums.txt has no entry for %s", file)
}

// Download fetches the release tag's archive for goos/goarch, checks it
// against the release's checksums.txt and returns the CLI binary inside.
func Download(ctx context.Context, tag, goos, goarch string) ([]byte, error) {
	asset := AssetName(goos, goarch)
	sums, err := get(ctx, ReleaseDownloadURL+"/"+tag+"/checksums.txt")
	if err != nil {
		return nil, err
	}
	want, err := checksumFor(string(sums), asset)
	if err != nil {
		return nil, err
	}

	archive, err := get(ctx, ReleaseDownloadURL+"/"+tag+"/"+asset)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(archive)
	if got := hex.EncodeToString(sum[:]); got != want {
		return nil, fmt.Errorf("%s does not match checksums.txt (expected %s, got %s)", asset, want, got)
	}

	return extractBinary(archive, goos)
}

func get(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: %s", url, res.Status)
	}
	return io.ReadAll(io.LimitReader(res.Body, maxDownload))
}

// extractBinary returns the squarecloud binary of a release archive: a zip
// on Windows, a tar.gz elsewhere.
func extractBinary(archive []byte, goos string) ([]byte, error) {
	if goos == "windows" {
		zr, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
		if err != nil {
			return nil, err
		}
		for _, f := range zr.File {
			if path.Base(f.Name) == "squarecloud.exe" {
				rc, err := f.Open()
				if err != nil {
					return nil, err
				}
				defer rc.Close()
				return io.ReadAll(io.LimitReader(rc, maxDownload))
			}
		}
		return nil, errors.New("squarecloud.exe not found in the release archive")
	}

	gz, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return nil, err
	}
	tr := tar.NewReader(gz)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			return nil, errors.New("squarecloud not found in the release archive")
		}
		if err != nil {
			return nil, err
		}
		if h.Typeflag == tar.TypeReg && path.Base(h.Name) == "squarecloud" {
			return io.ReadAll(io.LimitReader(tr, maxDownload))
		}
	}
}

// Replace swaps the binary at exe for bin, keeping its permissions. Windows
// can't overwrite a running executable but can rename it, so there the old
// one is left as exe+".old" (removed on the next run, by removeOldBinary).
func Replace(exe string, bin []byte, goos string) error {
	info, err := os.Stat(exe)
	if err != nil {
		return err
	}
	// Next to exe: a rename across file systems would fail.
	tmp := exe + ".new"
	_ = os.Remove(tmp) // a leftover keeps its old mode through WriteFile
	if err := os.WriteFile(tmp, bin, info.Mode().Perm()); err != nil {
		return err
	}

	if goos == "windows" {
		old := exe + ".old"
		_ = os.Remove(old)
		if err := os.Rename(exe, old); err != nil {
			_ = os.Remove(tmp)
			return err
		}
		if err := os.Rename(tmp, exe); err != nil {
			_ = os.Rename(old, exe)
			_ = os.Remove(tmp)
			return err
		}
		return nil
	}

	if err := os.Rename(tmp, exe); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}
