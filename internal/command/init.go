package command

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/i18n"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/cmdutil"
	"github.com/squarecloudofc/cli/internal/ui"
	"github.com/squarecloudofc/cli/pkg/squareconfig"
	"github.com/squarecloudofc/cli/pkg/squareignore"
)

// Minimum MEMORY, in MB, the platform accepts for each kind of app.
const (
	minMemoryBot  = 256
	minMemorySite = 512
)

// ignoreTemplate is the squarecloud.ignore init offers; squareignore.Defaults
// already apply on top of it.
const ignoreTemplate = "*.log\n.venv/\nvenv/\n__pycache__/\n"

var (
	subdomainRe    = regexp.MustCompile(`^[a-zA-Z0-9-]{1,63}$`)
	notSubdomainRe = regexp.MustCompile(`[^a-z0-9]+`)
)

// appConfig is the squarecloud.app init writes.
type appConfig struct {
	Main, DisplayName, Subdomain, Start string
	Memory                              int
	Site, AutoRestart                   bool
}

func (c appConfig) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "MAIN=%s\nMEMORY=%d\nVERSION=recommended\nDISPLAY_NAME=%s\n", c.Main, c.Memory, c.DisplayName)
	if c.Site && c.Subdomain != "" {
		fmt.Fprintf(&b, "SUBDOMAIN=%s\n", c.Subdomain)
	}
	if c.Start != "" {
		fmt.Fprintf(&b, "START=%s\n", c.Start)
	}
	if c.AutoRestart {
		b.WriteString("AUTORESTART=true\n")
	}
	return b.String()
}

func NewInitCommand(squareCli cli.SquareCLI) *cobra.Command {
	var yes, force bool

	cmd := &cobra.Command{
		Use:         "init",
		Short:       squareCli.I18n().T("metadata.commands.init.short"),
		Annotations: map[string]string{"skipAuthCheck": "true"},
		Args:        cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			t := squareCli.I18n()
			dir, err := os.Getwd()
			if err != nil {
				return err
			}

			// The file the CLI reads: --force overwrites that one, never a
			// shadowed one or a second file.
			name, exists := squareconfig.FileIn(dir)
			if exists && !force {
				return errors.New(t.T("commands.init.exists", map[string]any{"File": name}))
			}

			cfg := detectConfig(dir)
			ignore := !fileExists(filepath.Join(dir, squareignore.FileName))
			if yes {
				if cfg.Main == "" {
					return errors.New(t.T("commands.init.no_main"))
				}
			} else {
				if !cmdutil.IsTerminal(squareCli.In()) {
					return errors.New(t.T("errors.confirm.no_tty"))
				}
				p := prompter{r: bufio.NewReader(squareCli.In()), w: squareCli.Err(), t: t}
				if cfg, err = p.askConfig(dir, cfg); err != nil {
					return err
				}
				if ignore {
					if ignore, err = p.yesNo(t.T("commands.init.ask.ignore", map[string]any{"File": squareignore.FileName}), true); err != nil {
						return err
					}
				}
			}

			if err := os.WriteFile(filepath.Join(dir, name), []byte(cfg.String()), 0644); err != nil {
				return err
			}
			fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, t.T("commands.init.created", map[string]any{"File": name}))

			if ignore {
				if err := os.WriteFile(filepath.Join(dir, squareignore.FileName), []byte(ignoreTemplate), 0644); err != nil {
					return err
				}
				fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, t.T("commands.init.created", map[string]any{"File": squareignore.FileName}))
			}

			fmt.Fprintln(squareCli.Out(), t.T("commands.init.next"))
			return nil
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, squareCli.I18n().T("metadata.flags.yes_init"))
	cmd.Flags().BoolVar(&force, "force", false, squareCli.I18n().T("metadata.flags.force_init"))
	return cmd
}

// detectConfig fills the defaults for the project in dir: the main file and
// whether it is a site, the folder name, and the minimum memory.
func detectConfig(dir string) appConfig {
	main, site := detectMain(dir)
	name := filepath.Base(dir)
	if len([]rune(name)) > 32 {
		name = string([]rune(name)[:32])
	}
	cfg := appConfig{Main: main, Site: site, DisplayName: name, Subdomain: subdomainFrom(name), Memory: minMemoryBot}
	if site {
		cfg.Memory = minMemorySite
	}
	return cfg
}

// Web frameworks that make a project a site, by package manager.
var (
	nodeSiteDeps   = []string{"express", "fastify", "koa", "next", "nuxt", "@nestjs/core", "hono"}
	pythonSiteDeps = []string{"flask", "django", "fastapi", "uvicorn", "gunicorn"}
)

