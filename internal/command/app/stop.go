package app

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
)

func NewStopCommand(squareCli *cli.SquareCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "stop your application",
		RunE:  runStopCommand(squareCli),
	}

	return cmd
}

func runStopCommand(squareCli *cli.SquareCli) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) (err error) {
		var appId string

		if len(args) > 0 {
			appId = args[0]
		}

		if len(args) < 1 {
			id, err := CreateApplicationSelection(squareCli)
			if err != nil {
				return err
			}

			appId = id
		}

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
