package kclient

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/blang/semver"
	servicecatalogclienset "github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/discovery"
	appsclientset "k8s.io/client-go/kubernetes/typed/apps/v1"
	"k8s.io/klog"

	"github.com/openshift/odo/pkg/util"
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
	KubeClient           kubernetes.Interface
	KubeConfig           clientcmd.ClientConfig
	KubeClientConfig     *rest.Config
	Namespace            string
	OperatorClient       *operatorsclientset.OperatorsV1alpha1Client
	appsClient           appsclientset.AppsV1Interface
	serviceCatalogClient servicecatalogclienset.ServicecatalogV1beta1Interface
	// DynamicClient interacts with client-go's `dynamic` package. It is used
	// to dynamically create service from an operator. It can take an arbitrary
	// yaml and create k8s/OpenShift resource from it.
	DynamicClient      dynamic.Interface
	discoveryClient    discovery.DiscoveryInterface
	supportedResources map[string]bool
	// Is server side apply supported by cluster
	// Use IsSSASupported()
	isSSASupported                     *bool
	isNetworkingV1IngressSupported     bool
	isExtensionV1Beta1IngressSupported bool
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

	client.appsClient, err = appsclientset.NewForConfig(client.KubeClientConfig)
	if err != nil {
		return nil, err
	}

	client.serviceCatalogClient, err = servicecatalogclienset.NewForConfig(client.KubeClientConfig)
	if err != nil {
		return nil, err
	}

	client.discoveryClient, err = discovery.NewDiscoveryClientForConfig(client.KubeClientConfig)
	if err != nil {
		return nil, err
	}

	err = client.checkIngressSupport()
	if err != nil {
		return nil, err
	}

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
	err := c.appsClient.Deployments(c.Namespace).DeleteCollection(context.TODO(), metav1.DeleteOptions{PropagationPolicy: &deletionPolicy}, metav1.ListOptions{LabelSelector: selector})
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

	watcher, err := c.appsClient.Deployments(c.Namespace).Watch(context.TODO(), metav1.ListOptions{LabelSelector: selector})
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

func (c *Client) SetDiscoveryInterface(client discovery.DiscoveryInterface) {
	c.discoveryClient = client
}

func (c *Client) IsResourceSupported(apiGroup, apiVersion, resourceName string) (bool, error) {
	klog.V(4).Infof("Checking if %q resource supported", resourceName)

	if c.supportedResources == nil {
		c.supportedResources = make(map[string]bool, 7)
	}
	groupVersion := metav1.GroupVersion{Group: apiGroup, Version: apiVersion}.String()
	resource := metav1.GroupVersionResource{Group: apiGroup, Version: apiVersion, Resource: resourceName}
	groupVersionResource := resource.String()

	supported, found := c.supportedResources[groupVersionResource]
	if !found {
		list, err := c.discoveryClient.ServerResourcesForGroupVersion(groupVersion)
		if err != nil {
			if kerrors.IsNotFound(err) {
				supported = false
			} else {
				// don't record, just attempt again next time in case it's a transient error
				return false, err
			}
		} else {
			for _, resources := range list.APIResources {
				if resources.Name == resourceName {
					supported = true
					break
				}
			}
		}
		c.supportedResources[groupVersionResource] = supported
	}
	return supported, nil
}

// IsSSASupported checks if Server Side Apply is supported by cluster
// SSA was introduced in Kubernetes 1.16
// If there is an error while parsing versions, it assumes that SSA is supported by cluster.
// Most of clusters these days are 1.16 and up
func (c *Client) IsSSASupported() bool {
	// check if this was done before so we don't query cluster multiple times for the same info
	if c.isSSASupported == nil {
		versionWithSSA, err := semver.Make("1.16.0")
		if err != nil {
			klog.Warningf("unable to parse version %q", err)
		}

		kVersion, err := c.discoveryClient.ServerVersion()
		if err != nil {
			klog.Warningf("unable to get k8s server version %q", err)
			return true
		}
		klog.V(4).Infof("Kubernetes version is %q", kVersion.String())

		cleanupVersion := strings.TrimLeft(kVersion.String(), "v")
		serverVersion, err := semver.Make(cleanupVersion)
		if err != nil {
			klog.Warningf("unable to parse k8s server version %q", err)
			return true
		}

		isSSASupported := versionWithSSA.LE(serverVersion)
		c.isSSASupported = &isSSASupported

		klog.V(4).Infof("Cluster has support for SSA: %t", *c.isSSASupported)
	}
	return *c.isSSASupported

}

func (c *Client) checkIngressSupport() error {
	var err error
	c.isNetworkingV1IngressSupported, err = c.IsResourceSupported("networking.k8s.io", "v1", "ingresses")
	if err != nil {
		return fmt.Errorf("failed to check networking v1 ingress support %w", err)
	}
	c.isExtensionV1Beta1IngressSupported, err = c.IsResourceSupported("extensions", "v1beta1", "ingresses")
	if err != nil {
		return fmt.Errorf("failed to check extensions v1beta1 ingress support %w", err)
	}
	return nil
}
