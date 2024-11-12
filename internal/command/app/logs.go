package app

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
)

func NewLogsCommand(squareCli *cli.SquareCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs",
		Short: "See your application logs",
		RunE:  runLogsCommand(squareCli),
	}

	return cmd
}

func runLogsCommand(squareCli *cli.SquareCli) func(cmd *cobra.Command, args []string) error {
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

		result, err := rest.GetApplicationLogs(appId)
		if err != nil {
			return err
		}

		fmt.Fprint(squareCli.Out(), result.Logs)

		return nil
	}
}
