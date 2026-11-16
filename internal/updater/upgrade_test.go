package updater

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAssetName(t *testing.T) {
	for _, c := range []struct{ goos, goarch, want string }{
		{"linux", "amd64", "squarecloud_linux_amd64.tar.gz"},
		{"linux", "arm", "squarecloud_linux_armv7.tar.gz"},
		{"linux", "386", "squarecloud_linux_386.tar.gz"},
		{"darwin", "arm64", "squarecloud_darwin_arm64.tar.gz"},
		{"windows", "amd64", "squarecloud_windows_amd64.zip"},
	} {
		if got := AssetName(c.goos, c.goarch); got != c.want {
			t.Errorf("AssetName(%s, %s) = %s, want %s", c.goos, c.goarch, got, c.want)
		}
	}
}

func TestInstalledWithNPM(t *testing.T) {
	for exe, want := range map[string]bool{
		`C:\Users\me\AppData\Roaming\npm\node_modules\@squarecloud\cli\bin\squarecloud.exe`: true,
		"/usr/lib/node_modules/@squarecloud/cli/bin/squarecloud":                            true,
		"/home/me/.local/bin/squarecloud":                                                   false,
		`C:\Users\me\AppData\Local\squarecloud\bin\squarecloud.exe`:                         false,
	} {
		if got := InstalledWithNPM(exe); got != want {
			t.Errorf("InstalledWithNPM(%s) = %v, want %v", exe, got, want)
		}
	}
}

func TestChecksumFor(t *testing.T) {
	hash := strings.Repeat("ab", 32)
	sums := strings.Repeat("0", 64) + "  squarecloud_linux_amd64.tar.gz\n" +
		strings.ToUpper(hash) + "  squarecloud_windows_amd64.zip\r\n" +
		"nothex  squarecloud_darwin_arm64.tar.gz\n"

	if got, err := checksumFor(sums, "squarecloud_windows_amd64.zip"); err != nil || got != hash {
		t.Errorf("windows = %q, %v; want %q", got, err, hash)
	}
	for _, file := range []string{"squarecloud_darwin_arm64.tar.gz", "squarecloud_linux_arm64.tar.gz", "squarecloud_linux_amd64"} {
		if got, err := checksumFor(sums, file); err == nil {
			t.Errorf("%s = %q, want an error", file, got)
		}
	}
}

func tarGz(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	tw := tar.NewWriter(gz)
	for name, content := range files {
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0755, Size: int64(len(content)), Typeflag: tar.TypeReg}); err != nil {
			t.Fatal(err)
		}
		_, _ = tw.Write([]byte(content))
	}
	_ = tw.Close()
	_ = gz.Close()
	return b.Bytes()
}

func zipOf(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var b bytes.Buffer
	zw := zip.NewWriter(&b)
	for name, content := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = w.Write([]byte(content))
	}
	_ = zw.Close()
	return b.Bytes()
}

// fakeRelease serves a tag's checksums.txt and archives, as GitHub does,
// and points ReleaseDownloadURL at it.
func fakeRelease(t *testing.T, tag string, files map[string][]byte, sums string) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name, ok := strings.CutPrefix(r.URL.Path, "/"+tag+"/")
		if name == "checksums.txt" && ok {
			_, _ = w.Write([]byte(sums))
			return
		}
		if data, found := files[name]; ok && found {
			_, _ = w.Write(data)
			return
		}
		http.NotFound(w, r)
	}))
	t.Cleanup(srv.Close)
	old := ReleaseDownloadURL
	ReleaseDownloadURL = srv.URL
	t.Cleanup(func() { ReleaseDownloadURL = old })
}

