package version

import (
	"fmt"
	"strings"

	"github.com/blang/semver"

	dfutil "github.com/devfile/library/pkg/util"
)

const (
	// VersionFetchURL is the URL to fetch latest version number
	VersionFetchURL = "https://raw.githubusercontent.com/redhat-developer/odo/main/build/VERSION"
)

// getLatestReleaseTag polls odo's upstream GitHub repository to get the
// tag of the latest release
func getLatestReleaseTag() (string, error) {

	request := dfutil.HTTPRequestParams{
		URL: VersionFetchURL,
	}

	// Make request and cache response for 60 minutes
	body, err := dfutil.HTTPGetRequest(request, 60)
	if err != nil {
		return "", fmt.Errorf("error getting latest release: %w", err)
	}

	return strings.TrimSuffix(string(body), "\n"), nil
}

// checkLatestReleaseTag returns the latest release tag if a newer latest
// release is available, else returns an empty string
func checkLatestReleaseTag(currentVersion string) (string, error) {
	currentSemver, err := semver.Make(strings.TrimPrefix(currentVersion, "v"))
	if err != nil {
		return "", fmt.Errorf("unable to make semver from the current version: %v: %w", currentVersion, err)
	}

	latestTag, err := getLatestReleaseTag()
	if err != nil {
		return "", fmt.Errorf("unable to get latest release tag: %w", err)
	}

	latestSemver, err := semver.Make(strings.TrimPrefix(latestTag, "v"))
	if err != nil {
		return "", fmt.Errorf("unable to make semver from the latest release tag: %v: %w", latestTag, err)
	}

	if currentSemver.LT(latestSemver) {
		return latestTag, nil
	}

	return "", nil
}
