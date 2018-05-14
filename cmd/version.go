package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	// VERSION  is version number that will be displayed when running ./odo version
	VERSION = "v0.0.5"

	// GITCOMMIT is hash of the commit that wil be displayed when running ./odo version
	// this will be overwritten when running  build like this: go build -ldflags="-X github.com/redhat-developer/odo/cmd.GITCOMMIT=$(GITCOMMIT)"
	// HEAD is default indicating that this was not set during build
	GITCOMMIT = "HEAD"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of odo",
	Run: func(cmd *cobra.Command, args []string) {

		// If verbose mode is enabled, dump all KUBECLT_* env variables
		// this is usefull for debuging oc plugin integration
		if GlobalVerbose {
			for _, v := range os.Environ() {
				if strings.HasPrefix(v, "KUBECTL_") {
					fmt.Println(v)
				}
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
	rootCmd.AddCommand(versionCmd)
}
