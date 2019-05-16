package printtemplates

import "fmt"

// PrintPushMessage tells user to use odo push apply action on specified object (what) on
// the cluster
func PrintPushMessage(action string, what string) {
	fmt.Printf("To %s %s on the OpenShift Cluster, please use `odo push` \n", action, what)
}
