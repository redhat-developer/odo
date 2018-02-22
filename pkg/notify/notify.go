package notify

import (
	"context"
	"github.com/blang/semver"
	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	"strings"
)

const (
	// project GitHub organization name
	ghorg = "redhat-developer"
	// project GitHub repository name
	ghrepo = "ocdev"
	// URL of the installation shell script
	InstallScriptURL = "https://raw.githubusercontent.com/redhat-developer/ocdev/master/scripts/install.sh"
)

// getLatestReleaseTag polls ocdev's upstream GitHub repository to get the
// tag of the latest release
func getLatestReleaseTag() (string, error) {
	client := github.NewClient(nil)
	release, response, err := client.Repositories.GetLatestRelease(context.Background(), ghorg, ghrepo)
	defer func() {
		if err = response.Body.Close(); err != nil {
			err = errors.Wrap(err, "closing response body failed")
		}
	}()

	return *release.TagName, nil
}

// CheckLatestReleaseTag returns the latest release tag if a newer latest
// release is available, else returns an empty string
func CheckLatestReleaseTag(currentVersion string) (string, error) {
	currentSemver, err := semver.Make(strings.TrimPrefix(currentVersion, "v"))
	if err != nil {
		return "", errors.Wrapf(err, "unable to make semver from the current version: %v", currentVersion)
	}

	latestTag, err := getLatestReleaseTag()
	if err != nil {
		return "", errors.Wrap(err, "unable to get latest release tag")
	}

	latestSemver, err := semver.Make(strings.TrimPrefix(latestTag, "v"))
	if err != nil {
		return "", errors.Wrapf(err, "unable to make semver from the latest release tag: %v", latestTag)
	}

	if currentSemver.LT(latestSemver) {
		return latestTag, nil
	}

	return "", nil
}