func sha(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func TestDownload(t *testing.T) {
	linux := tarGz(t, map[string]string{"completions/x.bash": "x", "LICENSE": "l", "squarecloud": "linux-bin"})
	windows := zipOf(t, map[string]string{"README.md": "r", "squarecloud.exe": "windows-bin"})
	files := map[string][]byte{
		"squarecloud_linux_amd64.tar.gz":  linux,
		"squarecloud_windows_amd64.zip":   windows,
		"squarecloud_darwin_arm64.tar.gz": []byte("tampered"),
	}
	sums := fmt.Sprintf("%s  squarecloud_linux_amd64.tar.gz\n%s  squarecloud_windows_amd64.zip\n%s  squarecloud_darwin_arm64.tar.gz\n",
		sha(linux), sha(windows), sha([]byte("original")))
	fakeRelease(t, "v9.9.9", files, sums)
	ctx := context.Background()

	if bin, err := Download(ctx, "v9.9.9", "linux", "amd64"); err != nil || string(bin) != "linux-bin" {
		t.Errorf("linux = %q, %v", bin, err)
	}
	if bin, err := Download(ctx, "v9.9.9", "windows", "amd64"); err != nil || string(bin) != "windows-bin" {
		t.Errorf("windows = %q, %v", bin, err)
	}
	if _, err := Download(ctx, "v9.9.9", "darwin", "arm64"); err == nil || !strings.Contains(err.Error(), "does not match checksums.txt") {
		t.Errorf("tampered archive = %v, want a checksum error", err)
	}
	if _, err := Download(ctx, "v9.9.9", "linux", "arm64"); err == nil || !strings.Contains(err.Error(), "no entry") {
		t.Errorf("unlisted archive = %v, want a missing entry error", err)
	}
	if _, err := Download(ctx, "v0.0.1", "linux", "amd64"); err == nil || !strings.Contains(err.Error(), "404") {
		t.Errorf("unknown tag = %v, want a 404", err)
	}
}

func TestExtractBinaryMissing(t *testing.T) {
	if _, err := extractBinary(tarGz(t, map[string]string{"LICENSE": "l"}), "linux"); err == nil {
		t.Error("tar.gz without the binary: no error")
	}
	if _, err := extractBinary(zipOf(t, map[string]string{"squarecloud": "unix name"}), "windows"); err == nil {
		t.Error("zip without squarecloud.exe: no error")
	}
}

func TestReplace(t *testing.T) {
	for _, goos := range []string{"linux", "windows"} {
		t.Run(goos, func(t *testing.T) {
			dir := t.TempDir()
			exe := filepath.Join(dir, "squarecloud")
			if err := os.WriteFile(exe, []byte("old"), 0755); err != nil {
				t.Fatal(err)
			}
			// Leftovers of an earlier upgrade.
			_ = os.WriteFile(exe+".old", []byte("older"), 0644)
			_ = os.WriteFile(exe+".new", []byte("partial"), 0600)

			if err := Replace(exe, []byte("new"), goos); err != nil {
				t.Fatalf("Replace = %v", err)
			}
			if data, _ := os.ReadFile(exe); string(data) != "new" {
				t.Errorf("binary = %q, want new", data)
			}
			if _, err := os.Stat(exe + ".new"); !os.IsNotExist(err) {
				t.Errorf(".new left behind: %v", err)
			}
			old, err := os.ReadFile(exe + ".old")
			if goos == "windows" && string(old) != "old" {
				t.Errorf(".old = %q, %v; want the running binary kept aside", old, err)
			}
			if info, err := os.Stat(exe); err != nil || (os.PathSeparator == '/' && info.Mode().Perm() != 0755) {
				t.Errorf("mode = %v, %v; want 0755 kept", info.Mode(), err)
			}
		})
	}

	if err := Replace(filepath.Join(t.TempDir(), "missing"), []byte("new"), "linux"); err == nil {
		t.Error("Replace of a missing binary: no error")
	}
}

func TestPackagedBy(t *testing.T) {
	list := filepath.Join(t.TempDir(), "squarecloud.list")
	old := dpkgList
	dpkgList = list
	t.Cleanup(func() { dpkgList = old })

	if got := PackagedBy("/usr/bin/squarecloud"); got != "rpm" {
		t.Errorf("/usr/bin without a dpkg list = %q, want rpm", got)
	}
	if err := os.WriteFile(list, nil, 0644); err != nil {
		t.Fatal(err)
	}
	for exe, want := range map[string]string{
		"/usr/bin/squarecloud":                                      "deb",
		"/usr/local/bin/squarecloud":                                "",
		"/home/me/.local/bin/squarecloud":                           "",
		`C:\Users\me\AppData\Local\squarecloud\bin\squarecloud.exe`: "",
	} {
		if got := PackagedBy(exe); got != want {
			t.Errorf("PackagedBy(%s) = %q, want %q", exe, got, want)
		}
	}
}
