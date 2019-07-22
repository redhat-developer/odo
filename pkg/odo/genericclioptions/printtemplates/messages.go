package printtemplates

import (
	"fmt"
)

// PushMessage tells user to use odo push apply action on specified object (what) on
// the cluster
func PushMessage(action string, what string, config bool) string {
	var c string
	if config {
		c = " --config"
	}
	return fmt.Sprintf("To %s %s on the OpenShift Cluster, please use `odo push%s` \n", action, what, c)
}
