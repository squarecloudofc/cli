package updater

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

// LatestReleaseURL is the GitHub API route of the newest release.
const LatestReleaseURL = "https://api.github.com/repos/squarecloudofc/cli/releases/latest"

// CacheFileName holds the last check, next to the CLI config.
const CacheFileName = "update-check.json"

// cacheTTL keeps the check to one GitHub call a day (the API allows 60
// unauthenticated calls an hour per IP).
const cacheTTL = 24 * time.Hour

type cache struct {
	Latest    string    `json:"latest"`
	CheckedAt time.Time `json:"checked_at"`
}

// Latest returns the newest release tag, or "" when unknown; the update
// banner is best-effort. It reads cacheFile while it is younger than a day,
// otherwise asks url and rewrites the cache, failures included, so an
// offline machine waits on the network once a day, not on every command.
func Latest(ctx context.Context, url, cacheFile string, now time.Time) string {
	var c cache
	if data, err := os.ReadFile(cacheFile); err == nil && json.Unmarshal(data, &c) == nil {
		if age := now.Sub(c.CheckedAt); age >= 0 && age < cacheTTL {
			return c.Latest
		}
	}

	if tag := fetch(ctx, url); tag != "" {
		c.Latest = tag
	}
	c.CheckedAt = now
	if data, err := json.Marshal(c); err == nil && os.MkdirAll(filepath.Dir(cacheFile), 0700) == nil {
		_ = os.WriteFile(cacheFile, data, 0600)
	}
	return c.Latest
}

func fetch(ctx context.Context, url string) string {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return ""
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return ""
	}

	return release.TagName
}

// IsOutdated reports whether current is an older release than latest. Both
// must be plain releases (vX.Y.Z): a development build, a prerelease or
// git-describe build, or anything unparsable is never reported as outdated.
func IsOutdated(current, latest string) bool {
	cur, ok1 := parseVersion(current)
	lat, ok2 := parseVersion(latest)
	return ok1 && ok2 && slices.Compare(cur, lat) < 0
}

func parseVersion(v string) ([]int, bool) {
	parts := strings.Split(strings.TrimPrefix(v, "v"), ".")
	if len(parts) != 3 {
		return nil, false
	}
	nums := make([]int, 3)
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return nil, false
		}
		nums[i] = n
	}
	return nums, true
}