// detectMain finds the main file of the project in dir ("" when none) and
// whether it looks like a site. First match wins: Node.js, Python, Go, then
// a static site.
func detectMain(dir string) (main string, site bool) {
	first := func(names ...string) string {
		for _, n := range names {
			if fileExists(filepath.Join(dir, n)) {
				return n
			}
		}
		return ""
	}

	if data, err := os.ReadFile(filepath.Join(dir, "package.json")); err == nil {
		var pkg struct {
			Main            string            `json:"main"`
			Dependencies    map[string]string `json:"dependencies"`
			DevDependencies map[string]string `json:"devDependencies"`
		}
		_ = json.Unmarshal(data, &pkg) // a broken package.json still says Node.js
		for _, dep := range nodeSiteDeps {
			_, inDeps := pkg.Dependencies[dep]
			_, inDev := pkg.DevDependencies[dep]
			site = site || inDeps || inDev
		}
		main = filepath.ToSlash(filepath.Clean(pkg.Main))
		if pkg.Main == "" || !filepath.IsLocal(main) || !fileExists(filepath.Join(dir, main)) {
			main = first("index.js", "index.ts", "main.js", "main.ts", "bot.js", "app.js", "server.js", "src/index.js", "src/index.ts")
		}
		if main != "" {
			return main, site
		}
		site = false // no JS entry: the package.json is tooling, go on
	}

	if main = first("main.py", "app.py", "bot.py", "index.py", "src/main.py"); main != "" {
		reqs, _ := os.ReadFile(filepath.Join(dir, "requirements.txt"))
		for _, line := range strings.Split(strings.ToLower(string(reqs)), "\n") {
			pkg := strings.TrimSpace(line) // "flask[async]>=3 ; python_version..." -> flask
			if i := strings.IndexAny(pkg, "=<>~![; "); i >= 0 {
				pkg = pkg[:i]
			}
			site = site || slices.Contains(pythonSiteDeps, pkg)
		}
		return main, site
	}

	if main = first("main.go", "index.js", "index.ts"); main != "" {
		return main, false
	}
	if main = first("index.html"); main != "" {
		return main, true
	}
	return "", false
}

// subdomainFrom turns a folder name into a valid subdomain, or "".
func subdomainFrom(name string) string {
	s := strings.Trim(notSubdomainRe.ReplaceAllString(strings.ToLower(name), "-"), "-")
	if len(s) > 63 {
		s = strings.Trim(s[:63], "-")
	}
	return s
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// prompter asks init's questions on w and reads the answers from r.
type prompter struct {
	r *bufio.Reader
	w io.Writer
	t i18n.Localizer
}

// ask shows question with its default and returns the answer, or def for an
// empty one. EOF is an error: there is nobody to answer.
func (p prompter) ask(question, def string) (string, error) {
	if def != "" {
		fmt.Fprintf(p.w, "%s [%s]: ", question, def)
	} else {
		fmt.Fprintf(p.w, "%s: ", question)
	}
	line, err := p.r.ReadString('\n')
	if err != nil && line == "" {
		fmt.Fprintln(p.w)
		return "", errors.New(p.t.T("errors.confirm.no_tty"))
	}
	if answer := strings.TrimSpace(line); answer != "" {
		return answer, nil
	}
	return def, nil
}

// askValid asks until valid accepts the answer, printing invalid after each
// rejected one.
func (p prompter) askValid(question, def, invalid string, valid func(string) bool) (string, error) {
	for {
		answer, err := p.ask(question, def)
		if err != nil || valid(answer) {
			return answer, err
		}
		fmt.Fprintf(p.w, "%s %s\n", ui.XMark, invalid)
	}
}

func (p prompter) yesNo(question string, def bool) (bool, error) {
	hint := "y/N"
	if def {
		hint = "Y/n"
	}
	for {
		fmt.Fprintf(p.w, "%s [%s]: ", question, hint)
		line, err := p.r.ReadString('\n')
		if err != nil && line == "" {
			fmt.Fprintln(p.w)
			return false, errors.New(p.t.T("errors.confirm.no_tty"))
		}
		switch strings.ToLower(strings.TrimSpace(line)) {
		case "":
			return def, nil
		case "y", "yes":
			return true, nil
		case "n", "no":
			return false, nil
		}
	}
}

// askConfig walks through the fields of cfg, its detected values as
// defaults.
func (p prompter) askConfig(dir string, cfg appConfig) (appConfig, error) {
	t := p.t
	var err error

	inProject := func(s string) bool {
		rel := filepath.Clean(s)
		return filepath.IsLocal(rel) && fileExists(filepath.Join(dir, rel))
	}
	if cfg.Main, err = p.askValid(t.T("commands.init.ask.main"), cfg.Main, t.T("commands.init.invalid.main"), inProject); err != nil {
		return cfg, err
	}
	cfg.Main = filepath.ToSlash(filepath.Clean(cfg.Main))

	if cfg.DisplayName, err = p.askValid(t.T("commands.init.ask.name"), cfg.DisplayName, t.T("commands.init.invalid.name"), func(s string) bool {
		return s != "" && len([]rune(s)) <= 32
	}); err != nil {
		return cfg, err
	}

	if cfg.Site, err = p.yesNo(t.T("commands.init.ask.site"), cfg.Site); err != nil {
		return cfg, err
	}
	minimum := minMemoryBot
	if cfg.Site {
		minimum = minMemorySite
	}
	memory, err := p.askValid(t.T("commands.init.ask.memory", map[string]any{"Min": minimum}), strconv.Itoa(max(cfg.Memory, minimum)),
		t.T("commands.init.invalid.memory", map[string]any{"Min": minimum}), func(s string) bool {
			n, err := strconv.Atoi(s)
			return err == nil && n >= minimum
		})
	if err != nil {
		return cfg, err
	}
	cfg.Memory, _ = strconv.Atoi(memory)

	if cfg.Site {
		if cfg.Subdomain, err = p.askValid(t.T("commands.init.ask.subdomain"), cfg.Subdomain, t.T("commands.init.invalid.subdomain"), subdomainRe.MatchString); err != nil {
			return cfg, err
		}
	}

	if cfg.Start, err = p.ask(t.T("commands.init.ask.start"), ""); err != nil {
		return cfg, err
	}
	cfg.AutoRestart, err = p.yesNo(t.T("commands.init.ask.autorestart"), false)
	return cfg, err
}
