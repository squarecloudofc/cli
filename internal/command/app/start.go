package app

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
)

func NewStartCommand(squareCli *cli.SquareCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start your application",
		RunE:  runStartCommand(squareCli),
	}

	return cmd
}

func runStartCommand(squareCli *cli.SquareCli) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) (err error) {
		appId := args[0]
		success, err := squareCli.Rest.ApplicationStart(appId)
		if err != nil {
			return
		}

		if success {
			fmt.Fprintln(squareCli.Out(), "Your application has been successfuly started")
		} else {
			fmt.Fprintln(squareCli.Out(), "Failed to start your application")
		}

		return nil
	}
}
