package genericclioptions

import (
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/occlient"
	"github.com/openshift/odo/pkg/odo/util"
	"github.com/spf13/cobra"
)

// Client returns an oc client "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"d for this command's options
func Client(command *cobra.Command) *occlient.Client {
	return client(command)
}

// client creates an oc client based on the command flags
func client(command *cobra.Command) *occlient.Client {
	client, err := occlient.New()
	util.LogErrorAndExit(err, "")

	return client
}

// kClient creates an kclient based on the command flags
func kClient(command *cobra.Command) *kclient.Client {
	kClient, err := kclient.New()
	util.LogErrorAndExit(err, "")

	return kClient
}
