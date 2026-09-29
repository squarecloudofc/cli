package auth

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/cmdutil"
	"github.com/squarecloudofc/cli/internal/deviceauth"
	"github.com/squarecloudofc/cli/internal/ui"
	"github.com/squarecloudofc/sdk-api-go/v3"
	"golang.org/x/term"
)

func NewLoginCommand(squareCli cli.SquareCLI) *cobra.Command {
	var token string
	var withToken bool
	var noBrowser bool
	var yes bool

	cmd := &cobra.Command{
		Use:         "login",
		Short:       squareCli.I18n().T("metadata.commands.auth.login.short"),
		Annotations: map[string]string{"skipAuthCheck": "true"},
		Args:        cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if withToken {
				line, _ := bufio.NewReader(squareCli.In()).ReadString('\n')
				if token = strings.TrimSpace(line); token == "" {
					return errors.New(squareCli.I18n().T("commands.auth.login.empty_token"))
				}
			}

			if token != "" {
				return loginWithToken(cmd.Context(), squareCli, token, yes)
			}

			// The browser login's key expires in 90 days and only a person
			// can renew it: a pipeline would break silently. CI takes a key
			// made in the dashboard, which does not expire.
			if ci, _ := strconv.ParseBool(os.Getenv("CI")); ci {
				return errors.New(squareCli.I18n().T("commands.auth.login.ci", map[string]any{"Link": "https://squarecloud.app/account/security"}))
			}

			if f, ok := squareCli.In().(*os.File); !ok || !term.IsTerminal(int(f.Fd())) {
				return errors.New(squareCli.I18n().T("commands.auth.login.no_tty"))
			}

			return loginWithBrowser(cmd.Context(), squareCli, noBrowser, yes)
		},
	}

	cmd.Flags().StringVar(&token, "token", "", squareCli.I18n().T("metadata.flags.token"))
	cmd.Flags().BoolVar(&withToken, "with-token", false, squareCli.I18n().T("metadata.flags.with_token"))
	cmd.Flags().BoolVar(&noBrowser, "no-browser", false, squareCli.I18n().T("metadata.flags.no_browser"))
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, squareCli.I18n().T("metadata.flags.yes_switch_account"))
	cmd.MarkFlagsMutuallyExclusive("token", "with-token")
	return cmd
}

// loginWithToken validates a pasted key before saving it. Its expiry and
// scopes are unknown.
func loginWithToken(ctx context.Context, squareCli cli.SquareCLI, token string, yes bool) error {
	me, err := cli.NewClient(token).Account.Me(ctx)
	if err != nil {
		var apiErr *squarecloud.APIError
		if errors.As(err, &apiErr) && apiErr.Status == http.StatusUnauthorized {
			return errors.New(squareCli.I18n().T("commands.auth.login.error"))
		}
		return err
	}

	if err := confirmAccountSwitch(ctx, squareCli, me.User.ID, me.User.Email, yes); err != nil {
		return err
	}

	configuration := squareCli.Config()
	configuration.SetKey(token, me.User.ID, nil, nil)
	if err := configuration.Save(); err != nil {
		return err
	}

	fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.auth.login.success.0", map[string]any{"User": me.User.Name}))
	fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.auth.login.success.1"))
	printLoginNotes(squareCli, me.User.Plan.Name)
	return nil
}

