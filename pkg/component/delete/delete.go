package delete

import "github.com/redhat-developer/odo/pkg/kclient"

type DeleteComponentClient struct {
	kubeClient kclient.ClientInterface
}

func NewDeleteComponentClient(kubeClient kclient.ClientInterface) *DeleteComponentClient {
	return &DeleteComponentClient{
		kubeClient: kubeClient,
	}
}
