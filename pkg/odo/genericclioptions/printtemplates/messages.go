package printtemplates

import (
	"fmt"
)

// PushMessage tells user to use odo push apply action on specified object (what) on
// the cluster
func PushMessage(action string, what string) string {
	return fmt.Sprintf("To %s %s on the OpenShift Cluster, please use `odo push` \n", action, what)
}
