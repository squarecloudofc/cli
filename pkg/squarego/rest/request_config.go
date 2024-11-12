package rest

import (
	"context"
	"fmt"
	"net/http"
)

type RequestConfig struct {
	Request *http.Request
	Ctx     context.Context
}

func DefaultRequestConfig(req *http.Request) *RequestConfig {
	return &RequestConfig{
		Request: req,
		Ctx:     context.TODO(),
	}
}

type RequestOpt func(config *RequestConfig)

func (c *RequestConfig) Apply(opts []RequestOpt) {
	for _, opt := range opts {
		opt(c)
	}
	if c.Ctx == nil {
		c.Ctx = context.TODO()
	}
}

func WithHeader(key string, value string) RequestOpt {
	return func(config *RequestConfig) {
		config.Request.Header.Set(key, value)
	}
}

func WithToken(token string) RequestOpt {
	return WithHeader("Authorization", token)
}

func WithQueryParam(param string, value any) RequestOpt {
	return func(config *RequestConfig) {
		values := config.Request.URL.Query()
		values.Add(param, fmt.Sprint(value))
		config.Request.URL.RawQuery = values.Encode()
	}
}
