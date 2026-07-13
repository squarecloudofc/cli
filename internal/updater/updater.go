package updater

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

const latestReleaseURL = "https://api.github.com/repos/squarecloudofc/cli/releases/latest"

// LatestVersion returns the tag name of the latest GitHub release, or "" on
// any failure — the update banner is best-effort.
func LatestVersion(ctx context.Context) string {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, latestReleaseURL, nil)
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
