package auth

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/cmdutil"
	"github.com/squarecloudofc/cli/internal/ui"
)

func NewWhoamiCommand(squareCli cli.SquareCLI) *cobra.Command {
	return &cobra.Command{
		Use:         "whoami",
		Short:       squareCli.I18n().T("metadata.commands.auth.whoami.short"),
		Annotations: map[string]string{"json": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			me, err := squareCli.API().Account.Me(cmd.Context())
			if err != nil || me.User.Name == "" {
				fmt.Fprintln(squareCli.Err(), squareCli.I18n().T("commands.whoami.none"))
				return err
			}
			self := me.User

			if cmdutil.WantsJSON(cmd) {
				return cmdutil.PrintJSON(squareCli.Out(), self)
			}

			fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.auth.whoami.logged.plan", map[string]any{
				"User": map[string]any{
					"Name": ui.TextGreen.Render(self.Name),
					"Plan": self.Plan.Name,
				},
			}))

			if self.Plan.Name == "free" {
				fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.auth.whoami.logged.expired", map[string]any{
					"Link": ui.TextBlue.Render("https://squarecloud.app/pricing"),
				}))
			} else if self.Plan.Duration != nil {
				fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.auth.whoami.logged.remaining", map[string]any{
					"User": map[string]any{
						"PlanRemaining": daysUntil(*self.Plan.Duration, time.Now()),
					},
				}))
			}

			// Expiry and scopes describe the saved key, not an env override.
			if os.Getenv(cli.APIKeyEnv) != "" {
				fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.auth.whoami.key.env"))
				return nil
			}

			config := squareCli.Config()
			if config.KeyExpiresAt != nil {
				fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.auth.whoami.key.expires", map[string]any{
					"Date": config.KeyExpiresAt.Local().Format(time.DateOnly),
				}))
			}
			if len(config.KeyScopes) > 0 {
				fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.auth.whoami.key.scopes", map[string]any{
					"Scopes": strings.Join(config.KeyScopes, ", "),
				}))
			}

			return nil
		},
	}
}

// daysUntil converts the plan expiry (Unix milliseconds) into whole days
// left from now.
func daysUntil(expiryMs int64, now time.Time) int {
	return int(time.UnixMilli(expiryMs).Sub(now).Hours() / 24)
}
