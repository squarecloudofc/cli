package app

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/ui/application_selector"
)

func NewLogsCommand(squareCli cli.SquareCLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs",
		Short: "See your application logs",
		RunE:  runLogsCommand(squareCli),
	}

	return cmd
}

func runLogsCommand(squareCli cli.SquareCLI) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) (err error) {
		var appId string
		rest := squareCli.Rest()

		if len(args) > 0 {
			appId = args[0]
		}

		if len(args) < 1 {
			m, err := application_selector.RunSelector(squareCli)
			if err != nil {
				return err
			}

			appId = m.ID
		}

		result, err := rest.GetApplicationLogs(appId)
		if err != nil {
			return err
		}

		fmt.Fprint(squareCli.Out(), result.Logs)

		return nil
	}
}
