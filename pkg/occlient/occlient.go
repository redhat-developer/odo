package occlient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/preference"
	"github.com/openshift/odo/pkg/util"

	// api clientsets
	appsclientset "github.com/openshift/client-go/apps/clientset/versioned/typed/apps/v1"
	buildclientset "github.com/openshift/client-go/build/clientset/versioned/typed/build/v1"
	imageclientset "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"
	projectclientset "github.com/openshift/client-go/project/clientset/versioned/typed/project/v1"
	routeclientset "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	userclientset "github.com/openshift/client-go/user/clientset/versioned/typed/user/v1"

	oauthv1client "github.com/openshift/client-go/oauth/clientset/versioned/typed/oauth/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
)

// CreateArgs is a container of attributes of component create action
type CreateArgs struct {
	Name            string
	SourcePath      string
	SourceRef       string
	ImageName       string
	EnvVars         []string
	Ports           []string
	Resources       *corev1.ResourceRequirements
	ApplicationName string
	Wait            bool
	// StorageToBeMounted describes the storage to be created
	// storagePath is the key of the map, the generatedPVC is the value of the map
	StorageToBeMounted map[string]*corev1.PersistentVolumeClaim
	StdOut             io.Writer
}

const (
	OpenShiftNameSpace = "openshift"

	// timeout for waiting for project deletion
	waitForProjectDeletionTimeOut = 3 * time.Minute

	// The length of the string to be generated for names of resources
	nameLength = 5
)

// UpdateComponentParams serves the purpose of holding the arguments to a component update request
type UpdateComponentParams struct {
	// CommonObjectMeta is the object meta containing the labels and annotations expected for the new deployment
	CommonObjectMeta metav1.ObjectMeta
	// ResourceLimits are the cpu and memory constraints to be applied on to the component
	ResourceLimits corev1.ResourceRequirements
	// EnvVars to be exposed
	EnvVars []corev1.EnvVar
	// StorageToBeMounted describes the storage to be mounted
	// storagePath is the key of the map, the generatedPVC is the value of the map
	StorageToBeMounted map[string]*corev1.PersistentVolumeClaim
	// StorageToBeUnMounted describes the storage to be unmounted
	// path is the key of the map,storageName is the value of the map
	StorageToBeUnMounted map[string]string
}

// errorMsg is the message for user when invalid configuration error occurs
const errorMsg = `
Please login to your server: 

odo login https://mycluster.mydomain.com
`

type Client struct {
	kubeClient    *kclient.Client
	imageClient   imageclientset.ImageV1Interface
	appsClient    appsclientset.AppsV1Interface
	buildClient   buildclientset.BuildV1Interface
	projectClient projectclientset.ProjectV1Interface
	routeClient   routeclientset.RouteV1Interface
	userClient    userclientset.UserV1Interface
	KubeConfig    clientcmd.ClientConfig
	Namespace     string
}

// New creates a new client
func New() (*Client, error) {
	var client Client

	// initialize client-go clients
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	client.KubeConfig = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	config, err := client.KubeConfig.ClientConfig()
	if err != nil {
		return nil, errors.New(err.Error() + errorMsg)
	}

	client.kubeClient, err = kclient.NewForConfig(client.KubeConfig)
	if err != nil {
		return nil, err
	}

	client.imageClient, err = imageclientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	client.appsClient, err = appsclientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	client.buildClient, err = buildclientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	client.projectClient, err = projectclientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	client.routeClient, err = routeclientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	client.userClient, err = userclientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	client.Namespace, _, err = client.KubeConfig.Namespace()
	if err != nil {
		return nil, err
	}

	return &client, nil
}

// ParseImageName parse image reference
// returns (imageNamespace, imageName, tag, digest, error)
// if image is referenced by tag (name:tag)  than digest is ""
// if image is referenced by digest (name@digest) than  tag is ""
func ParseImageName(image string) (string, string, string, string, error) {
	digestParts := strings.Split(image, "@")
	if len(digestParts) == 2 {
		// image is references digest
		// Safe path image name and digest are non empty, else error
		if digestParts[0] != "" && digestParts[1] != "" {
			// Image name might be fully qualified name of form: Namespace/ImageName
			imangeNameParts := strings.Split(digestParts[0], "/")
			if len(imangeNameParts) == 2 {
				return imangeNameParts[0], imangeNameParts[1], "", digestParts[1], nil
			}
			return "", imangeNameParts[0], "", digestParts[1], nil
		}
	} else if len(digestParts) == 1 && digestParts[0] != "" { // Filter out empty image name
		tagParts := strings.Split(image, ":")
		if len(tagParts) == 2 {
			// ":1.0.0 is invalid image name"
			if tagParts[0] != "" {
				// Image name might be fully qualified name of form: Namespace/ImageName
				imangeNameParts := strings.Split(tagParts[0], "/")
				if len(imangeNameParts) == 2 {
					return imangeNameParts[0], imangeNameParts[1], tagParts[1], "", nil
				}
				return "", tagParts[0], tagParts[1], "", nil
			}
		} else if len(tagParts) == 1 {
			// Image name might be fully qualified name of form: Namespace/ImageName
			imangeNameParts := strings.Split(tagParts[0], "/")
			if len(imangeNameParts) == 2 {
				return imangeNameParts[0], imangeNameParts[1], "latest", "", nil
			}
			return "", tagParts[0], "latest", "", nil
		}
	}
	return "", "", "", "", fmt.Errorf("invalid image reference %s", image)

}

