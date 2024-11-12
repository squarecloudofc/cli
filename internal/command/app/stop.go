package app

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/ui"
	"github.com/squarecloudofc/cli/pkg/squarego/square"
)

func NewStopCommand(squareCli *cli.SquareCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop your application",
		RunE:  runStopCommand(squareCli),
	}

	return cmd
}

func runStopCommand(squareCli *cli.SquareCli) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) (err error) {
		var appId string
		rest := squareCli.Rest()

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

		err = rest.PostApplicationSignal(appId, square.ApplicationSignalStop)
		if err != nil {
			fmt.Fprintf(squareCli.Out(), "%s Failed to stop your application", ui.XMark)
			return
		}

		fmt.Fprintf(squareCli.Out(), "%s Your application has been successfuly stopped", ui.CheckMark)
		return nil
	}
}
