package updater

import (
	"context"
	"os"

	"github.com/google/go-github/v58/github"
	"github.com/squarecloudofc/cli/internal/build"
)

func GetLatestRelease(ctx context.Context) (*github.RepositoryRelease, error) {
	client := github.NewClient(nil)
	release, _, err := client.Repositories.GetLatestRelease(ctx, "squarecloudofc", "cli")

	if release != nil && build.Version != *release.TagName {
		return release, nil
	}

	return release, err
}

func IsCI() bool {
	return os.Getenv("CI") != "" || // GitHub Actions, Travis CI, CircleCI, Cirrus CI, GitLab CI, AppVeyor, CodeShip, dsari
		os.Getenv("BUILD_NUMBER") != "" // Jenkins, TeamCity
}
