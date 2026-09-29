package zipper

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"testing"

	"github.com/squarecloudofc/cli/pkg/squareignore"
)

// writeFiles creates each slash-separated name under dir with its name as
// content.
func writeFiles(t *testing.T, dir string, names ...string) {
	t.Helper()
	for _, name := range names {
		p := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(name), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// zipEntries zips folder and returns entry name -> content.
func zipEntries(t *testing.T, folder string, ignore *squareignore.Matcher, skipped func(string)) map[string]string {
	t.Helper()
	var buf bytes.Buffer
	if err := ZipFolderW(&buf, folder, ignore, skipped); err != nil {
		t.Fatal(err)
	}
	r, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}

	entries := map[string]string{}
	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		data, _ := io.ReadAll(rc)
		rc.Close()
		entries[f.Name] = string(data)
	}
	return entries
}

func names(entries map[string]string) []string {
	var out []string
	for name := range entries {
		out = append(out, name)
	}
	slices.Sort(out)
	return out
}

func TestZipFolderEntryNames(t *testing.T) {
	// The folder name matches an ignore pattern; only its contents are matched.
	dir := filepath.Join(t.TempDir(), "api-dist")
	writeFiles(t, dir, "index.js", "src/app.js", "dist/bundle.js")

	// A trailing separator (a drive or filesystem root) must not shift names.
	for _, folder := range []string{dir, dir + string(os.PathSeparator)} {
		got := names(zipEntries(t, folder, squareignore.New("dist"), nil))
		if want := []string{"index.js", "src/app.js"}; !slices.Equal(got, want) {
			t.Errorf("ZipFolderW(%q) entries = %q, want %q", folder, got, want)
		}
	}
}

func TestZipFolderIgnoreRules(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir,
		"index.js", "nodist/a.js", "src/build/out.js", "build/out.js",
		"dist/bundle.js", "dist/keep.js",
		"app.log", "keep.log", "logs/today.txt",
		"node_modules/x/index.js", "yarn.lock", "squarecloud.ignore",
	)

	ignore := squareignore.New("dist/", "!dist/keep.js", "/build", "*.log", "!keep.log", "logs/", "!yarn.lock")
	got := names(zipEntries(t, dir, ignore, nil))
	want := []string{"index.js", "keep.log", "nodist/a.js", "squarecloud.ignore", "src/build/out.js", "yarn.lock"}
	if !slices.Equal(got, want) {
		t.Errorf("entries = %q, want %q", got, want)
	}
}

func TestZipFolderSymlinks(t *testing.T) {
	outside := t.TempDir()
	writeFiles(t, outside, "shared.txt", "lib/util.js")

	dir := t.TempDir()
	writeFiles(t, dir, "index.js")
	links := map[string]string{
		"shared.txt": filepath.Join(outside, "shared.txt"),
		"lib":        filepath.Join(outside, "lib"),
		"broken":     filepath.Join(outside, "missing.txt"),
	}
	for name, target := range links {
		if err := os.Symlink(target, filepath.Join(dir, name)); err != nil {
			t.Skipf("cannot create symlinks here: %v", err)
		}
	}

	var skipped []string
	entries := zipEntries(t, dir, squareignore.New(), func(rel string) { skipped = append(skipped, rel) })

	if got, want := names(entries), []string{"index.js", "shared.txt"}; !slices.Equal(got, want) {
		t.Errorf("entries = %q, want %q", got, want)
	}
	if entries["shared.txt"] != "shared.txt" {
		t.Errorf("link to a file stored %q, want the file's content", entries["shared.txt"])
	}
	slices.Sort(skipped)
	if want := []string{"broken", "lib"}; !slices.Equal(skipped, want) {
		t.Errorf("skipped = %q, want %q", skipped, want)
	}
}

// Junctions are Windows' unprivileged directory links (symlinks need
// Developer Mode, so the test above is often skipped there).
func TestZipFolderJunctions(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("junctions are Windows-only")
	}
	outside := t.TempDir()
	writeFiles(t, outside, "lib/util.js", "gone/x.js")

	dir := t.TempDir()
	writeFiles(t, dir, "index.js")
	for _, name := range []string{"lib", "gone"} {
		if out, err := exec.Command("cmd", "/c", "mklink", "/J", filepath.Join(dir, name), filepath.Join(outside, name)).CombinedOutput(); err != nil {
			t.Skipf("mklink /J: %v %s", err, out)
		}
	}
	if err := os.RemoveAll(filepath.Join(outside, "gone")); err != nil {
		t.Fatal(err)
	}

	var skipped []string
	entries := zipEntries(t, dir, squareignore.New(), func(rel string) { skipped = append(skipped, rel) })
	if got, want := names(entries), []string{"index.js"}; !slices.Equal(got, want) {
		t.Errorf("entries = %q, want %q", got, want)
	}
	slices.Sort(skipped)
	if want := []string{"gone", "lib"}; !slices.Equal(skipped, want) {
		t.Errorf("skipped = %q, want %q", skipped, want)
	}
}
