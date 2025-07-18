package auth

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/ui"
)

func NewWhoamiCommand(squareCli cli.SquareCLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "whoami",
		Short: "Print the user informations associated with current Square Cloud Token",
		RunE:  runWhoamiCommand(squareCli),
	}

	return cmd
}

func runWhoamiCommand(squareCli cli.SquareCLI) RunEFunc {
	return func(cmd *cobra.Command, args []string) (err error) {
		rest := squareCli.Rest()
		self, err := rest.SelfUser()
		if err != nil || self.Name == "" {
			fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.whoami.none"))
			return err
		}

		username := ui.TextGreen.SetString(self.Name)

		diff := time.Until(time.Unix(self.Plan.Duration, 0))
		daysRemaining := int(diff.Hours() / 24)

		fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.auth.whoami.logged.plan", map[string]any{
			"User": map[string]any{
				"Name": username.String(),
				"Plan": self.Plan.Name,
			},
		}))

		if self.Plan.Name == "free" {
			fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.auth.whoami.logged.expired", map[string]any{
				"Link": ui.TextBlue.Render("https://squarecloud.app/pricing"),
			}))
		} else {
			fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.auth.whoami.logged.remaining", map[string]any{
				"User": map[string]any{
					"PlanRemaining": daysRemaining,
				},
			}))
		}

		return
	}
}
