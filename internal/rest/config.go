package rest

type RequestOption func(c *RequestConfig)

func WithHeader(key, value string) RequestOption {
	return func(c *RequestConfig) {
		c.Request.Header.Set(key, value)
	}
}
