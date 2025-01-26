package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/squarecloudofc/cli/pkg/squarego/squarecloud"
)

type Client interface {
	HTTPClient() *http.Client
	Request(method, url string, rqBody []byte, rsBody any, options ...RequestOpt) error

	Close()
}

type clientImpl struct {
	config Config
	token  string
}

func NewClient(token string, opts ...ConfigOpt) Client {
	config := DefaultConfig()
	config.Apply(opts)

	return &clientImpl{
		config: *config,
		token:  token,
	}
}

func (c *clientImpl) HTTPClient() *http.Client {
	return c.config.HTTPClient
}

func (c *clientImpl) Close() {
	c.config.HTTPClient.CloseIdleConnections()
}

func (c *clientImpl) Request(method, url string, rqBody []byte, rsBody any, options ...RequestOpt) error {
	req, err := http.NewRequest(method, c.config.URL+url, bytes.NewReader(rqBody))
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", c.config.UserAgent)

	if c.token != "" {
		req.Header.Set("Authorization", c.token)
	}

	cfg := DefaultRequestConfig(req)
	cfg.Apply(options)

	response, err := c.HTTPClient().Do(cfg.Request)
	if err != nil {
		return err
	}

	var rawResponseBody []byte
	if response.Body != nil {
		if rawResponseBody, err = io.ReadAll(response.Body); err != nil {
			return fmt.Errorf("error reading response body: %w", err)
		}

		c.config.Logger.Debug("response", slog.String("code", response.Status), slog.String("body", string(rawResponseBody)))
	}

	switch response.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusNoContent:
		if rsBody != nil && response.Body != nil {
			if err := json.Unmarshal(rawResponseBody, rsBody); err != nil {
				return fmt.Errorf("error unmarshalling response body: %w", err)
			}
		}

		return nil
	default:
		var r squarecloud.APIResponse[any]
		if err := json.Unmarshal(rawResponseBody, &r); err != nil {
			return fmt.Errorf("error unmarshalling response body: %w", err)
		}

		return ParseError(&r)
	}
}
