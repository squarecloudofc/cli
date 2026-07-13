package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/build"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/command"
	"github.com/squarecloudofc/cli/internal/ui"
	"github.com/squarecloudofc/cli/internal/updater"
)

func newSquareCloudCommand(squareCli cli.SquareCLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "squarecloud COMMAND",
		Short:             "A command line application to manage your Square Cloud applications",
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
			return fmt.Errorf("%s is not a command. See 'squarecloud --help'", args[0])
		},
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Parent() != nil {
				if cmd.Parent().Name() != "completion" && cli.ShouldCheckAuth(squareCli, cmd) && !cli.CheckAuth(squareCli) {
					fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.XMark, squareCli.I18n().T("errors.common.not_logged"))
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
	cmd.Flags().BoolP("version", "v", false, "Print CLI version")
	cmd.PersistentFlags().Bool("json", false, "Output raw JSON")

	command.AddCommands(cmd, squareCli)
	return cmd
}

func main() {
	squareCli, err := cli.NewSquareCli()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	latestVersion := make(chan string, 1)
	go func() {
		latestVersion <- updater.LatestVersion(ctx)
	}()

	cmd := newSquareCloudCommand(squareCli)
	if err := cmd.ExecuteContext(ctx); err != nil {
		var authErr *cli.AuthError
		if !errors.As(err, &authErr) {
			fmt.Fprintln(squareCli.Err(), err)
		}
		os.Exit(1)
	}

	if release := <-latestVersion; build.Version != "development" && release != "" && release != build.Version {
		fmt.Fprintln(squareCli.Out(), "")
		fmt.Fprintln(squareCli.Out(), ui.TextYellow.SetString("You're using an old version of Square Cloud CLI: "+build.Version))
		fmt.Fprintf(squareCli.Out(), " Please update to %s\n", ui.TextGreen.SetString(release))
	}
}
