package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
)

var (
	// VERSION  is version number that will be displayed when running ./odo version
	VERSION = "v0.0.12"

	// GITCOMMIT is hash of the commit that wil be displayed when running ./odo version
	// this will be overwritten when running  build like this: go build -ldflags="-X github.com/redhat-developer/odo/cmd.GITCOMMIT=$(GITCOMMIT)"
	// HEAD is default indicating that this was not set during build
	GITCOMMIT = "HEAD"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the client version information",
	Long:  "Print the client version information",
	Example: `  # Print the client version of Odo
  odo version
	`,
	Run: func(cmd *cobra.Command, args []string) {

		// If verbose mode is enabled, dump all KUBECLT_* env variables
		// this is usefull for debuging oc plugin integration
		for _, v := range os.Environ() {
			if strings.HasPrefix(v, "KUBECTL_") {
				glog.V(4).Info(v)
			}
		}

		fmt.Println("odo " + VERSION + " (" + GITCOMMIT + ")")

		// Lets fetch the info about the server
		serverInfo, err := getOcClient().GetServerVersion()
		checkError(err, "")
		fmt.Printf("\n"+
			"Server: %v\n"+
			"OpenShift: %v\n"+
			"Kubernetes: %v\n",
			serverInfo.Address,
			serverInfo.OpenShiftVersion,
			serverInfo.KubernetesVersion)
	},
}

func init() {
	// Add a defined annotation in order to appear in the help menu
	versionCmd.Annotations = map[string]string{"command": "utility"}
	versionCmd.SetUsageTemplate(cmdUsageTemplate)

	rootCmd.AddCommand(versionCmd)
}
