package version

import (
	"context"
	"fmt"
	"github.com/redhat-developer/odo/pkg/api"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/commonflags"
	"github.com/redhat-developer/odo/pkg/podman"
	"os"
	"strings"

	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	odoversion "github.com/redhat-developer/odo/pkg/version"

	"github.com/spf13/cobra"
	"k8s.io/klog"
	ktemplates "k8s.io/kubectl/pkg/util/templates"

	"github.com/redhat-developer/odo/pkg/odo/util"
)

// RecommendedCommandName is the recommended version command name
const RecommendedCommandName = "version"

// OdoReleasesPage is the GitHub page where we do all our releases
const OdoReleasesPage = "https://github.com/redhat-developer/odo/releases"

var versionLongDesc = ktemplates.LongDesc("Print the client version information")

var versionExample = ktemplates.Examples(`
# Print the client version of odo
%[1]s`,
)

// VersionOptions encapsulates all options for odo version command
type VersionOptions struct {
	// Flags
	clientFlag bool

	// serverInfo contains the remote server information if the user asked for it, nil otherwise
	serverInfo *kclient.ServerInfo
	podmanInfo podman.SystemVersionReport
	clientset  *clientset.Clientset
}

var _ genericclioptions.Runnable = (*VersionOptions)(nil)
var _ genericclioptions.JsonOutputter = (*VersionOptions)(nil)

// NewVersionOptions creates a new VersionOptions instance
func NewVersionOptions() *VersionOptions {
	return &VersionOptions{}
}

func (o *VersionOptions) SetClientset(clientset *clientset.Clientset) {
	o.clientset = clientset
}

// Complete completes VersionOptions after they have been created
func (o *VersionOptions) Complete(ctx context.Context, cmdline cmdline.Cmdline, args []string) (err error) {
	if !o.clientFlag {
		// Let's fetch the info about the server, ignoring errors
		if o.clientset.KubernetesClient != nil {
			o.serverInfo, err = o.clientset.KubernetesClient.GetServerVersion(o.clientset.PreferenceClient.GetTimeout())
			if err != nil {
				klog.V(4).Info("unable to fetch the server version: ", err)
			}
		}

		if o.clientset.PodmanClient != nil {
			o.podmanInfo, err = o.clientset.PodmanClient.Version(ctx)
			if err != nil {
				klog.V(4).Info("unable to fetch the podman client version: ", err)
			}
		}
	}

	if o.serverInfo == nil {
		log.Warning("unable to fetch the cluster server version")
	}
	if o.podmanInfo.Client == nil {
		log.Warning("unable to fetch the podman client version")
	}
	return nil
}

// Validate validates the VersionOptions based on completed values
func (o *VersionOptions) Validate(ctx context.Context) (err error) {
	return nil
}

func (o *VersionOptions) RunForJsonOutput(ctx context.Context) (out interface{}, err error) {
	return o.run(), nil
}

func (o *VersionOptions) run() api.OdoVersion {
	result := api.OdoVersion{
		Version:   odoversion.VERSION,
		GitCommit: odoversion.GITCOMMIT,
	}

	if o.clientFlag {
		return result
	}

	if o.serverInfo != nil {
		clusterInfo := &api.ClusterInfo{
			ServerURL:  o.serverInfo.Address,
			Kubernetes: api.ClusterClientInfo{Version: o.serverInfo.KubernetesVersion},
			OpenShift:  api.ClusterClientInfo{Version: o.serverInfo.OpenShiftVersion},
		}
		result.Cluster = clusterInfo
	}

	if o.podmanInfo.Client != nil {
		podmanInfo := &api.PodmanInfo{Client: api.PodmanClientInfo{Version: o.podmanInfo.Client.Version}}
		result.Podman = podmanInfo
	}

	return result
}

// Run contains the logic for the odo service create command
func (o *VersionOptions) Run(ctx context.Context) (err error) {
	// If verbose mode is enabled, dump all KUBECTL_* env variables
	// this is useful for debugging oc plugin integration
	for _, v := range os.Environ() {
		if strings.HasPrefix(v, "KUBECTL_") {
			klog.V(4).Info(v)
		}
	}

	odoVersion := o.run()
	fmt.Println("odo " + odoVersion.Version + " (" + odoVersion.GitCommit + ")")

	if o.clientFlag {
		return nil
	}

	message := "\n"
	if odoVersion.Cluster != nil {
		cluster := odoVersion.Cluster
		message += fmt.Sprintf("Server: %v\n", cluster.ServerURL)

		// make sure we only include OpenShift info if we actually have it
		if cluster.OpenShift.Version != "" {
			message += fmt.Sprintf("OpenShift: %v\n", cluster.OpenShift.Version)
		}

		message += fmt.Sprintf("Kubernetes: %v\n", cluster.Kubernetes.Version)
	}

	if odoVersion.Podman != nil {
		message += fmt.Sprintf("Podman Client: %v\n", odoVersion.Podman.Client.Version)
	}

	fmt.Print(message)

	return nil
}

// NewCmdVersion implements the version odo command
func NewCmdVersion(name, fullName string, testClientset clientset.Clientset) *cobra.Command {
	o := NewVersionOptions()
	// versionCmd represents the version command
	var versionCmd = &cobra.Command{
		Use:     name,
		Short:   versionLongDesc,
		Long:    versionLongDesc,
		Example: fmt.Sprintf(versionExample, fullName),
		RunE: func(cmd *cobra.Command, args []string) error {
			return genericclioptions.GenericRun(o, testClientset, cmd, args)
		},
	}
	commonflags.UseOutputFlag(versionCmd)
	clientset.Add(versionCmd, clientset.PREFERENCE, clientset.KUBERNETES_NULLABLE, clientset.PODMAN_NULLABLE)
	util.SetCommandGroup(versionCmd, util.UtilityGroup)

	versionCmd.SetUsageTemplate(util.CmdUsageTemplate)
	versionCmd.Flags().BoolVar(&o.clientFlag, "client", false, "Client version only (no server required).")

	return versionCmd
}

// GetLatestReleaseInfo Gets information about the latest release
func GetLatestReleaseInfo(info chan<- string) {
	newTag, err := checkLatestReleaseTag(odoversion.VERSION)
	if err != nil {
		// The error is intentionally not being handled because we don't want
		// to stop the execution of the program because of this failure
		klog.V(4).Infof("Error checking if newer odo release is available: %v", err)
	}
	if len(newTag) > 0 {
		info <- fmt.Sprintf(`
---
A newer version of odo (%s) is available,
visit %s to update.
If you wish to disable this notification, run:
odo preference set UpdateNotification false
---`, fmt.Sprint(newTag), OdoReleasesPage)

	}
}
