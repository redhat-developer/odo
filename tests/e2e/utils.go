package e2e

import "strings"

// determineRouteURL returns the http URL where the current component exposes it's service
// this URL can then be used in order to interact with the deployed service running in Openshift
// keeping with the spirit of the e2e tests, this expects, odo, sed and awk to be on the PATH
func determineRouteURL() string {
	output := runCmdShouldPass("odo url list  | sed -n '1!p' | awk 'FNR==2 { print $2 }'")
	return strings.TrimSpace(output)
}
