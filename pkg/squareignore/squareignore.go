// Package squareignore reads squarecloud.ignore, the paths a project upload
// leaves out. It is the file the Square Cloud VS Code extension reads, with
// the same gitignore syntax and the same defaults.
package squareignore

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// FileName is the ignore file in the project root.
const FileName = "squarecloud.ignore"

// LegacyFileNames are read, in this order, only when FileName is missing.
var LegacyFileNames = []string{".squareignore", ".squarecloudignore", "square.ignore"}

// Defaults are applied before the file's rules, so "!yarn.lock" re-includes
// one. Keep in sync with the extension's resources/squarecloud.ignore.
var Defaults = []string{"node_modules", ".git", ".github", ".vscode", "package-lock.json", "pnpm-lock.yaml", "yarn.lock"}

// Matcher is a compiled rule list; the last rule matching a path decides.
type Matcher struct{ rules []rule }

type rule struct {
	re      *regexp.Regexp
	negate  bool
	dirOnly bool
}

// Load reads FileName, or else the first legacy file, from dir. legacy names
// the legacy file used, if any, so the caller can suggest renaming it.
func Load(dir string) (m *Matcher, legacy string, err error) {
	for _, name := range append([]string{FileName}, LegacyFileNames...) {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, "", err
		}
		if name != FileName {
			legacy = name
		}
		content := strings.TrimPrefix(string(data), "\xef\xbb\xbf") // BOM from Windows editors
		return New(strings.Split(content, "\n")...), legacy, nil
	}
	return New(), "", nil
}

// New compiles Defaults followed by lines, each a gitignore pattern.
func New(lines ...string) *Matcher {
	m := &Matcher{}
	for _, line := range append(append([]string{}, Defaults...), lines...) {
		if r, ok := parse(line); ok {
			m.rules = append(m.rules, r)
		}
	}
	return m
}

// Match reports whether rel, a slash-separated path relative to the project
// root, is ignored. The caller walks top-down and skips ignored directories
// whole, so, like git, nothing inside one can be re-included.
func (m *Matcher) Match(rel string, isDir bool) bool {
	ignored := false
	for _, r := range m.rules {
		if (isDir || !r.dirOnly) && r.re.MatchString(rel) {
			ignored = !r.negate
		}
	}
	return ignored
}

// parse compiles one gitignore line; ok is false for blanks and comments.
func parse(line string) (r rule, ok bool) {
	line = strings.TrimSuffix(line, "\r")
	for strings.HasSuffix(line, " ") && !strings.HasSuffix(line, `\ `) {
		line = line[:len(line)-1] // trailing spaces, unless escaped
	}
	if line == "" || line[0] == '#' {
		return r, false
	}

	if line[0] == '!' {
		r.negate, line = true, line[1:]
	}
	if strings.HasSuffix(line, "/") {
		r.dirOnly, line = true, strings.TrimRight(line, "/")
	}
	// A slash at the start or in the middle anchors the pattern to the root;
	// without one it matches at any depth.
	if !strings.Contains(line, "/") {
		line = "**/" + line
	}
	line = strings.TrimPrefix(line, "/")
	if line == "" {
		return r, false
	}

	r.re = regexp.MustCompile(translate(line))
	return r, true
}

// translate turns a gitignore glob into a regexp. Case-insensitive, like the
// extension's `ignore` package.
func translate(glob string) string {
	var b strings.Builder
	b.WriteString("(?i)^")
	for i := 0; i < len(glob); i++ {
		atSegment := i == 0 || glob[i-1] == '/'
		switch c := glob[i]; {
		case atSegment && strings.HasPrefix(glob[i:], "**/"):
			b.WriteString("(?:.*/)?") // zero or more directories
			i += 2
		case atSegment && glob[i:] == "**":
			b.WriteString(".*") // everything inside
			i++
		case c == '*':
			b.WriteString("[^/]*")
		case c == '?':
			b.WriteString("[^/]")
		case c == '[':
			class, n := bracket(glob[i:])
			if n == 0 {
				b.WriteString(`\[`) // never closed: a literal
				continue
			}
			b.WriteString(class)
			i += n - 1
		case c == '\\' && i+1 < len(glob):
			i++
			b.WriteString(regexp.QuoteMeta(glob[i : i+1]))
		default:
			b.WriteString(regexp.QuoteMeta(glob[i : i+1]))
		}
	}
	b.WriteString("$")
	return b.String()
}

// bracket translates the [...] class at the start of s and returns it with
// its length in s, or n == 0 when the class is not closed.
// ponytail: no POSIX classes ([[:alpha:]]); add if anyone asks.
func bracket(s string) (class string, n int) {
	i := 1
	negate := i < len(s) && (s[i] == '!' || s[i] == '^')
	if negate {
		i++
	}
	start := i
	if i < len(s) && s[i] == ']' {
		i++ // a leading ] is a member
	}
	for i < len(s) && s[i] != ']' {
		if s[i] == '\\' {
			i++
		}
		i++
	}
	if i >= len(s) {
		return "", 0
	}

	var b strings.Builder
	b.WriteString("[")
	if negate {
		b.WriteString("^/")
	}
	for j := start; j < i; j++ {
		c := s[j]
		escaped := c == '\\' && j+1 < i
		if escaped {
			j++
			c = s[j]
		}
		if c < 0x80 && strings.IndexByte(`\]-[^!.*?$+(){}|/#`, c) >= 0 && (escaped || c != '-') {
			b.WriteByte('\\')
		}
		b.WriteByte(c)
	}
	b.WriteString("]")
	return b.String(), i + 1
}
