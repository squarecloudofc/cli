package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/build"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/command"
)

func newSquareCloudCommand(squareCli *cli.SquareCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:              "squarecloud COMMAND",
		Short:            "A command line application to manage your Square Cloud applications",
		SilenceErrors:    true,
		SilenceUsage:     true,
		TraverseChildren: true,
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
		Version: fmt.Sprintf("%s, commit %s, built at %s", build.Version, build.GitCommit, build.BuildTime),
	}

	cmd.SetOut(squareCli.Out())
	cmd.SetIn(squareCli.In())
	cmd.SetErr(squareCli.Err())

	cmd.SetVersionTemplate("SquareCloud CLI version {{.Version}}\n")
	cmd.Flags().BoolP("version", "v", false, "Print CLI version")

	command.AddCommands(cmd, squareCli)
	return cmd
}

func run(squareCli *cli.SquareCli) (err error) {
	cmd := newSquareCloudCommand(squareCli)

	return cmd.Execute()
}

func main() {
	squareCli := cli.NewSquareCli()

	if err := run(squareCli); err != nil {
		fmt.Fprintln(squareCli.Err(), err)
		os.Exit(1)
	}
}
