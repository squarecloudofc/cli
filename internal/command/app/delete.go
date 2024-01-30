package app

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
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

		success, err := squareCli.Rest.ApplicationDelete(appId)
		if err != nil {
			return err
		}

		if success {
			fmt.Fprintln(squareCli.Out(), "Your application has been successfuly deleted")
		} else {
			fmt.Fprintln(squareCli.Out(), "Failed to delete your application")
		}

		return nil
	}
}
