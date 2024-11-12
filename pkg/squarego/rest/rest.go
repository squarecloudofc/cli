package rest

import (
	"net/http"

	"github.com/squarecloudofc/cli/pkg/squarego/square"
)

var _ Rest = (*restImpl)(nil)

type Rest interface {
	SelfUser(opts ...RequestOpt) (square.User, error)

	Applications
}

type restImpl struct {
	Client
	Applications
}

func New(client Client) Rest {
	return &restImpl{
		Client:       client,
		Applications: NewApplications(client),
	}
}

func (s *restImpl) SelfUser(opts ...RequestOpt) (square.User, error) {
	var r square.APIResponse[square.User]
	err := s.Request(http.MethodGet, EndpointUser(), nil, &r)

	return r.Response, err
}
