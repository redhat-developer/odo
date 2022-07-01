package kclient

import (
	"fmt"
	"strings"

	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/blang/semver"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"

	// api clientsets
	servicecatalogclienset "github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	projectclientset "github.com/openshift/client-go/project/clientset/versioned/typed/project/v1"
	routeclientset "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	userclientset "github.com/openshift/client-go/user/clientset/versioned/typed/user/v1"
	operatorsclientset "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned/typed/operators/v1alpha1"
	appsclientset "k8s.io/client-go/kubernetes/typed/apps/v1"

	_ "k8s.io/client-go/plugin/pkg/client/auth" // Required for Kube clusters which use auth plugins
)

const (
	// errorMsg is the message for user when invalid configuration error occurs
	errorMsg = `
Please ensure you have an active kubernetes context to your cluster. 
Consult your Kubernetes distribution's documentation for more details.
Error: %w
`
	defaultQPS   = 200
	defaultBurst = 200
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
	DynamicClient   dynamic.Interface
	discoveryClient discovery.DiscoveryInterface
	restmapper      *restmapper.DeferredDiscoveryRESTMapper

	supportedResources map[string]bool
	// Is server side apply supported by cluster
	// Use IsSSASupported()
	isSSASupported *bool
	// checkIngressSupports is used to check ingress support
	// (used to prevent duplicate checks and disable check in UTs)
	checkIngressSupports               bool
	isNetworkingV1IngressSupported     bool
	isExtensionV1Beta1IngressSupported bool

	// openshift clients
	userClient    userclientset.UserV1Interface
	projectClient projectclientset.ProjectV1Interface
	routeClient   routeclientset.RouteV1Interface
}

var _ ClientInterface = (*Client)(nil)

// New creates a new client
func New() (*Client, error) {
	return NewForConfig(nil)
}

func (c *Client) GetClient() kubernetes.Interface {
	return c.KubeClient
}

func (c *Client) GetConfig() clientcmd.ClientConfig {
	return c.KubeConfig
}

func (c *Client) GetClientConfig() *rest.Config {
	return c.KubeClientConfig
}

func (c *Client) GetDynamicClient() dynamic.Interface {
	return c.DynamicClient
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
		return nil, fmt.Errorf(errorMsg, err)
	}

	// For the rest CLIENT, we set the QPS and Burst to high values so
	// we do not receive throttling error messages when using the REST client.
	// Inadvertently, this also increases the speed of which we use the REST client
	// to safe values without increased error / query information.
	// See issue: https://github.com/kubernetes/client-go/issues/610
	// and reference implementation: https://github.com/vmware-tanzu/tanzu-framework/pull/1656
	client.KubeClientConfig.QPS = defaultQPS
	client.KubeClientConfig.Burst = defaultBurst

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

	noWarningConfig := rest.CopyConfig(client.KubeClientConfig)
	// set the warning handler for this client to ignore warnings
	noWarningConfig.WarningHandler = rest.NoWarnings{}
	client.DynamicClient, err = dynamic.NewForConfig(noWarningConfig)
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

	config_flags := genericclioptions.NewConfigFlags(true)
	client.discoveryClient, err = config_flags.ToDiscoveryClient()
	if err != nil {
		return nil, err
	}

	client.restmapper = restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(client.discoveryClient))

	client.checkIngressSupports = true

	client.userClient, err = userclientset.NewForConfig(client.KubeClientConfig)
	if err != nil {
		return nil, err
	}

	client.projectClient, err = projectclientset.NewForConfig(client.KubeClientConfig)
	if err != nil {
		return nil, err
	}

	client.routeClient, err = routeclientset.NewForConfig(client.KubeClientConfig)
	if err != nil {
		return nil, err
	}

	return client, nil
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
	klog.V(4).Infof("Checking if %q resource is supported", resourceName)

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