// loginWithBrowser runs the device authorization: show a code, let the
// person approve it in the browser, poll until the API issues a key.
func loginWithBrowser(ctx context.Context, squareCli cli.SquareCLI, noBrowser, yes bool) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	client := &deviceauth.Client{
		BaseURL:    cli.BaseURL,
		HTTPClient: cli.HTTPClient(),
		UserAgent:  cli.UserAgent(),
	}

	grant, err := client.Start(ctx)
	if err != nil {
		return err
	}

	if !noBrowser {
		openBrowser(grant.VerificationURI)
	}

	model := newDeviceModel(squareCli, grant)
	program := tea.NewProgram(model, tea.WithContext(ctx), tea.WithOutput(squareCli.Out()))

	go func() {
		res, err := client.Wait(ctx, grant, func(code string) {
			program.Send(warnMsg(warningText(squareCli, code)))
		})
		program.Send(doneMsg{res, err})
	}()

	if _, err := program.Run(); err != nil && !errors.Is(err, tea.ErrProgramKilled) {
		return err
	}

	switch {
	case model.canceled || (model.result == nil && model.err == nil):
		return errors.New(squareCli.I18n().T("commands.auth.login.device.cancelled"))
	case model.err != nil:
		var apiErr *squarecloud.APIError
		if errors.As(model.err, &apiErr) && apiErr.Code == deviceauth.CodeInvalidGrant {
			return errors.New(squareCli.I18n().T("commands.auth.login.device.expired"))
		}
		return model.err
	}

	res := model.result
	if err := confirmAccountSwitch(ctx, squareCli, res.Account.ID, res.Account.Email, yes); err != nil {
		return err
	}

	configuration := squareCli.Config()
	configuration.SetKey(res.APIKey, res.Account.ID, res.ExpiresAt, res.Scopes)
	if err := configuration.Save(); err != nil {
		return err
	}

	fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.auth.login.device.success", map[string]any{"Email": res.Account.Email}))
	if res.ExpiresAt != nil {
		fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.auth.login.key_expires", map[string]any{"Date": res.ExpiresAt.Local().Format(time.DateOnly)}))
	}

	// Only for the plan warning: a failure here does not undo the login.
	var plan string
	if me, err := cli.NewClient(res.APIKey).Account.Me(ctx); err == nil {
		plan = me.User.Plan.Name
	}
	printLoginNotes(squareCli, plan)
	return nil
}

// confirmAccountSwitch asks before a login replaces the saved key of another
// account. Whoever approves a device code picks the account the key comes
// from; a silent switch would send this machine's deploys to that account.
func confirmAccountSwitch(ctx context.Context, squareCli cli.SquareCLI, accountID, email string, yes bool) error {
	saved := squareCli.Config()
	if saved.AuthToken == "" {
		return nil
	}

	savedID := saved.AccountID
	if savedID == "" {
		// Saved before the CLI recorded the account: ask the API. A key it
		// rejects (expired, revoked) has nothing left to protect.
		me, err := cli.NewClient(saved.AuthToken).Account.Me(ctx)
		if err != nil {
			return nil
		}
		savedID = me.User.ID
	}
	if savedID == accountID {
		return nil
	}

	return cmdutil.Confirm(squareCli, squareCli.I18n().T("commands.auth.login.switch_account", map[string]any{"Email": email}), yes)
}

func printLoginNotes(squareCli cli.SquareCLI, plan string) {
	if plan == "free" {
		fmt.Fprintln(squareCli.Out(), ui.TextDanger.Render(squareCli.I18n().T("commands.auth.login.warnings.no_plan", map[string]any{
			"Link": "https://squarecloud.app/pricing",
		})))
	}

	if os.Getenv(cli.APIKeyEnv) != "" {
		fmt.Fprintln(squareCli.Out(), ui.TextYellow.Render(squareCli.I18n().T("commands.auth.login.env_override")))
	}
}

func warningText(squareCli cli.SquareCLI, code string) string {
	if code == deviceauth.CodeKeyLimit {
		return squareCli.I18n().T("commands.auth.login.device.key_limit", map[string]any{"Link": "https://squarecloud.app/account/security"})
	}
	return squareCli.I18n().T("commands.auth.login.device.rate_limited")
}

// openBrowser opens an https link with the platform's handler, best effort:
// the URL is always on screen too.
func openBrowser(link string) {
	if u, err := url.Parse(link); err != nil || u.Scheme != "https" {
		return
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", link)
	case "darwin":
		cmd = exec.Command("open", link)
	default:
		cmd = exec.Command("xdg-open", link)
	}
	if cmd.Start() == nil {
		go cmd.Wait() //nolint:errcheck // reap the child
	}
}
