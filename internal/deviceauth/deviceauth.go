// Package deviceauth implements the Square Cloud device authorization login:
// the CLI asks for a short user code, the person approves it in the browser,
// and the CLI polls until the API hands over a fresh API key. It is the same
// grant the VS Code extension uses, with PKCE (S256) binding the claim to the
// process that started it.
package deviceauth

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	mrand "math/rand/v2"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/squarecloudofc/sdk-api-go/v3"
)

// ClientID identifies the CLI to the authorize routes.
const ClientID = "cli"

// Scopes are the permissions the CLI's commands need, nothing more.
var Scopes = []string{
	"account:read",
	"apps:read",
	"apps:control",
	"apps:write",
	"apps:deploy",
	"files:read",
	"files:write",
	"envs:read",
	"envs:write",
	"snapshots:read",
	"snapshots:write",
	"databases:read",
	"databases:write",
	"databases:credentials",
	"workspaces:manage",
}

// Codes of the authorize routes that Wait reacts to.
const (
	CodePending      = "AUTHORIZATION_PENDING"
	CodeInvalidGrant = "INVALID_GRANT"
	CodeKeyLimit     = "APIKEY_LIMIT_REACHED"
	CodeRateLimited  = squarecloud.CodeRateLimited
)

const (
	requestTimeout    = 30 * time.Second
	fallbackInterval  = 10 * time.Second
	fallbackExpiresIn = 600 * time.Second
	pollJitter        = 0.15
	// rateLimitBackoff: a 429 blocks for 300s while the grant lives 600s.
	rateLimitBackoff = 60 * time.Second
)

// Client talks to the authorize routes. Only BaseURL is required.
type Client struct {
	// BaseURL is the API root, e.g. squarecloud.DefaultBaseURL.
	BaseURL    string
	HTTPClient *http.Client
	UserAgent  string

	// Test hooks: a clock and a ctx-aware sleep (polls and backoffs).
	now   func() time.Time
	sleep func(context.Context, time.Duration) error
}

// Grant is a started authorization. Show UserCode and VerificationURI to the
// person, then call Wait.
type Grant struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"` // seconds
	Interval        int    `json:"interval"`   // seconds

	verifier string
	deadline time.Time
}

// ExpiresAt is when the grant stops being claimable.
func (g *Grant) ExpiresAt() time.Time { return g.deadline }

// Result is a claimed grant: a new API key and whom it belongs to.
type Result struct {
	APIKey  string `json:"api_key"`
	Account struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	} `json:"account"`
	Scopes    []string   `json:"scopes"`
	ExpiresAt *time.Time `json:"expires_at"`
}

// Start opens a grant (POST /account/authorize/device).
func (c *Client) Start(ctx context.Context) (*Grant, error) {
	verifier, challenge, err := newPKCE()
	if err != nil {
		return nil, err
	}

	var g Grant
	if err := c.post(ctx, "device", map[string]any{
		"client_id":        ClientID,
		"scopes":           Scopes,
		"challenge":        challenge,
		"challenge_method": "S256",
	}, &g); err != nil {
		return nil, err
	}

	expiresIn := time.Duration(g.ExpiresIn) * time.Second
	if expiresIn <= 0 {
		expiresIn = fallbackExpiresIn
	}
	g.verifier = verifier
	g.deadline = c.clock().Add(expiresIn)
	return &g, nil
}

