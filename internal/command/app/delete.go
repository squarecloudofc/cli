package app

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/ui"
)

func NewDeleteCommand(squareCli *cli.SquareCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete your application",
		RunE:  runDeleteCommand(squareCli),
	}

	return cmd
}

func runDeleteCommand(squareCli *cli.SquareCli) func(cmd *cobra.Command, args []string) error {
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

		success, err := rest.ApplicationDelete(appId)
		if err != nil {
			return err
		}

		if success {
			fmt.Fprintf(squareCli.Out(), "%s Your application has been successfuly deleted\n", ui.CheckMark)
		} else {
			fmt.Fprintf(squareCli.Out(), "%s Failed to delete your application\n", ui.XMark)
		}

		return nil
	}
}
