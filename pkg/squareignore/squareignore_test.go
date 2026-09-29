package squareignore

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMatch(t *testing.T) {
	type check struct {
		path  string
		isDir bool
		want  bool
	}
	cases := []struct {
		name   string
		rules  []string
		checks []check
	}{
		{"name matches at any depth, not as a suffix", []string{"dist"}, []check{
			{"dist", true, true}, {"dist", false, true}, {"src/dist", true, true},
			{"nodist", true, false}, {"dist.js", false, false},
		}},
		{"leading slash anchors to the root", []string{"/build"}, []check{
			{"build", true, true}, {"src/build", true, false},
		}},
		{"middle slash anchors to the root", []string{"src/build"}, []check{
			{"src/build", true, true}, {"build", true, false}, {"app/src/build", true, false},
		}},
		{"negation, last match wins", []string{"*.log", "!keep.log"}, []check{
			{"app.log", false, true}, {"logs/app.log", false, true},
			{"keep.log", false, false}, {"sub/keep.log", false, false},
		}},
		{"a later rule overrides a negation", []string{"*.log", "!keep.log", "keep.log"}, []check{
			{"keep.log", false, true},
		}},
		{"trailing slash matches directories only", []string{"logs/"}, []check{
			{"logs", true, true}, {"app/logs", true, true}, {"logs", false, false},
		}},
		{"leading **/", []string{"**/tmp"}, []check{
			{"tmp", true, true}, {"a/b/tmp", false, true}, {"tmpx", true, false},
		}},
		{"middle /**/", []string{"a/**/b"}, []check{
			{"a/b", false, true}, {"a/x/b", false, true}, {"a/x/y/b", true, true},
			{"x/a/b", false, false}, {"a/bb", false, false},
		}},
		{"trailing /**", []string{"x/**"}, []check{
			{"x/a", false, true}, {"x/a/b", true, true}, {"x", true, false},
		}},
		{"* and ? do not cross slashes", []string{"src/*.js", "file?.txt"}, []check{
			{"src/a.js", false, true}, {"src/x/a.js", false, false},
			{"file1.txt", false, true}, {"file10.txt", false, false},
		}},
		{"bracket classes", []string{"[ab].js", "[!0-9]x"}, []check{
			{"a.js", false, true}, {"b.js", false, true}, {"c.js", false, false},
			{"yx", false, true}, {"5x", false, false},
		}},
		{"comments, blanks, escapes and trailing spaces", []string{"# dist", "", `\#hash`, `\!bang`, "spaced   ", `keep\ `, "crlf\r"}, []check{
			{"dist", true, false}, {"#hash", false, true}, {"!bang", false, true},
			{"spaced", false, true}, {"keep ", false, true}, {"keep", false, false},
			{"crlf", false, true},
		}},
		{"case-insensitive, like the extension", []string{"Dist"}, []check{
			{"dist", true, true}, {"DIST", true, true},
		}},
		{"defaults come first and can be re-included", []string{"!yarn.lock"}, []check{
			{"node_modules", true, true}, {"packages/app/node_modules", true, true},
			{".git", true, true}, {".github", true, true}, {".vscode", true, true},
			{"package-lock.json", false, true}, {"pnpm-lock.yaml", false, true},
			{"yarn.lock", false, false}, {"src/index.js", false, false},
		}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m := New(c.rules...)
			for _, ck := range c.checks {
				if got := m.Match(ck.path, ck.isDir); got != ck.want {
					t.Errorf("rules %q: Match(%q, dir=%v) = %v, want %v", c.rules, ck.path, ck.isDir, got, ck.want)
				}
			}
		})
	}
}

func TestLoad(t *testing.T) {
	write := func(t *testing.T, dir, name, content string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("squarecloud.ignore wins over legacy files", func(t *testing.T) {
		dir := t.TempDir()
		write(t, dir, FileName, "\xef\xbb\xbfsecret.txt\r\n")
		write(t, dir, ".squareignore", "other.txt\n")

		m, legacy, err := Load(dir)
		if err != nil || legacy != "" {
			t.Fatalf("Load = %v, legacy %q", err, legacy)
		}
		if !m.Match("secret.txt", false) || m.Match("other.txt", false) {
			t.Fatal("rules must come from squarecloud.ignore only")
		}
	})

	t.Run("legacy file is read and reported", func(t *testing.T) {
		dir := t.TempDir()
		write(t, dir, "square.ignore", "late.txt\n")
		write(t, dir, ".squareignore", "early.txt\n")

		m, legacy, err := Load(dir)
		if err != nil || legacy != ".squareignore" {
			t.Fatalf("Load = %v, legacy %q", err, legacy)
		}
		if !m.Match("early.txt", false) || m.Match("late.txt", false) {
			t.Fatal("the first legacy name must win")
		}
	})

	t.Run("no file: defaults only", func(t *testing.T) {
		m, legacy, err := Load(t.TempDir())
		if err != nil || legacy != "" {
			t.Fatalf("Load = %v, legacy %q", err, legacy)
		}
		if !m.Match("node_modules", true) || m.Match("index.js", false) {
			t.Fatal("defaults not applied")
		}
	})
}
