package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/squarecloudofc/cli/internal/build"
	"github.com/squarecloudofc/cli/internal/config"
	"github.com/squarecloudofc/cli/pkg/squarego/rest"
)

type SquareCli struct {
	Config *config.Config
}

func NewSquareCli() *SquareCli {
	config, err := config.Load()
	if err != nil {
		panic("could not load config file")
	}

	squareCli := &SquareCli{
		Config: config,
	}

	return squareCli
}

func (squareCli *SquareCli) Rest() rest.Rest {
	restClient := rest.NewClient(squareCli.Config.AuthToken, rest.WithUserAgent(fmt.Sprintf("Square Cloud CLI (%s)", build.Version)))
	return rest.New(restClient)
}

func (squareCli *SquareCli) Err() io.Writer {
	return os.Stderr
}

func (squareCli *SquareCli) In() io.ReadCloser {
	return os.Stdin
}

func (squareCli *SquareCli) Out() io.Writer {
	return os.Stdout
}
