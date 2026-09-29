package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/i18n"
	"github.com/squarecloudofc/cli/internal/build"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/cmdutil"
	"github.com/squarecloudofc/cli/internal/command"
	"github.com/squarecloudofc/cli/internal/config"
	"github.com/squarecloudofc/cli/internal/ui"
	"github.com/squarecloudofc/cli/internal/updater"
	"golang.org/x/term"
)

func newSquareCloudCommand(squareCli cli.SquareCLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "squarecloud COMMAND",
		Short:             squareCli.I18n().T("metadata.commands.root.short"),
		SilenceErrors:     true,
		SilenceUsage:      true,
		TraverseChildren:  true,
		ValidArgsFunction: cobra.NoFileCompletions,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd:   false,
			HiddenDefaultCmd:    true,
			DisableDescriptions: true,
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			return errors.New(squareCli.I18n().T("errors.common.not_a_command", map[string]any{"Command": args[0]}))
		},
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// --json is global; a command that has no JSON output says so
			// instead of printing text a script would fail to parse.
			if cmdutil.WantsJSON(cmd) && cmd.Annotations["json"] != "true" {
				return errors.New(squareCli.I18n().T("errors.common.json_unsupported", map[string]any{"Command": cmd.CommandPath()}))
			}

			if cmd.Parent() != nil {
				if cmd.Parent().Name() != "completion" && cli.ShouldCheckAuth(squareCli, cmd) && !cli.CheckAuth(squareCli) {
					fmt.Fprintf(squareCli.Err(), "%s %s\n", ui.XMark, squareCli.I18n().T("errors.common.not_logged"))
					return &cli.AuthError{}
				}
			}

			return nil
		},
		Version: fmt.Sprintf("%s, commit %s, commited at %s", build.Version, build.Commit, build.CommitDate),
	}

	cmd.SetOut(squareCli.Out())
	cmd.SetIn(squareCli.In())
	cmd.SetErr(squareCli.Err())

	cmd.SetVersionTemplate("Square Cloud CLI version {{.Version}}\n")
	cmd.Flags().BoolP("version", "v", false, squareCli.I18n().T("metadata.flags.version"))
	cmd.PersistentFlags().Bool("json", false, squareCli.I18n().T("metadata.flags.json"))

	// Cobra's help flag, help command and usage headings are English; these
	// replace their text only. The help flag is cobra's (same name, shorthand
	// and annotation), declared once here for every command to inherit.
	cmd.PersistentFlags().BoolP("help", "h", false, squareCli.I18n().T("metadata.help.flag"))
	_ = cmd.PersistentFlags().SetAnnotation("help", cobra.FlagSetByCobraAnnotation, []string{"true"})
	cmd.SetUsageTemplate(translateUsage(squareCli.I18n(), cmd.UsageTemplate()))

	command.AddCommands(cmd, squareCli)

	cmd.InitDefaultHelpCmd() // now rather than at Execute, to translate it
	for _, c := range cmd.Commands() {
		if c.Name() == "help" {
			c.Short, c.Long = squareCli.I18n().T("metadata.help.command"), ""
		}
	}
	return cmd
}

// translateUsage swaps the English headings of cobra's usage template for
// translated ones, leaving its logic as is.
func translateUsage(t i18n.Localizer, template string) string {
	return strings.NewReplacer(
		"Usage:", t.T("metadata.help.usage"),
		"Aliases:", t.T("metadata.help.aliases"),
		"Examples:", t.T("metadata.help.examples"),
		"Available Commands:", t.T("metadata.help.available_commands"),
		"Additional Commands:", t.T("metadata.help.additional_commands"),
		"Global Flags:", t.T("metadata.help.global_flags"),
		"Flags:", t.T("metadata.help.flags"),
		"Additional help topics:", t.T("metadata.help.additional_topics"),
		`Use "{{.CommandPath}} [command] --help" for more information about a command.`,
		t.T("metadata.help.more_info", map[string]any{"Command": "{{.CommandPath}}"}),
	).Replace(template)
}

func main() {
	squareCli, err := cli.NewSquareCli()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	// The first Ctrl+C cancels ctx; restoring the default lets a second one
	// kill a command blocked where ctx is not checked (a prompt, stdin).
	context.AfterFunc(ctx, stop)

	// The update check runs beside the command; nil when skipped. Only for a
	// person at a terminal: CI logs and pipes would just lose the banner.
	var latestVersion chan string
	if dir := config.ConfigDir(); dir != "" && term.IsTerminal(int(os.Stderr.Fd())) && wantsUpdateCheck(os.Args[1:]) {
		latestVersion = make(chan string, 1)
		go func() {
			latestVersion <- updater.Latest(ctx, updater.LatestReleaseURL, filepath.Join(dir, updater.CacheFileName), time.Now())
		}()
	}

	cmd := newSquareCloudCommand(squareCli)
	if err := cmd.ExecuteContext(ctx); err != nil {
		cli.PrintError(squareCli, err)
		os.Exit(1)
	}

	if latestVersion == nil {
		return
	}
	// stderr: on stdout the banner would corrupt --json output and generated
	// completion scripts.
	if release := <-latestVersion; updater.IsOutdated(build.Version, release) {
		fmt.Fprintln(squareCli.Err(), "")
		fmt.Fprintln(squareCli.Err(), ui.TextYellow.Render(squareCli.I18n().T("ui.update.outdated", map[string]any{"Version": build.Version})))
		fmt.Fprintf(squareCli.Err(), " %s\n", squareCli.I18n().T("ui.update.hint", map[string]any{"Version": ui.TextGreen.Render(release)}))
	}
}

// wantsUpdateCheck is false for runs where the banner is noise: shell
// completion (it runs on every TAB), help and --version.
func wantsUpdateCheck(args []string) bool {
	for _, arg := range args {
		switch arg {
		case cobra.ShellCompRequestCmd, cobra.ShellCompNoDescRequestCmd, "completion", "help", "--help", "-h", "--version", "-v":
			return false
		}
	}
	return true
}
