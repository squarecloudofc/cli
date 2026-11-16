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
	"github.com/squarecloudofc/sdk-api-go/v4"
)

// APIKeyEnv, when set, overrides the saved key of every profile (CI).
const APIKeyEnv = "SQUARECLOUD_API_KEY"

// ProfileEnv picks the saved login to use when --profile is not given.
const ProfileEnv = "SQUARECLOUD_PROFILE"

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

// NewSquareCli builds the CLI without a profile: the root command applies
// --profile or SQUARECLOUD_PROFILE with UseProfile once flags are parsed, so
// the flag wins and a bad value of either only fails the commands that use
// it.
func NewSquareCli() SquareCLI {
	return &squarecliImpl{
		i18n: i18n.NewLocalizer(i18n.DetectSystemLanguage()),

		err: os.Stderr,
		in:  os.Stdin,
		out: os.Stdout,
	}
}

// UseProfile switches to the saved login of profile ("" is the default one).
func (squareCli *squarecliImpl) UseProfile(profile string) error {
	cfg, err := config.Load(strings.TrimSpace(profile))
	if errors.Is(err, config.ErrInvalidProfile) {
		return errors.New(squareCli.i18n.T("errors.config.invalid_profile", map[string]any{"Profile": profile}))
	}
	if err != nil {
		return errors.New(squareCli.i18n.T("errors.config.load", map[string]any{"Error": err.Error()}))
	}
	squareCli.config = cfg
	squareCli.api = nil // built on first use, with this profile's key
	return nil
}

// APIKey is the key the CLI authenticates with: SQUARECLOUD_API_KEY when set,
// otherwise the one `auth login` saved for the profile. The keychain is only
// read when the variable is unset.
func APIKey(cfg *config.Config) string {
	if key := strings.TrimSpace(os.Getenv(APIKeyEnv)); key != "" {
		return key
	}
	return cfg.Key()
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
	if squareCli.api == nil {
		squareCli.api = NewClient(APIKey(squareCli.config))
	}
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
