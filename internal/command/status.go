package command

import (
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/i18n"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/cmdutil"
	"github.com/squarecloudofc/cli/internal/ui"
	"github.com/squarecloudofc/sdk-api-go/v4"
)

// NewStatusCommand reports the platform health (GET /service/status).
func NewStatusCommand(squareCli cli.SquareCLI) *cobra.Command {
	return &cobra.Command{
		Use:         "status",
		Short:       squareCli.I18n().T("metadata.commands.status.short"),
		Annotations: map[string]string{"skipAuthCheck": "true", "json": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			status, err := squareCli.API().Service.Status(cmd.Context())
			if err != nil {
				return err
			}

			if cmdutil.WantsJSON(cmd) {
				return cmdutil.PrintJSON(squareCli.Out(), status)
			}

			printStatus(squareCli.Out(), squareCli.I18n(), status)
			return nil
		},
	}
}

// printStatus renders three states: online, a problem (maintenance,
// degraded) and unknown, which means the check couldn't run, never that the
// platform is down. Open incidents and the components that aren't
// operational follow.
func printStatus(w io.Writer, t i18n.Localizer, status squarecloud.ServiceStatus) {
	switch status.Status {
	case "online":
		fmt.Fprintf(w, "%s %s\n", ui.CheckMark, status.Message)
	case "unknown":
		fmt.Fprintf(w, "%s %s\n", ui.UnknownMark, t.T("commands.status.unknown"))
	default:
		fmt.Fprintf(w, "%s %s\n", ui.WarnMark, status.Message)
	}

	if len(status.Incidents) > 0 {
		fmt.Fprintf(w, "\n%s\n", t.T("commands.status.incidents"))
		for _, incident := range status.Incidents {
			fmt.Fprintf(w, "  %s %s (%s)\n", ui.WarnMark, incident.Title, incident.Status)
			when := t.T("commands.status.started", map[string]any{"Date": incident.StartedAt.Local().Format(time.DateTime)})
			if incident.EndsAt != nil {
				when += " · " + t.T("commands.status.ends", map[string]any{"Date": incident.EndsAt.Local().Format(time.DateTime)})
			}
			fmt.Fprintf(w, "    %s\n", when)
		}
	}

	if len(status.Components) > 0 {
		fmt.Fprintf(w, "\n%s\n", t.T("commands.status.components"))
		for _, component := range status.Components {
			fmt.Fprintf(w, "  %s %s (%s)\n", ui.WarnMark, component.Name, component.Status)
		}
	}
}
