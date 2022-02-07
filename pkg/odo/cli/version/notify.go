package version

import (
	"strings"

	"github.com/blang/semver"
	"github.com/pkg/errors"
	"github.com/redhat-developer/odo/pkg/util"
)

const (
	// VersionFetchURL is the URL to fetch latest version number
	VersionFetchURL = "https://raw.githubusercontent.com/redhat-developer/odo/master/build/VERSION"
)

// getLatestReleaseTag polls odo's upstream GitHub repository to get the
// tag of the latest release
func getLatestReleaseTag() (string, error) {

	request := util.HTTPRequestParams{
		URL: VersionFetchURL,
	}

	// Make request and cache response for 60 minutes
	body, err := util.HTTPGetRequest(request, 60)
	if err != nil {
		return "", errors.Wrap(err, "error getting latest release")
	}

	return strings.TrimSuffix(string(body), "\n"), nil
}

// checkLatestReleaseTag returns the latest release tag if a newer latest
// release is available, else returns an empty string
func checkLatestReleaseTag(currentVersion string) (string, error) {
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
