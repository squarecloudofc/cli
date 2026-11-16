package command

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/squarecloudofc/cli/i18n"
	"github.com/squarecloudofc/cli/internal/config"
	"github.com/squarecloudofc/cli/pkg/squareconfig"
	"github.com/squarecloudofc/cli/pkg/squareignore"
	"github.com/squarecloudofc/sdk-api-go/v4"
)

type fakeCLI struct{ out, err bytes.Buffer }

func (f *fakeCLI) Config() *config.Config   { return &config.Config{} }
func (f *fakeCLI) API() *squarecloud.Client { return nil }
func (f *fakeCLI) I18n() i18n.Localizer     { return i18n.NewLocalizer("en") }
func (f *fakeCLI) Err() io.Writer           { return &f.err }
func (f *fakeCLI) In() io.ReadCloser        { return io.NopCloser(strings.NewReader("")) } // not a terminal
func (f *fakeCLI) Out() io.Writer           { return &f.out }

// writeFiles creates files (path -> content) under dir.
func writeFiles(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for name, content := range files {
		path := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestDetectMain(t *testing.T) {
	cases := []struct {
		name     string
		files    map[string]string
		wantMain string
		wantSite bool
	}{
		{"package.json main", map[string]string{"package.json": `{"main":"./src/bot.js"}`, "src/bot.js": "", "index.js": ""}, "src/bot.js", false},
		{"package.json main missing on disk", map[string]string{"package.json": `{"main":"dist/index.js"}`, "index.js": ""}, "index.js", false},
		{"package.json main outside the project", map[string]string{"package.json": `{"main":"../x.js"}`, "index.js": ""}, "index.js", false},
		{"package.json without main", map[string]string{"package.json": `{}`, "src/index.ts": ""}, "src/index.ts", false},
		{"express is a site", map[string]string{"package.json": `{"dependencies":{"express":"^5"}}`, "server.js": ""}, "server.js", true},
		{"next as a dev dependency", map[string]string{"package.json": `{"devDependencies":{"next":"15"}}`, "index.js": ""}, "index.js", true},
		{"broken package.json", map[string]string{"package.json": `{`, "main.js": ""}, "main.js", false},
		{"Node.js wins over Python", map[string]string{"package.json": `{}`, "index.js": "", "main.py": ""}, "index.js", false},
		// A package.json without a JS entry is tooling: detection goes on.
		{"flask app with a tooling package.json", map[string]string{"package.json": `{"devDependencies":{"prettier":"3"}}`, "app.py": "", "requirements.txt": "flask\n"}, "app.py", true},
		{"static site with a package.json", map[string]string{"package.json": `{"devDependencies":{"tailwindcss":"4"}}`, "index.html": ""}, "index.html", true},
		{"go with a package.json", map[string]string{"package.json": `{}`, "main.go": ""}, "main.go", false},
		{"node deps without a JS entry make no site", map[string]string{"package.json": `{"dependencies":{"express":"5"}}`, "main.py": ""}, "main.py", false},
		{"python bot", map[string]string{"main.py": "", "requirements.txt": "discord.py>=2.3\n"}, "main.py", false},
		{"flask site", map[string]string{"app.py": "", "requirements.txt": "# web\nFlask[async]==3.0 ; python_version > '3.8'\n\n"}, "app.py", true},
		{"python without requirements", map[string]string{"bot.py": ""}, "bot.py", false},
		{"go", map[string]string{"main.go": "", "go.mod": ""}, "main.go", false},
		{"index.js alone", map[string]string{"index.js": ""}, "index.js", false},
		{"static site", map[string]string{"index.html": ""}, "index.html", true},
		{"nothing to run", map[string]string{"README.md": ""}, "", false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			writeFiles(t, dir, c.files)
			main, site := detectMain(dir)
			if main != c.wantMain || site != c.wantSite {
				t.Errorf("detectMain = (%q, %v), want (%q, %v)", main, site, c.wantMain, c.wantSite)
			}
		})
	}
}

func TestSubdomainFrom(t *testing.T) {
	for name, want := range map[string]string{
		"My Bot_2":                     "my-bot-2",
		"--site--":                     "site",
		"日本語":                          "",
		strings.Repeat("a", 70):        strings.Repeat("a", 63),
		strings.Repeat("a", 62) + "-b": strings.Repeat("a", 62),
	} {
		if got := subdomainFrom(name); got != want {
			t.Errorf("subdomainFrom(%q) = %q, want %q", name, got, want)
		}
	}
}

func runInit(t *testing.T, dir string, args ...string) (*fakeCLI, error) {
	t.Helper()
	t.Chdir(dir)
	f := &fakeCLI{}
	cmd := NewInitCommand(f)
	cmd.SetArgs(args)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	return f, cmd.Execute()
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestInitYes(t *testing.T) {
	en := i18n.NewLocalizer("en")
	dir := filepath.Join(t.TempDir(), "My Site")
	writeFiles(t, dir, map[string]string{"package.json": `{"dependencies":{"express":"^5"}}`, "index.js": ""})
	name := squareconfig.GetDefaultConfig()

	if _, err := runInit(t, dir, "--yes"); err != nil {
		t.Fatalf("init --yes = %v", err)
	}
	want := "MAIN=index.js\nMEMORY=512\nVERSION=recommended\nDISPLAY_NAME=My Site\nSUBDOMAIN=my-site\n"
	if got := readFile(t, filepath.Join(dir, name)); got != want {
		t.Errorf("%s =\n%s\nwant\n%s", name, got, want)
	}
	if got := readFile(t, filepath.Join(dir, squareignore.FileName)); got != ignoreTemplate {
		t.Errorf("%s = %q", squareignore.FileName, got)
	}

	// A second run aborts; --force overwrites, and keeps a user's ignore file.
	writeFiles(t, dir, map[string]string{squareignore.FileName: "mine\n"})
	if _, err := runInit(t, dir, "--yes"); err == nil || err.Error() != en.T("commands.init.exists", map[string]any{"File": name}) {
		t.Fatalf("init over an existing file = %v", err)
	}
	writeFiles(t, dir, map[string]string{name: "ID=keep\n"})
	if _, err := runInit(t, dir, "--yes", "--force"); err != nil {
		t.Fatalf("init --force = %v", err)
	}
	if got := readFile(t, filepath.Join(dir, name)); got != want {
		t.Errorf("after --force %s =\n%s", name, got)
	}
	if got := readFile(t, filepath.Join(dir, squareignore.FileName)); got != "mine\n" {
		t.Errorf("--force replaced the ignore file: %q", got)
	}
}

func TestInitForceKeepsTheExistingName(t *testing.T) {
	dir := t.TempDir()
	other := squareconfig.SquareConfigNames[0]
	if other == squareconfig.GetDefaultConfig() {
		other = squareconfig.SquareConfigNames[1]
	}
	writeFiles(t, dir, map[string]string{"main.py": "", other: "MAIN=old.py\n"})

	if _, err := runInit(t, dir, "--yes", "--force"); err != nil {
		t.Fatalf("init --force = %v", err)
	}
	if !strings.HasPrefix(readFile(t, filepath.Join(dir, other)), "MAIN=main.py\n") {
		t.Errorf("%s was not overwritten", other)
	}
	if fileExists(filepath.Join(dir, squareconfig.GetDefaultConfig())) {
		t.Errorf("--force wrote a second config file")
	}
}

// TestInitForceOverwritesTheFileTheCLIReads: --force rewrites the config file
// squareconfig resolves, never one that file shadows.
func TestInitForceOverwritesTheFileTheCLIReads(t *testing.T) {
	read, shadowed := squareconfig.SquareConfigNames[1], squareconfig.SquareConfigNames[0]

	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{"main.py": "", shadowed: "MAIN=old.py\n", read: "MAIN=old.py\n"})
	if name, found := squareconfig.FileIn(dir); name != read || !found {
		t.Fatalf("FileIn = %q, %v; want %q, true", name, found, read)
	}
	if _, err := runInit(t, dir, "--yes", "--force"); err != nil {
		t.Fatalf("init --force = %v", err)
	}
	if !strings.HasPrefix(readFile(t, filepath.Join(dir, read)), "MAIN=main.py\n") || readFile(t, filepath.Join(dir, shadowed)) != "MAIN=old.py\n" {
		t.Errorf("--force did not rewrite %s alone", read)
	}

	// The name read is not a file: an error, not a write to the shadowed one.
	dir = t.TempDir()
	writeFiles(t, dir, map[string]string{"main.py": "", shadowed: "MAIN=old.py\n"})
	if err := os.Mkdir(filepath.Join(dir, read), 0755); err != nil {
		t.Fatal(err)
	}
	if _, err := runInit(t, dir, "--yes", "--force"); err == nil {
		t.Error("init --force over a folder succeeded")
	}
	if got := readFile(t, filepath.Join(dir, shadowed)); got != "MAIN=old.py\n" {
		t.Errorf("--force wrote the shadowed %s: %q", shadowed, got)
	}
}

func TestInitRefusals(t *testing.T) {
	en := i18n.NewLocalizer("en")
	dir := t.TempDir()
	if _, err := runInit(t, dir, "--yes"); err == nil || err.Error() != en.T("commands.init.no_main") {
		t.Errorf("init --yes without a main file = %v", err)
	}
	if _, err := runInit(t, dir); err == nil || err.Error() != en.T("errors.confirm.no_tty") {
		t.Errorf("init without a terminal = %v", err)
	}
	if fileExists(filepath.Join(dir, squareconfig.GetDefaultConfig())) {
		t.Error("a refused init wrote the config file")
	}
}

func TestAskConfig(t *testing.T) {
	en := i18n.NewLocalizer("en")
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{"main.py": "", "src/app.py": ""})
	detected := appConfig{Main: "main.py", DisplayName: "bot", Subdomain: "bot", Memory: minMemoryBot}

	answers := strings.Join([]string{
		"missing.py", "./src/app.py", // main: rejected, then accepted
		"",            // name: keep the default
		"y",           // a site
		"300", "1024", // memory below the site minimum, then valid
		"no_way", "", // subdomain: rejected, then the default
		"python app.py", // start
		"",              // autorestart: default no
	}, "\n") + "\n"

	var prompts bytes.Buffer
	p := prompter{r: bufio.NewReader(strings.NewReader(answers)), w: &prompts, t: en}
	cfg, err := p.askConfig(dir, detected)
	if err != nil {
		t.Fatalf("askConfig = %v", err)
	}
	want := appConfig{Main: "src/app.py", DisplayName: "bot", Subdomain: "bot", Start: "python app.py", Memory: 1024, Site: true}
	if cfg != want {
		t.Errorf("askConfig = %+v, want %+v", cfg, want)
	}
	for _, msg := range []string{
		en.T("commands.init.invalid.main"),
		en.T("commands.init.invalid.memory", map[string]any{"Min": minMemorySite}),
		en.T("commands.init.invalid.subdomain"),
		en.T("commands.init.ask.memory", map[string]any{"Min": minMemorySite}) + " [512]",
	} {
		if !strings.Contains(prompts.String(), msg) {
			t.Errorf("prompts lack %q:\n%s", msg, prompts.String())
		}
	}

	// Input that ends early is an error, not a half-made file.
	p = prompter{r: bufio.NewReader(strings.NewReader("\n")), w: io.Discard, t: en}
	if _, err := p.askConfig(dir, detected); err == nil || err.Error() != en.T("errors.confirm.no_tty") {
		t.Errorf("askConfig at EOF = %v", err)
	}
}
