package e2e

import "strings"

func determineRouteURL() string {
	return strings.TrimSpace(runCmd("odo url list  | sed -n '1!p' | awk 'FNR==2 { print $2 }'"))
}
