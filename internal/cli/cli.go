package cli

import (
	"io"
	"os"

	"github.com/squarecloudofc/cli/internal/config"
	"github.com/squarecloudofc/cli/internal/rest"
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

func (squareCli *SquareCli) Rest() *rest.RestClient {
	return rest.NewClient(squareCli.Config.AuthToken)
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
