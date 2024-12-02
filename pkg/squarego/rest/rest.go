package rest

import (
	"net/http"

	"github.com/squarecloudofc/cli/pkg/squarego/squarecloud"
)

var _ Rest = (*restImpl)(nil)

type Rest interface {
	SelfUser(opts ...RequestOpt) (squarecloud.User, error)

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

type responseUser struct {
	Applications []squarecloud.UserApplication `json:"applications"`
	User         squarecloud.User              `json:"user"`
}

func (s *restImpl) SelfUser(opts ...RequestOpt) (squarecloud.User, error) {
	var r squarecloud.APIResponse[responseUser]
	err := s.Request(http.MethodGet, EndpointUser(), nil, &r, opts...)

	return r.Response.User, err
}
