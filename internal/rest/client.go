package rest

type RestClient struct {
	token string
}

func NewClient(token string) *RestClient {
	rest := &RestClient{
		token: token,
	}

	return rest
}

func (c *RestClient) Token() string {
	return c.token
}
