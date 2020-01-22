package kclient

import (
	"github.com/pkg/errors"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	// api clientsets
	operatorsclientset "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned/typed/operators/v1alpha1"

	// api resourcetypes
	olmv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
)

// errorMsg is the message for user when invalid configuration error occurs
const errorMsg = `
Please ensure you have an active kubernetes context to your cluster. 
Consult your Kubernetes distribution's documentation for more details
`

// Client is a collection of fields used for client configuration and interaction
type Client struct {
	KubeClient       kubernetes.Interface
	KubeConfig       clientcmd.ClientConfig
	KubeClientConfig *rest.Config
	Namespace        string
	OperatorClient   *operatorsclientset.OperatorsV1alpha1Client
}

// New creates a new client
func New() (*Client, error) {
	var client Client
	var err error

	// initialize client-go clients
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	client.KubeConfig = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	client.KubeClientConfig, err = client.KubeConfig.ClientConfig()
	if err != nil {
		return nil, errors.Wrapf(err, errorMsg)
	}

	client.KubeClient, err = kubernetes.NewForConfig(client.KubeClientConfig)
	if err != nil {
		return nil, err
	}

	client.Namespace, _, err = client.KubeConfig.Namespace()
	if err != nil {
		return nil, err
	}

	client.OperatorClient, err = operatorsclientset.NewForConfig(client.KubeClientConfig)
	if err != nil {
		return nil, err
	}

	return &client, nil
}

func (c *Client) GetClusterServiceVersions() (olmv1alpha1.ClusterVersionList, error) {
	csvs, err := c.OperatorClient.ClusterServiceVersions(c.Namespace).List(v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return csvs, nil
}
