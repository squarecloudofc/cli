package main

import (
	"context"
	"fmt"
	"os"

	"github.com/google/go-github/v58/github"
	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/build"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/command"
	"github.com/squarecloudofc/cli/internal/rest"
	"github.com/squarecloudofc/cli/internal/ui"
)

func newSquareCloudCommand(squareCli *cli.SquareCli) *cobra.Command {
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
				if cli.ShouldCheckAuth(squareCli, cmd) && !cli.CheckAuth(squareCli) {
					return fmt.Errorf("you must be logged to execute this command, try to execute: squarecloud login")
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

	cmd.Flags().BoolP("debug", "d", false, "Debug Mode")
	cmd.Flags().MarkHidden("debug")

	command.AddCommands(cmd, squareCli)
	return cmd
}

func run(context context.Context, squareCli *cli.SquareCli) (err error) {
	cmd := newSquareCloudCommand(squareCli)

	return cmd.ExecuteContext(context)
}

func main() {
	squareCli := cli.NewSquareCli()

	ctx := context.Background()

	updateContext, updateCancel := context.WithCancel(ctx)
	defer updateCancel()

	updateMessageChannel := make(chan *github.RepositoryRelease)
	go func() {
		client := github.NewClient(nil)
		release, _, _ := client.Repositories.GetLatestRelease(updateContext, "squarecloudofc", "cli")

		updateMessageChannel <- release
	}()

	if err := run(ctx, squareCli); err != nil {
		switch err.(type) {
		case rest.RestError:
			fmt.Fprintln(squareCli.Err(), err)
			os.Exit(0)
		default:
			fmt.Fprintln(squareCli.Err(), err)
			os.Exit(1)
		}
	}

	updateCancel()
	release := <-updateMessageChannel
	if release != nil && build.Version != *release.TagName {
		version := ui.GreenText.SetString(*release.TagName)

		fmt.Fprintln(squareCli.Out(), "")
		fmt.Fprintln(squareCli.Out(), ui.YellowText.SetString("You're using an old version of Square Cloud CLI: "+build.Version))
		fmt.Fprintf(squareCli.Out(), " Please update to %s\n", version)
	}
}
