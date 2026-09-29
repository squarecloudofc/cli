package cli

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/squarecloudofc/cli/i18n"
	"github.com/squarecloudofc/cli/internal/config"
	"github.com/squarecloudofc/cli/internal/ui/selector"
	"github.com/squarecloudofc/sdk-api-go/v3"
)

type fakeCLI struct{ err bytes.Buffer }

func (f *fakeCLI) Config() *config.Config   { return &config.Config{} }
func (f *fakeCLI) API() *squarecloud.Client { return nil }
func (f *fakeCLI) I18n() i18n.Localizer     { return i18n.NewLocalizer("en") }
func (f *fakeCLI) Err() io.Writer           { return &f.err }
func (f *fakeCLI) In() io.ReadCloser        { return nil }
func (f *fakeCLI) Out() io.Writer           { return io.Discard }

func TestPrintError(t *testing.T) {
	en := i18n.NewLocalizer("en")
	cases := []struct {
		err  error
		want string
	}{
		{fmt.Errorf("wrapped: %w", &squarecloud.APIError{Status: 401, Code: squarecloud.CodeAccessDenied}), en.T("errors.api.unauthorized")},
		{&squarecloud.APIError{Status: 429, Code: squarecloud.CodeKeepCalm}, en.T("errors.api.rate_limited")},
		{&squarecloud.APIError{Status: 404, Code: "APP_NOT_FOUND"}, "APP_NOT_FOUND"},
		{errors.New("plain"), "plain"},
		{fmt.Errorf("save: %w", config.ErrNoConfigDir), en.T("errors.config.no_dir")},
		{selector.ErrCanceled, en.T("errors.common.selection_canceled")},
	}

	for _, c := range cases {
		f := &fakeCLI{}
		PrintError(f, c.err)
		if !strings.Contains(f.err.String(), c.want) {
			t.Errorf("PrintError(%v) = %q, want it to contain %q", c.err, f.err.String(), c.want)
		}
	}

	f := &fakeCLI{}
	PrintError(f, &AuthError{})
	if f.err.Len() != 0 {
		t.Errorf("AuthError printed %q", f.err.String())
	}
}
