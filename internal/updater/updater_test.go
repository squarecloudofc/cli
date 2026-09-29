package updater

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"
)

func TestIsOutdated(t *testing.T) {
	cases := []struct {
		current, latest string
		want            bool
	}{
		{"v1.2.3", "v1.2.4", true},
		{"v1.2.3", "v1.10.0", true}, // numeric, not string, order
		{"1.9.9", "v2.0.0", true},
		{"v1.2.3", "v1.2.3", false},
		{"v1.3.0", "v1.2.9", false}, // newer than the latest release
		{"development", "v9.9.9", false},
		{"v1.2.4-preview", "v1.2.3", false},
		{"v1.2.3-5-gabc1234", "v1.2.4", false}, // git describe build
		{"v1.2.3", "", false},
		{"v1.2.3", "nightly", false},
	}
	for _, c := range cases {
		if got := IsOutdated(c.current, c.latest); got != c.want {
			t.Errorf("IsOutdated(%q, %q) = %v, want %v", c.current, c.latest, got, c.want)
		}
	}
}

func TestLatestCachesForADay(t *testing.T) {
	calls, status := 0, http.StatusOK
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(status)
		_, _ = w.Write([]byte(`{"tag_name":"v1.2.3"}`))
	}))
	defer srv.Close()

	cacheFile := filepath.Join(t.TempDir(), "sub", CacheFileName)
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	latest := func(at time.Duration) string {
		return Latest(context.Background(), srv.URL, cacheFile, start.Add(at))
	}

	if got := latest(0); got != "v1.2.3" || calls != 1 {
		t.Fatalf("first check = %q after %d calls, want v1.2.3 after 1", got, calls)
	}
	if got := latest(23 * time.Hour); got != "v1.2.3" || calls != 1 {
		t.Fatalf("fresh cache = %q after %d calls, want no new call", got, calls)
	}

	// Stale: one call; a failed one keeps the last known tag and is cached
	// too, so the next command does not wait on the network again.
	status = http.StatusForbidden // GitHub's rate limit answer
	if got := latest(25 * time.Hour); got != "v1.2.3" || calls != 2 {
		t.Fatalf("stale check = %q after %d calls, want v1.2.3 after 2", got, calls)
	}
	if got := latest(26 * time.Hour); got != "v1.2.3" || calls != 2 {
		t.Fatalf("after a failed check = %q after %d calls, want no new call", got, calls)
	}

	// A clock that went backwards does not trust the cache.
	status = http.StatusOK
	if latest(-time.Hour); calls != 3 {
		t.Fatalf("calls = %d, want a new call for a cache from the future", calls)
	}
}
