package kclient

import (
	"github.com/pkg/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth" // Required for Kube clusters which use auth plugins
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	// api clientsets
	operatorsclientset "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned/typed/operators/v1alpha1"
)

const (
	// errorMsg is the message for user when invalid configuration error occurs
	errorMsg = `
Please ensure you have an active kubernetes context to your cluster. 
Consult your Kubernetes distribution's documentation for more details
`
)

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

// CreateObjectMeta creates a common object meta
func CreateObjectMeta(name, namespace string, labels, annotations map[string]string) metav1.ObjectMeta {

	objectMeta := metav1.ObjectMeta{
		Name:        name,
		Namespace:   namespace,
		Labels:      labels,
		Annotations: annotations,
	}

	return objectMeta
}
