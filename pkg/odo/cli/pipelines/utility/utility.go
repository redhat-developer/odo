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

// RemoveEmptyStrings returns a slice with all the empty strings removed from the
// source slice.
func RemoveEmptyStrings(s []string) []string {
	nonempty := []string{}
	for _, v := range s {
		if v != "" {
			nonempty = append(nonempty, v)
		}
	}
	return nonempty
}

func MaybeCompletePrefix(s string) string {
	if s != "" && !strings.HasSuffix(s, "-") {
		return s + "-"
	}
	return s
}
