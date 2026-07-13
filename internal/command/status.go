package command

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/cmdutil"
	"github.com/squarecloudofc/cli/internal/ui"
)

// NewStatusCommand reports the platform health (GET /service/status).
func NewStatusCommand(squareCli cli.SquareCLI) *cobra.Command {
	return &cobra.Command{
		Use:         "status",
		Short:       squareCli.I18n().T("metadata.commands.status.short"),
		Annotations: map[string]string{"skipAuthCheck": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			status, err := squareCli.Rest().ServiceStatus()
			if err != nil {
				return err
			}

			if cmdutil.WantsJSON(cmd) {
				return cmdutil.PrintJSON(squareCli.Out(), status)
			}

			mark := ui.CheckMark
			if status.Status != "online" {
				mark = ui.XMark
			}

			fmt.Fprintf(squareCli.Out(), "%s %s\n", mark, status.Message)
			return nil
		},
	}
}
