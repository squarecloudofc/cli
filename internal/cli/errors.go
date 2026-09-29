package cli

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/squarecloudofc/cli/internal/config"
	"github.com/squarecloudofc/cli/internal/ui"
	"github.com/squarecloudofc/cli/internal/ui/selector"
	"github.com/squarecloudofc/sdk-api-go/v3"
)

type AuthError struct {
	err error
}

func (ae *AuthError) Error() string {
	if ae.err == nil {
		return "not logged in"
	}
	return ae.err.Error()
}

// PrintError reports a command's error. API errors with a known remedy (an
// expired or revoked key, a rate limit) get a localized hint instead of the
// raw error text.
func PrintError(squareCli SquareCLI, err error) {
	var authErr *AuthError
	if errors.As(err, &authErr) {
		return // already explained by the auth check
	}

	if hint := Hint(squareCli, err); hint != "" {
		fmt.Fprintf(squareCli.Err(), "%s %s\n", ui.XMark, hint)
		return
	}

	fmt.Fprintln(squareCli.Err(), err)
}

// Hint is the localized explanation of an error the person can act on (sign
// in again, slow down, set HOME), or "" for any other error.
func Hint(squareCli SquareCLI, err error) string {
	if errors.Is(err, config.ErrNoConfigDir) {
		return squareCli.I18n().T("errors.config.no_dir")
	}
	if errors.Is(err, selector.ErrCanceled) {
		return squareCli.I18n().T("errors.common.selection_canceled")
	}

	var apiErr *squarecloud.APIError
	if !errors.As(err, &apiErr) {
		return ""
	}

	switch {
	case apiErr.Status == http.StatusUnauthorized || apiErr.Code == squarecloud.CodeAccessDenied:
		return squareCli.I18n().T("errors.api.unauthorized")
	case apiErr.Status == http.StatusTooManyRequests || apiErr.Code == squarecloud.CodeRateLimited || apiErr.Code == squarecloud.CodeKeepCalm:
		return squareCli.I18n().T("errors.api.rate_limited")
	}
	return ""
}
