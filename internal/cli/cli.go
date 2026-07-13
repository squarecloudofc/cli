package cli

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/squarecloudofc/cli/i18n"
	"github.com/squarecloudofc/cli/internal/build"
	"github.com/squarecloudofc/cli/internal/config"
	"github.com/squarecloudofc/sdk-api-go/v2/rest"
)

var _ SquareCLI = (*squarecliImpl)(nil)

type SquareCLI interface {
	Config() *config.Config
	Rest() rest.Rest
	I18n() i18n.Localizer

	Err() io.Writer
	In() io.ReadCloser
	Out() io.Writer
}

type squarecliImpl struct {
	config *config.Config
	rest   rest.Rest
	i18n   i18n.Localizer

	err io.Writer
	in  io.ReadCloser
	out io.Writer
}

func NewSquareCli() (SquareCLI, error) {
	config, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("could not load config file: %w", err)
	}

	restOpts := []rest.ConfigOpt{
		rest.WithUserAgent(fmt.Sprintf("Square Cloud CLI (%s)", build.Version)),
	}

	// Set SQUARECLOUD_DEBUG=1 to log every request/response to stderr.
	if os.Getenv("SQUARECLOUD_DEBUG") != "" {
		logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
		restOpts = append(restOpts, rest.WithLogger(logger))
	}

	restClient := rest.NewClient(config.AuthToken, restOpts...)

	return &squarecliImpl{
		config: config,
		rest:   rest.New(restClient),
		i18n:   i18n.NewLocalizer(config.Locale),

		err: os.Stderr,
		in:  os.Stdin,
		out: os.Stdout,
	}, nil
}

func (squareCli *squarecliImpl) Config() *config.Config {
	return squareCli.config
}

func (squareCli *squarecliImpl) Rest() rest.Rest {
	return squareCli.rest
}

func (squareCli *squarecliImpl) I18n() i18n.Localizer {
	return squareCli.i18n
}

func (squareCli *squarecliImpl) Err() io.Writer {
	return squareCli.err
}

func (squareCli *squarecliImpl) In() io.ReadCloser {
	return squareCli.in
}

func (squareCli *squarecliImpl) Out() io.Writer {
	return squareCli.out
}
