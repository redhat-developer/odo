package genericclioptions

import (
	"github.com/openshift/odo/v2/pkg/kclient"
	"github.com/openshift/odo/v2/pkg/occlient"
)

// Client returns an oc client with the kClient set
func Client() (*occlient.Client, error) {
	ocClient, err := ocClient()
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

// ocClient creates an oc client
func ocClient() (*occlient.Client, error) {
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
