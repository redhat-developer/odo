package utility

import (
	"strings"

	"github.com/openshift/odo/pkg/log"
)

// AddGitSuffixIfNecessary will append .git to URL if necessary
func AddGitSuffixIfNecessary(url string) string {
	if url == "" || strings.HasSuffix(strings.ToLower(url), ".git") {
		return url
	}
	log.Infof("Adding .git to %s", url)
	return url + ".git"
}
