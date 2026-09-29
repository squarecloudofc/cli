package cli

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/squarecloudofc/cli/i18n"
	"github.com/squarecloudofc/cli/internal/build"
	"github.com/squarecloudofc/cli/internal/config"
	"github.com/squarecloudofc/sdk-api-go/v3"
)

// APIKeyEnv, when set, overrides the key stored in the config file (CI).
const APIKeyEnv = "SQUARECLOUD_API_KEY"

var _ SquareCLI = (*squarecliImpl)(nil)

type SquareCLI interface {
	Config() *config.Config
	API() *squarecloud.Client
	I18n() i18n.Localizer

	Err() io.Writer
	In() io.ReadCloser
	Out() io.Writer
}

type squarecliImpl struct {
	config *config.Config
	api    *squarecloud.Client
	i18n   i18n.Localizer

	err io.Writer
	in  io.ReadCloser
	out io.Writer
}

func NewSquareCli() (SquareCLI, error) {
	localizer := i18n.NewLocalizer(i18n.DetectSystemLanguage())
	config, err := config.Load()
	if err != nil {
		return nil, errors.New(localizer.T("errors.config.load", map[string]any{"Error": err.Error()}))
	}

	return &squarecliImpl{
		config: config,
		api:    NewClient(APIKey(config)),
		i18n:   localizer,

		err: os.Stderr,
		in:  os.Stdin,
		out: os.Stdout,
	}, nil
}

// APIKey is the key the CLI authenticates with: SQUARECLOUD_API_KEY when set,
// otherwise the one saved by `auth login`.
func APIKey(cfg *config.Config) string {
	if key := strings.TrimSpace(os.Getenv(APIKeyEnv)); key != "" {
		return key
	}
	return cfg.AuthToken
}

// UserAgent is sent on every request the CLI makes.
func UserAgent() string {
	return fmt.Sprintf("Square Cloud CLI (%s)", build.Version)
}

// HTTPClient is the client shared by the SDK and the login flow. With
// SQUARECLOUD_DEBUG set it traces each request to stderr.
func HTTPClient() *http.Client {
	if os.Getenv("SQUARECLOUD_DEBUG") == "" {
		return &http.Client{}
	}
	return &http.Client{Transport: debugTransport{http.DefaultTransport}}
}

// BaseURL is the API root of every client; tests point it at a fake API.
var BaseURL = squarecloud.DefaultBaseURL

// NewClient builds an SDK client for key with the CLI's User-Agent.
func NewClient(key string) *squarecloud.Client {
	return squarecloud.New(key,
		squarecloud.WithBaseURL(BaseURL),
		squarecloud.WithUserAgent(UserAgent()),
		squarecloud.WithHTTPClient(HTTPClient()),
	)
}

// debugTransport logs method, path, status and duration. Never headers,
// query strings or bodies: they carry the API key and signed URLs.
type debugTransport struct{ next http.RoundTripper }

func (t debugTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()
	res, err := t.next.RoundTrip(req)
	status := "ERR"
	if err == nil {
		status = res.Status
	}
	fmt.Fprintf(os.Stderr, "[debug] %s %s -> %s (%s)\n", req.Method, req.URL.Path, status, time.Since(start).Round(time.Millisecond))
	return res, err
}

func (squareCli *squarecliImpl) Config() *config.Config {
	return squareCli.config
}

func (squareCli *squarecliImpl) API() *squarecloud.Client {
	return squareCli.api
}

func (squareCli *squarecliImpl) I18n() i18n.Localizer {
	return squareCli.i18n
}

func (squareCli *squarecliImpl) Err() io.Writer {
	return squareCli.err
}

func (squareCli *squarecliImpl) In() io.ReadCloser {
	return squareCli.in
}

func (squareCli *squarecliImpl) Out() io.Writer {
	return squareCli.out
}
