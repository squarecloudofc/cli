package command

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
)

func NewStopCommand(squareCli *cli.SquareCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "Stop",
		Short: "stop your application",
		RunE:  runStopCommand(squareCli),
	}

	return cmd
}

func runStopCommand(squareCli *cli.SquareCli) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) (err error) {
		appId := args[0]
		success, err := squareCli.Rest.ApplicationStop(appId)
		if err != nil {
			return
		}

		if success {
			fmt.Fprintln(squareCli.Out(), "Your application has been successfuly stopped")
		} else {
			fmt.Fprintln(squareCli.Out(), "Failed to stop your application")
		}

		return nil
	}
}
