package auth

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/ui"
)

func NewWhoamiCommand(squareCli cli.SquareCLI) *cobra.Command {
	return &cobra.Command{
		Use:   "whoami",
		Short: squareCli.I18n().T("metadata.commands.auth.whoami.short"),
		RunE: func(cmd *cobra.Command, args []string) error {
			self, err := squareCli.Rest().SelfUser()
			if err != nil || self.Name == "" {
				fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.whoami.none"))
				return err
			}

			username := ui.TextGreen.SetString(self.Name)

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
				return nil
			}

			if self.Plan.Duration != nil {
				daysRemaining := int(time.Until(time.Unix(*self.Plan.Duration, 0)).Hours() / 24)
				fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.auth.whoami.logged.remaining", map[string]any{
					"User": map[string]any{
						"PlanRemaining": daysRemaining,
					},
				}))
			}

			return nil
		},
	}
}
