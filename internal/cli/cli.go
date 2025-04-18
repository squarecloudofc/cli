package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/squarecloudofc/cli/internal/build"
	"github.com/squarecloudofc/cli/internal/config"
	"github.com/squarecloudofc/cli/internal/i18n"
	"github.com/squarecloudofc/sdk-api-go/rest"
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

func NewSquareCli() SquareCLI {
	config, err := config.Load()
	if err != nil {
		panic("could not load config file")
	}

	restClient := rest.NewClient(
		config.AuthToken,
		rest.WithUserAgent(fmt.Sprintf("Square Cloud CLI (%s)", build.Version)),
	)

	i18n := i18n.NewLocalizer()

	squareCli := &squarecliImpl{
		config: config,
		rest:   rest.New(restClient),
		i18n:   i18n,

		err: os.Stderr,
		in:  os.Stdin,
		out: os.Stdout,
	}

	return squareCli
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
	return os.Stderr
}

func (squareCli *squarecliImpl) In() io.ReadCloser {
	return os.Stdin
}

func (squareCli *squarecliImpl) Out() io.Writer {
	return os.Stdout
}
