package genericclioptions

import (
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/occlient"
)

// Client returns an oc client "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"d for this command's options
func Client() (*occlient.Client, error) {
	ocClient, err := client()
	if err != nil {
		return nil, err
	}
	kClient, err := kClient()
	if err != nil {
		return nil, err
	}
	ocClient.SetKubeClient(kClient)
	return ocClient, nil
}

// client creates an oc client
func client() (*occlient.Client, error) {
	client, err := occlient.New()
	if err != nil {
		return nil, err
	}

	return client, nil
}

// kClient creates an kclient
func kClient() (*kclient.Client, error) {
	kClient, err := kclient.New()
	if err != nil {
		return nil, err
	}
	return kClient, nil
}