// RunLogout logs out the current user from cluster
func (c *Client) RunLogout(stdout io.Writer) error {
	output, err := c.userClient.Users().Get(context.TODO(), "~", metav1.GetOptions{})
	if err != nil {
		klog.V(1).Infof("%v : unable to get userinfo", err)
	}

	// read the current config form ~/.kube/config
	conf, err := c.KubeConfig.ClientConfig()
	if err != nil {
		klog.V(1).Infof("%v : unable to get client config", err)
	}
	// initialising oauthv1client
	client, err := oauthv1client.NewForConfig(conf)
	if err != nil {
		klog.V(1).Infof("%v : unable to create a new OauthV1Client", err)
	}

	// deleting token form the server
	if err := client.OAuthAccessTokens().Delete(context.TODO(), conf.BearerToken, metav1.DeleteOptions{}); err != nil {
		klog.V(1).Infof("%v", err)
	}

	rawConfig, err := c.KubeConfig.RawConfig()
	if err != nil {
		klog.V(1).Infof("%v : unable to switch to  project", err)
	}

	// deleting token for the current server from local config
	for key, value := range rawConfig.AuthInfos {
		if key == rawConfig.Contexts[rawConfig.CurrentContext].AuthInfo {
			value.Token = ""
		}
	}
	err = clientcmd.ModifyConfig(clientcmd.NewDefaultClientConfigLoadingRules(), rawConfig, true)
	if err != nil {
		klog.V(1).Infof("%v : unable to write config to config file", err)
	}

	_, err = io.WriteString(stdout, fmt.Sprintf("Logged \"%v\" out on \"%v\"\n", output.Name, conf.Host))
	return err
}

// isServerUp returns true if server is up and running
// server parameter has to be a valid url
func isServerUp(server string) bool {
	// initialising the default timeout, this will be used
	// when the value is not readable from config
	ocRequestTimeout := preference.DefaultTimeout * time.Second
	// checking the value of timeout in config
	// before proceeding with default timeout
	cfg, configReadErr := preference.New()
	if configReadErr != nil {
		klog.V(3).Info(errors.Wrap(configReadErr, "unable to read config file"))
	} else {
		ocRequestTimeout = time.Duration(cfg.GetTimeout()) * time.Second
	}
	address, err := util.GetHostWithPort(server)
	if err != nil {
		klog.V(3).Infof("Unable to parse url %s (%s)", server, err)
	}
	klog.V(3).Infof("Trying to connect to server %s", address)
	_, connectionError := net.DialTimeout("tcp", address, time.Duration(ocRequestTimeout))
	if connectionError != nil {
		klog.V(3).Info(errors.Wrap(connectionError, "unable to connect to server"))
		return false
	}

	klog.V(3).Infof("Server %v is up", server)
	return true
}

// ServerInfo contains the fields that contain the server's information like
// address, OpenShift and Kubernetes versions
type ServerInfo struct {
	Address           string
	OpenShiftVersion  string
	KubernetesVersion string
}

// GetServerVersion will fetch the Server Host, OpenShift and Kubernetes Version
// It will be shown on the execution of odo version command
func (c *Client) GetServerVersion() (*ServerInfo, error) {
	var info ServerInfo

	// This will fetch the information about Server Address
	config, err := c.KubeConfig.ClientConfig()
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get server's address")
	}
	info.Address = config.Host

	// checking if the server is reachable
	if !isServerUp(config.Host) {
		return nil, errors.New("Unable to connect to OpenShift cluster, is it down?")
	}

	// fail fast if user is not connected (same logic as `oc whoami`)
	_, err = c.userClient.Users().Get(context.TODO(), "~", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// This will fetch the information about OpenShift Version
	coreGet := c.kubeClient.KubeClient.CoreV1().RESTClient().Get()
	rawOpenShiftVersion, err := coreGet.AbsPath("/version/openshift").Do(context.TODO()).Raw()
	if err != nil {
		// when using Minishift (or plain 'oc cluster up' for that matter) with OKD 3.11, the version endpoint is missing...
		klog.V(3).Infof("Unable to get OpenShift Version - endpoint '/version/openshift' doesn't exist")
	} else {
		var openShiftVersion version.Info
		if err := json.Unmarshal(rawOpenShiftVersion, &openShiftVersion); err != nil {
			return nil, errors.Wrapf(err, "unable to unmarshal OpenShift version %v", string(rawOpenShiftVersion))
		}
		info.OpenShiftVersion = openShiftVersion.GitVersion
	}

	// This will fetch the information about Kubernetes Version
	rawKubernetesVersion, err := coreGet.AbsPath("/version").Do(context.TODO()).Raw()
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get Kubernetes Version")
	}
	var kubernetesVersion version.Info
	if err := json.Unmarshal(rawKubernetesVersion, &kubernetesVersion); err != nil {
		return nil, errors.Wrapf(err, "unable to unmarshal Kubernetes Version: %v", string(rawKubernetesVersion))
	}
	info.KubernetesVersion = kubernetesVersion.GitVersion

	return &info, nil
}

func (c *Client) GetKubeClient() *kclient.Client {
	return c.kubeClient
}

func (c *Client) SetKubeClient(client *kclient.Client) {
	c.kubeClient = client
}
