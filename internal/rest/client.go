package rest

type RestClient struct {
	token string
}

func NewClient(token string) *RestClient {
	return &RestClient{
		token: token,
	}
}
