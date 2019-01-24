package version

import (
	"fmt"
	"os"
	"strings"

	"github.com/golang/glog"
	"github.com/redhat-developer/odo/pkg/notify"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubernetes/pkg/kubectl/cmd/templates"
)

var (
	// VERSION  is version number that will be displayed when running ./odo version
	VERSION = "v0.0.17"

	// GITCOMMIT is hash of the commit that wil be displayed when running ./odo version
	// this will be overwritten when running  build like this: go build -ldflags="-X github.com/redhat-developer/odo/cmd.GITCOMMIT=$(GITCOMMIT)"
	// HEAD is default indicating that this was not set during build
	GITCOMMIT = "HEAD"
)

// RecommendedCommandName is the recommended version command name
const RecommendedCommandName = "version"

var versionLongDesc = ktemplates.LongDesc("Print the client version information")

var versionExample = ktemplates.Examples(`
# Print the client version of Odo
%[1]s`,
)

// VersionOptions encapsulates all options for odo version command
type VersionOptions struct {
	// clientFlag indicates if the user only wants client information
	clientFlag bool
	// generic context options common to all commands
	*genericclioptions.Context
}

// NewVersionOptions creates a new VersionOptions instance
func NewVersionOptions() *VersionOptions {
	return &VersionOptions{}
}

// Complete completes VersionOptions after they have been created
func (o *VersionOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	return
}

// Validate validates the VersionOptions based on completed values
func (o *VersionOptions) Validate() (err error) {
	// TODO : Figure out what validations need to be added
	return
}

// Run contains the logic for the odo service create command
func (o *VersionOptions) Run(cmd *cobra.Command) (err error) {
	// If verbose mode is enabled, dump all KUBECLT_* env variables
	// this is usefull for debuging oc plugin integration
	for _, v := range os.Environ() {
		if strings.HasPrefix(v, "KUBECTL_") {
			glog.V(4).Info(v)
		}
	}

	fmt.Println("odo " + VERSION + " (" + GITCOMMIT + ")")

	if !o.clientFlag {
		// Let's fetch the info about the server
		serverInfo, err := genericclioptions.ClientWithConnectionCheck(cmd, true).GetServerVersion()
		if err == nil {
			// make sure we only include Openshift info if we actually have it
			openshiftStr := ""
			if len(serverInfo.OpenShiftVersion) > 0 {
				openshiftStr = fmt.Sprintf("OpenShift: %v\n", serverInfo.OpenShiftVersion)
			}
			fmt.Printf("\n"+
				"Server: %v\n"+
				"%v"+
				"Kubernetes: %v\n",
				serverInfo.Address,
				openshiftStr,
				serverInfo.KubernetesVersion)
		}
	}
	return
}

// NewCmdVersion implements the version odo command
func NewCmdVersion(name, fullName string) *cobra.Command {
	o := NewVersionOptions()
	// versionCmd represents the version command
	var versionCmd = &cobra.Command{
		Use:     name,
		Short:   versionLongDesc,
		Long:    versionLongDesc,
		Example: fmt.Sprintf(versionExample, fullName),
		Run: func(cmd *cobra.Command, args []string) {
			util.LogErrorAndExit(o.Complete(name, cmd, args), "")
			util.LogErrorAndExit(o.Validate(), "")
			util.LogErrorAndExit(o.Run(cmd), "")
		},
	}

	// Add a defined annotation in order to appear in the help menu
	versionCmd.Annotations = map[string]string{"command": "utility"}
	versionCmd.SetUsageTemplate(util.CmdUsageTemplate)
	versionCmd.Flags().BoolVar(&o.clientFlag, "client", false, "Client version only (no server required).")

	return versionCmd
}

// GetLatestReleaseInfo Gets information about the latest release
func GetLatestReleaseInfo(info chan<- string) {
	newTag, err := notify.CheckLatestReleaseTag(VERSION)
	if err != nil {
		// The error is intentionally not being handled because we don't want
		// to stop the execution of the program because of this failure
		glog.V(4).Infof("Error checking if newer odo release is available: %v", err)
	}
	if len(newTag) > 0 {
		info <- "---\n" +
			"A newer version of odo (version: " + fmt.Sprint(newTag) + ") is available.\n" +
			"Update using your package manager, or run\n" +
			"curl " + notify.InstallScriptURL + " | sh\n" +
			"to update manually, or visit https://github.com/redhat-developer/odo/releases\n" +
			"---\n" +
			"If you wish to disable the update notifications, you can disable it by running\n" +
			"'odo utils config set UpdateNotification false'\n"
	}
}
