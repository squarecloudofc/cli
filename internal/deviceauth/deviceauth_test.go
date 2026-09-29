package deviceauth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/squarecloudofc/sdk-api-go/v3"
)

// fakeAPI serves /account/authorize/{device,claim}. claims answers the Nth
// claim (0-based); past its end the last entry repeats.
type fakeAPI struct {
	t         *testing.T
	expiresIn int
	claims    []func(w http.ResponseWriter)

	mu        sync.Mutex
	challenge string
	verifiers []string
	calls     int
}

func (f *fakeAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	_ = json.NewDecoder(r.Body).Decode(&body)

	f.mu.Lock()
	defer f.mu.Unlock()

	switch r.URL.Path {
	case "/v2/account/authorize/device":
		if body["client_id"] != ClientID || body["challenge_method"] != "S256" {
			f.t.Errorf("bad device body: %v", body)
		}
		f.challenge, _ = body["challenge"].(string)
		writeJSON(w, 200, map[string]any{"status": "success", "response": map[string]any{
			"device_code":      "0123456789abcdef0123456789abcdef",
			"user_code":        "ABCD2345",
			"verification_uri": "https://squarecloud.app/account/security/authorize",
			"expires_in":       f.expiresIn,
			"interval":         10,
		}})
	case "/v2/account/authorize/claim":
		if body["device_code"] != "0123456789abcdef0123456789abcdef" {
			f.t.Errorf("bad claim body: %v", body)
		}
		v, _ := body["verifier"].(string)
		f.verifiers = append(f.verifiers, v)
		i := min(f.calls, len(f.claims)-1)
		f.calls++
		f.claims[i](w)
	default:
		w.WriteHeader(404)
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func apiError(status int, code string) func(http.ResponseWriter) {
	return func(w http.ResponseWriter) {
		writeJSON(w, status, map[string]any{"status": "error", "code": code})
	}
}

func success(w http.ResponseWriter) {
	writeJSON(w, 200, map[string]any{"status": "success", "response": map[string]any{
		"api_key":    "test-key",
		"account":    map[string]any{"id": "u1", "email": "dev@example.com"},
		"scopes":     Scopes,
		"expires_at": "2026-12-27T00:00:00.000Z",
	}})
}

var pending = apiError(400, CodePending)

// run starts a grant against f and waits on it with a fake clock. It returns
// the result, the error, every requested sleep and every warning.
func run(t *testing.T, f *fakeAPI) (*Result, error, []time.Duration, []string) {
	t.Helper()
	f.t = t
	if f.expiresIn == 0 {
		f.expiresIn = 600
	}
	srv := httptest.NewServer(f)
	defer srv.Close()

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	var sleeps []time.Duration
	c := &Client{
		BaseURL: srv.URL + "/v2",
		now:     func() time.Time { return now },
		sleep: func(ctx context.Context, d time.Duration) error {
			sleeps = append(sleeps, d)
			now = now.Add(d)
			return ctx.Err()
		},
	}

	g, err := c.Start(context.Background())
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if g.UserCode != "ABCD2345" {
		t.Fatalf("user code = %q", g.UserCode)
	}

	var warnings []string
	res, err := c.Wait(context.Background(), g, func(code string) { warnings = append(warnings, code) })
	return res, err, sleeps, warnings
}

func codeOf(err error) string {
	var apiErr *squarecloud.APIError
	if errors.As(err, &apiErr) {
		return apiErr.Code
	}
	return ""
}

func TestSuccessAfterPending(t *testing.T) {
	f := &fakeAPI{claims: []func(http.ResponseWriter){pending, pending, pending, success}}
	res, err, sleeps, warnings := run(t, f)
	if err != nil {
		t.Fatalf("Wait: %v", err)
	}
	if res.APIKey != "test-key" || res.Account.Email != "dev@example.com" || res.ExpiresAt == nil {
		t.Fatalf("result = %+v", res)
	}
	if f.calls != 4 || len(warnings) != 0 {
		t.Fatalf("calls = %d, warnings = %v", f.calls, warnings)
	}
	for _, d := range sleeps {
		if d < 10*time.Second || d > 11500*time.Millisecond {
			t.Fatalf("poll interval %v outside 10s +15%%", d)
		}
	}
}

func TestPKCE(t *testing.T) {
	// RFC 7636, appendix B.
	if got := challengeOf("dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"); got != "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM" {
		t.Fatalf("challenge = %q", got)
	}

	f := &fakeAPI{claims: []func(http.ResponseWriter){success}}
	if _, err, _, _ := run(t, f); err != nil {
		t.Fatal(err)
	}
	v := f.verifiers[0]
	if !regexp.MustCompile(`^[A-Za-z0-9._~-]{43,128}$`).MatchString(v) {
		t.Fatalf("verifier %q is not a valid PKCE verifier", v)
	}
	if len(f.challenge) != 43 || f.challenge != challengeOf(v) {
		t.Fatalf("challenge %q does not match verifier", f.challenge)
	}
}

func TestKeyLimitWarnsOnceAndKeepsPolling(t *testing.T) {
	limit := apiError(403, CodeKeyLimit)
	f := &fakeAPI{claims: []func(http.ResponseWriter){limit, limit, limit, success}}
	res, err, _, warnings := run(t, f)
	if err != nil || res == nil {
		t.Fatalf("Wait: %v", err)
	}
	if len(warnings) != 1 || warnings[0] != CodeKeyLimit {
		t.Fatalf("warnings = %v", warnings)
	}
}

func TestRateLimitBacksOff(t *testing.T) {
	bare429 := func(w http.ResponseWriter) { w.WriteHeader(429) } // no body
	f := &fakeAPI{claims: []func(http.ResponseWriter){bare429, apiError(429, "KEEP_CALM"), success}}
	res, err, sleeps, warnings := run(t, f)
	if err != nil || res == nil {
		t.Fatalf("Wait: %v", err)
	}
	if len(warnings) != 1 || warnings[0] != CodeRateLimited {
		t.Fatalf("warnings = %v", warnings)
	}
	backoffs := 0
	for _, d := range sleeps {
		if d == rateLimitBackoff {
			backoffs++
		}
	}
	if backoffs != 2 {
		t.Fatalf("backoffs = %d, sleeps = %v", backoffs, sleeps)
	}
}

func TestNetworkErrorKeepsPolling(t *testing.T) {
	drop := func(w http.ResponseWriter) {
		conn, _, _ := w.(http.Hijacker).Hijack()
		conn.Close()
	}
	f := &fakeAPI{claims: []func(http.ResponseWriter){drop, success}}
	if res, err, _, _ := run(t, f); err != nil || res == nil {
		t.Fatalf("Wait: %v", err)
	}
	if f.calls != 2 {
		t.Fatalf("calls = %d", f.calls)
	}
}

func TestServerErrorKeepsPolling(t *testing.T) {
	badGateway := func(w http.ResponseWriter) { w.WriteHeader(502) } // proxy page, no JSON
	f := &fakeAPI{claims: []func(http.ResponseWriter){badGateway, apiError(503, "UNKNOWN"), success}}
	if res, err, _, _ := run(t, f); err != nil || res == nil {
		t.Fatalf("Wait: %v", err)
	}
	if f.calls != 3 {
		t.Fatalf("calls = %d", f.calls)
	}
}

func TestInvalidGrantStops(t *testing.T) {
	f := &fakeAPI{claims: []func(http.ResponseWriter){apiError(400, CodeInvalidGrant), success}}
	_, err, _, _ := run(t, f)
	if codeOf(err) != CodeInvalidGrant || f.calls != 1 {
		t.Fatalf("err = %v, calls = %d", err, f.calls)
	}
}

func TestFatalCodeStops(t *testing.T) {
	f := &fakeAPI{claims: []func(http.ResponseWriter){pending, apiError(400, "INVALID_VERIFIER"), success}}
	_, err, _, _ := run(t, f)
	if codeOf(err) != "INVALID_VERIFIER" || f.calls != 2 {
		t.Fatalf("err = %v, calls = %d", err, f.calls)
	}
}

func TestExpiry(t *testing.T) {
	f := &fakeAPI{expiresIn: 35, claims: []func(http.ResponseWriter){pending}}
	_, err, _, _ := run(t, f)
	var apiErr *squarecloud.APIError
	if !errors.As(err, &apiErr) || apiErr.Code != CodeInvalidGrant || apiErr.Status != 0 {
		t.Fatalf("err = %v", err)
	}
	if f.calls < 3 || f.calls > 4 {
		t.Fatalf("calls = %d, want 3-4 for 35s at 10s +15%%", f.calls)
	}
}

func TestCancel(t *testing.T) {
	f := &fakeAPI{t: t, expiresIn: 600, claims: []func(http.ResponseWriter){pending}}
	srv := httptest.NewServer(f)
	defer srv.Close()

	c := &Client{BaseURL: srv.URL + "/v2"} // real sleep: 10s polls
	g, err := c.Start(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(50*time.Millisecond, cancel)

	start := time.Now()
	_, err = c.Wait(ctx, g, nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
	if time.Since(start) > 2*time.Second {
		t.Fatalf("cancel took %v", time.Since(start))
	}
}
