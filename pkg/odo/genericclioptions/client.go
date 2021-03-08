package genericclioptions

import (
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/occlient"
	"github.com/spf13/cobra"
)

// Client returns an oc client "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"d for this command's options
func Client(command *cobra.Command) (*occlient.Client, error) {
	return client(command)
}

// client creates an oc client based on the command flags
func client(command *cobra.Command) (*occlient.Client, error) {
	client, err := occlient.New()
	if err != nil {
		return nil, err
	}

	return client, nil
}

// kClient creates an kclient based on the command flags
func kClient(command *cobra.Command) (*kclient.Client, error) {
	kClient, err := kclient.New()
	if err != nil {
		return nil, err
	}
	return kClient, nil
}
