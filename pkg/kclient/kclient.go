package kclient

import (
	"strings"
	"time"

	"github.com/openshift/odo/pkg/util"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/watch"
	appsclientset "k8s.io/client-go/kubernetes/typed/apps/v1"
	"k8s.io/klog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
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
	waitForComponentDeletionTimeout = 120 * time.Second
)

// Client is a collection of fields used for client configuration and interaction
type Client struct {
	KubeClient       kubernetes.Interface
	KubeConfig       clientcmd.ClientConfig
	KubeClientConfig *rest.Config
	Namespace        string
	OperatorClient   *operatorsclientset.OperatorsV1alpha1Client
	appsClient       appsclientset.AppsV1Interface
	// DynamicClient interacts with client-go's `dynamic` package. It is used
	// to dynamically create service from an operator. It can take an arbitrary
	// yaml and create k8s/OpenShift resource from it.
	DynamicClient dynamic.Interface
}

// New creates a new client
func New() (*Client, error) {
	return NewForConfig(nil)
}

// NewForConfig creates a new client with the provided configuration or initializes the configuration if none is provided
func NewForConfig(config clientcmd.ClientConfig) (client *Client, err error) {
	if config == nil {
		// initialize client-go clients
		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		configOverrides := &clientcmd.ConfigOverrides{}
		config = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	}

	client = new(Client)
	client.KubeConfig = config

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

	client.DynamicClient, err = dynamic.NewForConfig(client.KubeClientConfig)
	if err != nil {
		return nil, err
	}

	appsClient, err := appsclientset.NewForConfig(client.KubeClientConfig)
	if err != nil {
		return nil, err
	}
	client.appsClient = appsClient

	return client, nil
}

// Delete takes labels as a input and based on it, deletes respective resource
func (c *Client) Delete(labels map[string]string, wait bool) error {

	// convert labels to selector
	selector := util.ConvertLabelsToSelector(labels)
	klog.V(3).Infof("Selectors used for deletion: %s", selector)

	var errorList []string
	var deletionPolicy = metav1.DeletePropagationBackground

	// for --wait flag, it deletes component dependents first and then delete component
	if wait {
		deletionPolicy = metav1.DeletePropagationForeground
	}
	// Delete Deployments
	klog.V(3).Info("Deleting Deployments")
	err := c.appsClient.Deployments(c.Namespace).DeleteCollection(&metav1.DeleteOptions{PropagationPolicy: &deletionPolicy}, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		errorList = append(errorList, "unable to delete deployments")
	}

	// for --wait it waits for component to be deleted
	// TODO: Need to modify for `odo app delete`, currently wait flag is added only in `odo component delete`
	//       so only one component gets passed in selector
	if wait {
		err = c.WaitForComponentDeletion(selector)
		if err != nil {
			errorList = append(errorList, err.Error())
		}
	}

	// Error string
	errString := strings.Join(errorList, ",")
	if len(errString) != 0 {
		return errors.New(errString)
	}
	return nil

}

// WaitForComponentDeletion waits for component to be deleted
func (c *Client) WaitForComponentDeletion(selector string) error {

	klog.V(3).Infof("Waiting for component to get deleted")

	watcher, err := c.appsClient.Deployments(c.Namespace).Watch(metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return err
	}
	defer watcher.Stop()
	eventCh := watcher.ResultChan()

	for {
		select {
		case event, ok := <-eventCh:
			_, typeOk := event.Object.(*appsv1.Deployment)
			if !ok || !typeOk {
				return errors.New("Unable to watch deployments")
			}
			if event.Type == watch.Deleted {
				klog.V(3).Infof("WaitForComponentDeletion, Event Received:Deleted")
				return nil
			} else if event.Type == watch.Error {
				klog.V(3).Infof("WaitForComponentDeletion, Event Received:Deleted ")
				return errors.New("Unable to watch deployments")
			}
		case <-time.After(waitForComponentDeletionTimeout):
			klog.V(3).Infof("WaitForComponentDeletion, Timeout")
			return errors.New("Time out waiting for component to get deleted")
		}
	}
}

// GeneratePortForwardReq builds a port forward request
func (c *Client) GeneratePortForwardReq(podName string) *rest.Request {
	return c.KubeClient.CoreV1().RESTClient().
		Post().
		Resource("pods").
		Namespace(c.Namespace).
		Name(podName).
		SubResource("portforward")
}
