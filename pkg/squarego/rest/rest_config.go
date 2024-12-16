package rest

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/squarecloudofc/cli/internal/build"
)

func DefaultConfig() *Config {
	return &Config{
		Logger:     slog.Default(),
		HTTPClient: &http.Client{Timeout: 25 * time.Second},
		URL:        ApiURL,
		UserAgent:  fmt.Sprintf("Square Cloud CLI (v%s)", build.Version),
	}
}

type Config struct {
	Logger     *slog.Logger
	HTTPClient *http.Client
	URL        string
	UserAgent  string
}

type ConfigOpt func(config *Config)

func (c *Config) Apply(opts []ConfigOpt) {
	for _, opt := range opts {
		opt(c)
	}
}

func WithLogger(logger *slog.Logger) ConfigOpt {
	return func(config *Config) {
		config.Logger = logger
	}
}

func WithHTTPClient(httpClient *http.Client) ConfigOpt {
	return func(config *Config) {
		config.HTTPClient = httpClient
	}
}

func WithURL(url string) ConfigOpt {
	return func(config *Config) {
		config.URL = url
	}
}

func WithUserAgent(userAgent string) ConfigOpt {
	return func(config *Config) {
		config.UserAgent = userAgent
	}
}