// Wait polls the claim route until the person approves the grant, it
// expires, ctx is canceled or the API answers something unrecoverable. warn
// (optional) is called once per recoverable condition: CodeKeyLimit or
// CodeRateLimited. Expiry is reported as an *squarecloud.APIError with
// CodeInvalidGrant, like a grant the API itself rejects.
func (c *Client) Wait(ctx context.Context, g *Grant, warn func(code string)) (*Result, error) {
	interval := time.Duration(g.Interval) * time.Second
	if interval <= 0 {
		interval = fallbackInterval
	}
	warned := map[string]bool{}
	once := func(code string) {
		if !warned[code] && warn != nil {
			warn(code)
		}
		warned[code] = true
	}

	for c.clock().Before(g.deadline) {
		if err := c.pause(ctx, jitter(interval)); err != nil {
			return nil, err
		}

		var res Result
		err := c.post(ctx, "claim", map[string]string{
			"device_code": g.DeviceCode,
			"verifier":    g.verifier,
		}, &res)
		if err == nil {
			return &res, nil
		}
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		var apiErr *squarecloud.APIError
		if !errors.As(err, &apiErr) {
			return nil, err
		}
		switch {
		case apiErr.Code == CodePending:
			// The normal state; nothing was consumed.
		case apiErr.Status == 0, apiErr.Status >= http.StatusInternalServerError:
			// Network error or a 5xx from the API or its proxy: never abort.
			// If the key was created and the response lost, the next claim
			// returns the same key.
		case apiErr.Code == CodeKeyLimit:
			// The approved grant survives ~60s: the person can revoke a
			// key and the next poll picks it up.
			once(CodeKeyLimit)
		case apiErr.Status == http.StatusTooManyRequests,
			apiErr.Code == squarecloud.CodeRateLimited,
			apiErr.Code == squarecloud.CodeKeepCalm:
			once(CodeRateLimited)
			if err := c.pause(ctx, rateLimitBackoff); err != nil {
				return nil, err
			}
		default:
			// INVALID_GRANT, INVALID_VERIFIER, INVALID_REQUEST, ...: client
			// bugs or a dead grant. Retrying would burn the grant's attempts.
			return nil, err
		}
	}

	return nil, &squarecloud.APIError{Code: CodeInvalidGrant, Message: "authorization expired", Method: http.MethodPost, Path: "/account/authorize/claim"}
}

// post sends one JSON request to /account/authorize/<route> and decodes the
// {status, response} envelope into out. Every failure is an
// *squarecloud.APIError (Status 0 for network errors).
func (c *Client) post(ctx context.Context, route string, in, out any) error {
	body, err := json.Marshal(in)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	endpoint := strings.TrimRight(c.BaseURL, "/") + "/account/authorize/" + route
	path := endpoint
	if u, err := url.Parse(endpoint); err == nil {
		path = u.Path
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}

	hc := c.HTTPClient
	if hc == nil {
		hc = http.DefaultClient
	}
	res, err := hc.Do(req)
	if err != nil {
		return &squarecloud.APIError{Code: squarecloud.CodeNetworkError, Message: err.Error(), Method: http.MethodPost, Path: path}
	}
	defer res.Body.Close()

	data, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return &squarecloud.APIError{Code: squarecloud.CodeNetworkError, Message: err.Error(), Method: http.MethodPost, Path: path}
	}

	var envelope struct {
		Status   string          `json:"status"`
		Code     string          `json:"code"`
		Message  string          `json:"message"`
		Response json.RawMessage `json:"response"`
	}
	decodeErr := json.Unmarshal(data, &envelope)

	if res.StatusCode >= 300 || decodeErr != nil || envelope.Status == "error" {
		code := envelope.Code
		if code == "" {
			code = squarecloud.CodeUnknown
			if res.StatusCode == http.StatusTooManyRequests {
				code = squarecloud.CodeRateLimited // a 429 must read as one, body or not
			}
		}
		return &squarecloud.APIError{Status: res.StatusCode, Code: code, Message: envelope.Message, Method: http.MethodPost, Path: path}
	}

	if err := json.Unmarshal(envelope.Response, out); err != nil {
		return &squarecloud.APIError{Status: res.StatusCode, Code: squarecloud.CodeUnknown, Message: "invalid response: " + err.Error(), Method: http.MethodPost, Path: path}
	}
	return nil
}

// newPKCE returns a verifier (43 chars of the unreserved set) and its S256
// challenge (base64url of the SHA-256, no padding).
func newPKCE() (verifier, challenge string, err error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", "", err
	}
	verifier = base64.RawURLEncoding.EncodeToString(buf)
	return verifier, challengeOf(verifier), nil
}

func challengeOf(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

// jitter spreads polls up to 15% past the interval so clients behind one IP
// don't sync up (the poll budget is per IP). Never earlier: the API asks
// clients to wait at least the interval.
func jitter(d time.Duration) time.Duration {
	return time.Duration(float64(d) * (1 + mrand.Float64()*pollJitter))
}

func (c *Client) clock() time.Time {
	if c.now != nil {
		return c.now()
	}
	return time.Now()
}

func (c *Client) pause(ctx context.Context, d time.Duration) error {
	if c.sleep != nil {
		return c.sleep(ctx, d)
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
