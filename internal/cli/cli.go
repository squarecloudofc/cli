package cli

import (
	"io"
	"os"

	"github.com/squarecloudofc/cli/internal/rest"
)

type SquareCli struct {
	Rest *rest.RestClient
}

func NewSquareCli() (squareCli *SquareCli) {
	squareCli = &SquareCli{}

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
