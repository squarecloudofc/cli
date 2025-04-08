package app

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/ui"
	"github.com/squarecloudofc/squarego/squarecloud"
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

		err = rest.PostApplicationSignal(appId, squarecloud.ApplicationSignalStart)
		if err != nil {
			fmt.Fprintf(squareCli.Out(), "%s Failed to start your application", ui.XMark)
			return
		}

		fmt.Fprintf(squareCli.Out(), "%s Your application has been successfuly started", ui.CheckMark)
		return nil
	}
}
