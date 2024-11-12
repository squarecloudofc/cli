package square

type APIResponse[T any] struct {
	Response T      `json:"response"`
	Message  string `json:"message"`
	Status   string `json:"status"`
	Code     string `json:"code"`
}
