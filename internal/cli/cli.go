package cli

import (
	"io"
	"os"

	"github.com/squarecloudofc/cli/internal/config"
	"github.com/squarecloudofc/cli/internal/rest"
)

type SquareCli struct {
	Rest   *rest.RestClient
	Config *config.Config
}

func NewSquareCli() (squareCli *SquareCli) {
	config, err := config.Load()
	if err != nil {
		panic("could not load config file")
	}

	squareCli = &SquareCli{
		Config: config,
	}

	return
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
