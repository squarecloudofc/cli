package app

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/ui"
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

		success, err := squareCli.Rest.ApplicationRestart(appId)
		if err != nil {
			return err
		}

		if success {
			fmt.Fprintf(squareCli.Out(), "%s Your application has been successfuly restarted\n", ui.CheckMark)
		} else {
			fmt.Fprintf(squareCli.Out(), "%s Failed to restart your application\n", ui.XMark)
		}

		return nil
	}
}
