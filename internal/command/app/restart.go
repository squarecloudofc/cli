package command

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
)

func NewRestartCommand(squareCli *cli.SquareCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restart",
		Short: "Restart your application",
		RunE:  runRestartCommand(squareCli),
	}

	return cmd
}

func runRestartCommand(squareCli *cli.SquareCli) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) (err error) {
		appId := args[0]
		success, err := squareCli.Rest.ApplicationRestart(appId)
		if err != nil {
			return
		}

		if success {
			fmt.Fprintln(squareCli.Out(), "Your application has been successfuly restarted")
		} else {
			fmt.Fprintln(squareCli.Out(), "Failed to restart your application")
		}

		return nil
	}
}
