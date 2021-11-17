package genericclioptions

import (
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/occlient"
)

// Client returns an oc client with the kClient set
func Client() (*occlient.Client, error) {
	ocClient, err := occlient.New()
	if err != nil {
		return nil, err
	}
	kClient, err := kclient.New()
	if err != nil {
		return nil, err
	}
	ocClient.SetKubeClient(kClient)
	return ocClient, nil
}
