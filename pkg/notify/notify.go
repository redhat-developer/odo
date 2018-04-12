package notify

import (
	"context"
	"strings"

	"github.com/blang/semver"
	"github.com/google/go-github/github"
	"github.com/pkg/errors"
)

const (
	// project GitHub organization name
	ghorg = "redhat-developer"
	// project GitHub repository name
	ghrepo = "odo"
	// URL of the installation shell script
	InstallScriptURL = "https://raw.githubusercontent.com/redhat-developer/odo/master/scripts/install.sh"
)

// getLatestReleaseTag polls odo's upstream GitHub repository to get the
// tag of the latest release
func getLatestReleaseTag() (string, error) {
	client := github.NewClient(nil)
	release, response, err := client.Repositories.GetLatestRelease(context.Background(), ghorg, ghrepo)
	if response != nil {
		defer response.Body.Close()
	}
	if err != nil {
		return "", errors.Wrap(err, "error getting latest release")
	}
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
