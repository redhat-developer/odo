package kclient

import (
	"github.com/pkg/errors"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	// errorMsg is the message for user when invalid configuration error occurs
	errorMsg = `
Please ensure you have an active kubernetes context to your cluster. 
Consult your Kubernetes distribution's documentation for more details
`

	// The length of the string to be generated for names of resources
	nameLength = 5
)

// Client is a collection of fields used for client configuration and interaction
type Client struct {
	KubeClient       kubernetes.Interface
	KubeConfig       clientcmd.ClientConfig
	KubeClientConfig *rest.Config
	Namespace        string
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

// GenerateOwnerReference genertes an ownerReference  from the deployment which can then be set as
// owner for various Kubernetes objects and ensure that when the owner object is deleted from the
// cluster, all other objects are automatically removed by Kubernetes garbage collector
func GenerateOwnerReference(deployment *appsv1.Deployment) metav1.OwnerReference {

	ownerReference := metav1.OwnerReference{
		APIVersion: "extensions/v1beta1",
		Kind:       "Deployment",
		Name:       deployment.Name,
		UID:        deployment.UID,
	}

	return ownerReference
}
