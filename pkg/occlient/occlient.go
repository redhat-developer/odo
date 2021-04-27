package occlient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/openshift/odo/pkg/config"
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

	// api resource types
	appsv1 "github.com/openshift/api/apps/v1"
	imagev1 "github.com/openshift/api/image/v1"
	oauthv1client "github.com/openshift/client-go/oauth/clientset/versioned/typed/oauth/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/version"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
)

// CreateArgs is a container of attributes of component create action
type CreateArgs struct {
	Name            string
	SourcePath      string
	SourceRef       string
	SourceType      config.SrcType
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
	OcUpdateTimeout                 = 5 * time.Minute
	OpenShiftNameSpace              = "openshift"
	waitForComponentDeletionTimeout = 120 * time.Second

	// timeout for waiting for project deletion
	waitForProjectDeletionTimeOut = 3 * time.Minute

	// The length of the string to be generated for names of resources
	nameLength = 5

	// SupervisordVolumeName Create a custom name and (hope) that users don't use the *exact* same name in their deploymentConfig
	SupervisordVolumeName = "odo-supervisord-shared-data"

	// EnvS2IScriptsURL is an env var exposed to https://github.com/openshift/odo-init-image/blob/master/assemble-and-restart to indicate location of s2i scripts in this case assemble script
	EnvS2IScriptsURL = "ODO_S2I_SCRIPTS_URL"

	// EnvS2IScriptsProtocol is an env var exposed to https://github.com/openshift/odo-init-image/blob/master/assemble-and-restart to indicate the way to access location of s2i scripts indicated by ${${EnvS2IScriptsURL}} above
	EnvS2IScriptsProtocol = "ODO_S2I_SCRIPTS_PROTOCOL"

	// EnvS2ISrcOrBinPath is an env var exposed by s2i to indicate where the builder image expects the component source or binary to reside
	EnvS2ISrcOrBinPath = "ODO_S2I_SRC_BIN_PATH"

	// EnvS2ISrcBackupDir is the env var that points to the directory that holds a backup of component source
	// This is required bcoz, s2i assemble script moves(hence deletes contents) the contents of $ODO_S2I_SRC_BIN_PATH to $APP_ROOT during which $APP_DIR alo needs to be empty so that mv doesn't complain pushing to an already exisiting dir with same name
	EnvS2ISrcBackupDir = "ODO_SRC_BACKUP_DIR"

	// S2IScriptsURLLabel S2I script location Label name
	// Ref: https://docs.openshift.com/enterprise/3.2/creating_images/s2i.html#build-process
	S2IScriptsURLLabel = "io.openshift.s2i.scripts-url"

	// S2IBuilderImageName is the S2I builder image name
	S2IBuilderImageName = "name"

	// S2ISrcOrBinLabel is the label that provides, path where S2I expects component source or binary
	S2ISrcOrBinLabel = "io.openshift.s2i.destination"

	// EnvS2IBuilderImageName is the label that provides the name of builder image in component
	EnvS2IBuilderImageName = "ODO_S2I_BUILDER_IMG"

	// EnvS2IDeploymentDir is an env var exposed to https://github.com/openshift/odo-init-image/blob/master/assemble-and-restart to indicate s2i deployment directory
	EnvS2IDeploymentDir = "ODO_S2I_DEPLOYMENT_DIR"

	// DefaultS2ISrcOrBinPath is the default path where S2I expects source/binary artifacts in absence of $S2ISrcOrBinLabel in builder image
	// Ref: https://github.com/openshift/source-to-image/blob/master/docs/builder_image.md#required-image-contents
	DefaultS2ISrcOrBinPath = "/tmp"

	// DefaultS2ISrcBackupDir is the default path where odo backs up the component source
	DefaultS2ISrcBackupDir = "/opt/app-root/src-backup"

	// EnvS2IWorkingDir is an env var to odo-init-image assemble-and-restart.sh to indicate to it the s2i working directory
	EnvS2IWorkingDir = "ODO_S2I_WORKING_DIR"

	DefaultAppRootDir = "/opt/app-root"
)

// S2IPaths is a struct that will hold path to S2I scripts and the protocol indicating access to them, component source/binary paths, artifacts deployments directory
// These are passed as env vars to component pod
type S2IPaths struct {
	ScriptsPathProtocol string
	ScriptsPath         string
	SrcOrBinPath        string
	DeploymentDir       string
	WorkingDir          string
	SrcBackupPath       string
	BuilderImgName      string
}

// UpdateComponentParams serves the purpose of holding the arguments to a component update request
type UpdateComponentParams struct {
	// CommonObjectMeta is the object meta containing the labels and annotations expected for the new deployment
	CommonObjectMeta metav1.ObjectMeta
	// ResourceLimits are the cpu and memory constraints to be applied on to the component
	ResourceLimits corev1.ResourceRequirements
	// EnvVars to be exposed
	EnvVars []corev1.EnvVar
	// ExistingDC is the dc of the existing component that is requested for an update
	ExistingDC *appsv1.DeploymentConfig
	// DcRollOutWaitCond holds the logic to wait for dc with requested updates to be applied
	DcRollOutWaitCond dcRollOutWait
	// ImageMeta describes the image to be used in dc(builder image for local/binary and built component image for git deployments)
	ImageMeta CommonImageMeta
	// StorageToBeMounted describes the storage to be mounted
	// storagePath is the key of the map, the generatedPVC is the value of the map
	StorageToBeMounted map[string]*corev1.PersistentVolumeClaim
	// StorageToBeUnMounted describes the storage to be unmounted
	// path is the key of the map,storageName is the value of the map
	StorageToBeUnMounted map[string]string
}

// S2IDeploymentsDir is a set of possible S2I labels that provides S2I deployments directory
// This label is not uniform across different builder images. This slice is expected to grow as odo adds support to more component types and/or the respective builder images use different labels
var S2IDeploymentsDir = []string{
	"com.redhat.deployments-dir",
	"org.jboss.deployments-dir",
	"org.jboss.container.deployments-dir",
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

// addLabelsToArgs adds labels from map to args as a new argument in format that oc requires
// --labels label1=value1,label2=value2
func addLabelsToArgs(labels map[string]string, args []string) []string {
	if labels != nil {
		var labelsString []string
		for key, value := range labels {
			labelsString = append(labelsString, fmt.Sprintf("%s=%s", key, value))
		}
		args = append(args, "--labels")
		args = append(args, strings.Join(labelsString, ","))
	}

	return args
}

func getAppRootVolumeName(dcName string) string {
	return fmt.Sprintf("%s-s2idata", dcName)
}

// NewAppS2I is only used with "Git" as we need Build
// gitURL is the url of the git repo
// inputPorts is the array containing the string port values
// envVars is the array containing the string env var values
func (c *Client) NewAppS2I(params CreateArgs, commonObjectMeta metav1.ObjectMeta) error {
	klog.V(3).Infof("Using BuilderImage: %s", params.ImageName)
	imageNS, imageName, imageTag, _, err := ParseImageName(params.ImageName)
	if err != nil {
		return errors.Wrap(err, "unable to parse image name")
	}
	imageStream, err := c.GetImageStream(imageNS, imageName, imageTag)
	if err != nil {
		return errors.Wrap(err, "unable to retrieve ImageStream for NewAppS2I")
	}
	/*
	 Set imageNS to the commonObjectMeta.Namespace of above fetched imagestream because, the commonObjectMeta.Namespace passed here can potentially be emptystring
	 in which case, GetImageStream function resolves to correct commonObjectMeta.Namespace in accordance with priorities in GetImageStream
	*/

	imageNS = imageStream.ObjectMeta.Namespace
	klog.V(3).Infof("Using imageNS: %s", imageNS)

	imageStreamImage, err := c.GetImageStreamImage(imageStream, imageTag)
	if err != nil {
		return errors.Wrapf(err, "unable to create s2i app for %s", commonObjectMeta.Name)
	}

	var containerPorts []corev1.ContainerPort
	if len(params.Ports) == 0 {
		containerPorts, err = getExposedPortsFromISI(imageStreamImage)
		if err != nil {
			return errors.Wrapf(err, "unable to get exposed ports for %s:%s", imageName, imageTag)
		}
	} else {
		if err != nil {
			return errors.Wrapf(err, "unable to create s2i app for %s", commonObjectMeta.Name)
		}
		containerPorts, err = util.GetContainerPortsFromStrings(params.Ports)
		if err != nil {
			return errors.Wrapf(err, "unable to get container ports from %v", params.Ports)
		}
	}

	inputEnvVars, err := kclient.GetInputEnvVarsFromStrings(params.EnvVars)
	if err != nil {
		return errors.Wrapf(err, "error adding environment variables to the container")
	}

	// generate and create ImageStream
	is := imagev1.ImageStream{
		ObjectMeta: commonObjectMeta,
	}
	_, err = c.imageClient.ImageStreams(c.Namespace).Create(context.TODO(), &is, metav1.CreateOptions{FieldManager: kclient.FieldManager})
	if err != nil {
		return errors.Wrapf(err, "unable to create ImageStream for %s", commonObjectMeta.Name)
	}

	// if gitURL is not set, error out
	if params.SourcePath == "" {
		return errors.New("unable to create buildSource with empty gitURL")
	}

	// Deploy BuildConfig to build the container with Git
	buildConfig, err := c.CreateBuildConfig(commonObjectMeta, params.ImageName, params.SourcePath, params.SourceRef, inputEnvVars)
	if err != nil {
		return errors.Wrapf(err, "unable to deploy BuildConfig for %s", commonObjectMeta.Name)
	}

	// Generate and create the DeploymentConfig
	dc := generateGitDeploymentConfig(commonObjectMeta, buildConfig.Spec.Output.To.Name, containerPorts, inputEnvVars, params.Resources)
	err = addOrRemoveVolumeAndVolumeMount(c, &dc, params.StorageToBeMounted, nil)
	if err != nil {
		return errors.Wrapf(err, "failed to mount and unmount pvc to dc")
	}
	createdDC, err := c.appsClient.DeploymentConfigs(c.Namespace).Create(context.TODO(), &dc, metav1.CreateOptions{FieldManager: kclient.FieldManager})
	if err != nil {
		return errors.Wrapf(err, "unable to create DeploymentConfig for %s", commonObjectMeta.Name)
	}

	ownerReference := GenerateOwnerReference(createdDC)

	// update the owner references for the new storage
	for _, storage := range params.StorageToBeMounted {
		err := c.GetKubeClient().GetAndUpdateStorageOwnerReference(storage, ownerReference)
		if err != nil {
			return errors.Wrapf(err, "unable to update owner reference of storage")
		}
	}

	// Create a service
	commonObjectMeta.SetOwnerReferences(append(commonObjectMeta.GetOwnerReferences(), ownerReference))
	service := corev1.Service{
		ObjectMeta: commonObjectMeta,
		Spec:       generateServiceSpec(commonObjectMeta, dc.Spec.Template.Spec.Containers[0].Ports),
	}
	svc, err := c.GetKubeClient().CreateService(service)
	if err != nil {
		return errors.Wrapf(err, "unable to create Service for %s", commonObjectMeta.Name)
	}

	// Create secret(s)
	err = c.GetKubeClient().CreateSecrets(params.Name, commonObjectMeta, svc, ownerReference)

	return err
}

// getS2ILabelValue returns the requested S2I label value from the passed set of labels attached to builder image
// and the hard coded possible list(the labels are not uniform across different builder images) of expected labels
func getS2ILabelValue(labels map[string]string, expectedLabelsSet []string) string {
	for _, label := range expectedLabelsSet {
		if retVal, ok := labels[label]; ok {
			return retVal
		}
	}
	return ""
}

// uniqueAppendOrOverwriteEnvVars appends/overwrites the passed existing list of env vars with the elements from the to-be appended passed list of envs
func uniqueAppendOrOverwriteEnvVars(existingEnvs []corev1.EnvVar, envVars ...corev1.EnvVar) []corev1.EnvVar {
	mapExistingEnvs := make(map[string]corev1.EnvVar)
	var retVal []corev1.EnvVar

	// Convert slice of existing env vars to map to check for existence
	for _, envVar := range existingEnvs {
		mapExistingEnvs[envVar.Name] = envVar
	}

	// For each new envVar to be appended, Add(if envVar with same name doesn't already exist) / overwrite(if envVar with same name already exists) the map
	for _, newEnvVar := range envVars {
		mapExistingEnvs[newEnvVar.Name] = newEnvVar
	}

	// append the values to the final slice
	// don't loop because we need them in order
	for _, envVar := range existingEnvs {
		if val, ok := mapExistingEnvs[envVar.Name]; ok {
			retVal = append(retVal, val)
			delete(mapExistingEnvs, envVar.Name)
		}
	}

	for _, newEnvVar := range envVars {
		if val, ok := mapExistingEnvs[newEnvVar.Name]; ok {
			retVal = append(retVal, val)
		}
	}

	return retVal
}

// deleteEnvVars deletes the passed env var from the list of passed env vars
// Parameters:
//	existingEnvs: Slice of existing env vars
//	envTobeDeleted: The name of env var to be deleted
// Returns:
//	slice of env vars with delete reflected
func deleteEnvVars(existingEnvs []corev1.EnvVar, envTobeDeleted string) []corev1.EnvVar {
	retVal := make([]corev1.EnvVar, len(existingEnvs))
	copy(retVal, existingEnvs)
	for ind, envVar := range retVal {
		if envVar.Name == envTobeDeleted {
			retVal = append(retVal[:ind], retVal[ind+1:]...)
			break
		}
	}
	return retVal
}

// BootstrapSupervisoredS2I uses S2I (Source To Image) to inject Supervisor into the application container.
// Odo uses https://github.com/ochinchina/supervisord which is pre-built in a ready-to-deploy InitContainer.
// The supervisord binary is copied over to the application container using a temporary volume and overrides
// the built-in S2I run function for the supervisord run command instead.
//
// Supervisor keeps the pod running (as PID 1), so you it is possible to trigger assembly script inside running pod,
// and than restart application using Supervisor without need to restart the container/Pod.
//
func (c *Client) BootstrapSupervisoredS2I(params CreateArgs, commonObjectMeta metav1.ObjectMeta) error {
	imageNS, imageName, imageTag, _, err := ParseImageName(params.ImageName)

	if err != nil {
		return errors.Wrap(err, "unable to create new s2i git build ")
	}
	imageStream, err := c.GetImageStream(imageNS, imageName, imageTag)
	if err != nil {
		return errors.Wrap(err, "Failed to bootstrap supervisored")
	}
	/*
	 Set imageNS to the commonObjectMeta.Namespace of above fetched imagestream because, the commonObjectMeta.Namespace passed here can potentially be emptystring
	 in which case, GetImageStream function resolves to correct commonObjectMeta.Namespace in accordance with priorities in GetImageStream
	*/
	imageNS = imageStream.ObjectMeta.Namespace

	imageStreamImage, err := c.GetImageStreamImage(imageStream, imageTag)
	if err != nil {
		return errors.Wrap(err, "unable to bootstrap supervisord")
	}
	var containerPorts []corev1.ContainerPort
	containerPorts, err = util.GetContainerPortsFromStrings(params.Ports)
	if err != nil {
		return errors.Wrapf(err, "unable to get container ports from %v", params.Ports)
	}

	inputEnvs, err := kclient.GetInputEnvVarsFromStrings(params.EnvVars)
	if err != nil {
		return errors.Wrapf(err, "error adding environment variables to the container")
	}

	// generate and create ImageStream
	is := imagev1.ImageStream{
		ObjectMeta: commonObjectMeta,
	}
	_, err = c.imageClient.ImageStreams(c.Namespace).Create(context.TODO(), &is, metav1.CreateOptions{FieldManager: kclient.FieldManager})
	if err != nil {
		return errors.Wrapf(err, "unable to create ImageStream for %s", commonObjectMeta.Name)
	}

	commonImageMeta := CommonImageMeta{
		Name:      imageName,
		Tag:       imageTag,
		Namespace: imageNS,
		Ports:     containerPorts,
	}

	// Extract s2i scripts path and path type from imagestream image
	//s2iScriptsProtocol, s2iScriptsURL, s2iSrcOrBinPath, s2iDestinationDir
	s2iPaths, err := getS2IMetaInfoFromBuilderImg(imageStreamImage)
	if err != nil {
		return errors.Wrap(err, "unable to bootstrap supervisord")
	}

	// Append s2i related parameters extracted above to env
	inputEnvs = injectS2IPaths(inputEnvs, s2iPaths)

	if params.SourceType == config.LOCAL {
		inputEnvs = uniqueAppendOrOverwriteEnvVars(
			inputEnvs,
			corev1.EnvVar{
				Name:  EnvS2ISrcBackupDir,
				Value: s2iPaths.SrcBackupPath,
			},
		)
	}

	// Generate the DeploymentConfig that will be used.
	dc := generateSupervisordDeploymentConfig(
		commonObjectMeta,
		commonImageMeta,
		inputEnvs,
		[]corev1.EnvFromSource{},
		params.Resources,
	)
	if len(inputEnvs) != 0 {
		err = updateEnvVar(&dc, inputEnvs)
		if err != nil {
			return errors.Wrapf(err, "unable to add env vars to the container")
		}
	}

	addInitVolumesToDC(&dc, commonObjectMeta.Name, s2iPaths.DeploymentDir)

	err = addOrRemoveVolumeAndVolumeMount(c, &dc, params.StorageToBeMounted, nil)
	if err != nil {
		return errors.Wrapf(err, "failed to mount and unmount pvc to dc")
	}

	createdDC, err := c.appsClient.DeploymentConfigs(c.Namespace).Create(context.TODO(), &dc, metav1.CreateOptions{FieldManager: kclient.FieldManager})
	if err != nil {
		return errors.Wrapf(err, "unable to create DeploymentConfig for %s", commonObjectMeta.Name)
	}

	var jsonDC []byte
	jsonDC, _ = json.Marshal(createdDC)
	klog.V(5).Infof("Created new DeploymentConfig:\n%s\n", string(jsonDC))

	ownerReference := GenerateOwnerReference(createdDC)

	// update the owner references for the new storage
	for _, storage := range params.StorageToBeMounted {
		err := c.GetKubeClient().GetAndUpdateStorageOwnerReference(storage, ownerReference)
		if err != nil {
			return errors.Wrapf(err, "unable to update owner reference of storage")
		}
	}

	// Create a service
	commonObjectMeta.SetOwnerReferences(append(commonObjectMeta.GetOwnerReferences(), ownerReference))
	service := corev1.Service{
		ObjectMeta: commonObjectMeta,
		Spec:       generateServiceSpec(commonObjectMeta, dc.Spec.Template.Spec.Containers[0].Ports),
	}
	svc, err := c.GetKubeClient().CreateService(service)
	if err != nil {
		return errors.Wrapf(err, "unable to create Service for %s", commonObjectMeta.Name)
	}

	err = c.GetKubeClient().CreateSecrets(params.Name, commonObjectMeta, svc, ownerReference)
	if err != nil {
		return err
	}

	// Setup PVC.
	_, err = c.CreatePVC(getAppRootVolumeName(commonObjectMeta.Name), "1Gi", commonObjectMeta.Labels, ownerReference)
	if err != nil {
		return errors.Wrapf(err, "unable to create PVC for %s", commonObjectMeta.Name)
	}

	return nil
}

// updateEnvVar updates the environmental variables to the container in the DC
// dc is the deployment config to be updated
// envVars is the array containing the corev1.EnvVar values
func updateEnvVar(dc *appsv1.DeploymentConfig, envVars []corev1.EnvVar) error {
	numContainers := len(dc.Spec.Template.Spec.Containers)
	if numContainers != 1 {
		return fmt.Errorf("expected exactly one container in Deployment Config %v, got %v", dc.Name, numContainers)
	}

	dc.Spec.Template.Spec.Containers[0].Env = envVars
	return nil
}

// Define a function that is meant to update a DC in place
type dcStructUpdater func(dc *appsv1.DeploymentConfig) error
type dcRollOutWait func(*appsv1.DeploymentConfig, int64) bool

// PatchCurrentDC "patches" the current DeploymentConfig with a new one
// however... we make sure that configurations such as:
// - volumes
// - environment variables
// are correctly copied over / consistent without an issue.
// if prePatchDCHandler is specified (meaning not nil), then it's applied
// as the last action before the actual call to the Kubernetes API thus giving us the chance
// to perform arbitrary updates to a DC before it's finalized for patching
// isGit indicates if the deployment config belongs to a git component or a local/binary component
func (c *Client) PatchCurrentDC(dc appsv1.DeploymentConfig, prePatchDCHandler dcStructUpdater, existingCmpContainer corev1.Container, ucp UpdateComponentParams, isGit bool) error {

	name := ucp.CommonObjectMeta.Name
	currentDC := ucp.ExistingDC
	modifiedDC := *currentDC

	waitCond := ucp.DcRollOutWaitCond

	// copy the any remaining volumes and volume mounts
	copyVolumesAndVolumeMounts(dc, currentDC, existingCmpContainer)

	if prePatchDCHandler != nil {
		err := prePatchDCHandler(&dc)
		if err != nil {
			return errors.Wrapf(err, "Unable to correctly update dc %s using the specified prePatch handler", name)
		}
	}

	// now mount/unmount the newly created/deleted pvc
	err := addOrRemoveVolumeAndVolumeMount(c, &dc, ucp.StorageToBeMounted, ucp.StorageToBeUnMounted)
	if err != nil {
		return err
	}

	// Replace the current spec with the new one
	modifiedDC.Spec = dc.Spec

	// Replace the old annotations with the new ones too
	// the reason we do this is because Kubernetes handles metadata such as resourceVersion
	// that should not be overridden.
	modifiedDC.ObjectMeta.Annotations = dc.ObjectMeta.Annotations
	modifiedDC.ObjectMeta.Labels = dc.ObjectMeta.Labels

	// Update the current one that's deployed with the new Spec.
	// despite the "patch" function name, we use update since `.Patch` requires
	// use to define each and every object we must change. Updating makes it easier.
	updatedDc, err := c.appsClient.DeploymentConfigs(c.Namespace).Update(context.TODO(), &modifiedDC, metav1.UpdateOptions{FieldManager: kclient.FieldManager})

	if err != nil {
		return errors.Wrapf(err, "unable to update DeploymentConfig %s", name)
	}

	// if isGit is true, the DC belongs to a git component
	// since build happens for every push in case of git and a new image is pushed, we need to wait
	// so git oriented deployments, we start the deployment before waiting for it to be updated
	if isGit {
		_, err := c.StartDeployment(updatedDc.Name)
		if err != nil {
			return errors.Wrapf(err, "unable to start deployment")
		}
	} else {
		// not a git oriented deployment, check before waiting
		// we check after the update that the template in the earlier and the new dc are same or not
		// if they are same, we don't wait as new deployment won't run and we will wait till timeout
		// inspired from https://github.com/openshift/origin/blob/bb1b9b5223dd37e63790d99095eec04bfd52b848/pkg/apps/controller/deploymentconfig/deploymentconfig_controller.go#L609
		if reflect.DeepEqual(updatedDc.Spec.Template, currentDC.Spec.Template) {
			return nil
		}
		currentDCBytes, errCurrent := json.Marshal(currentDC.Spec.Template)
		updatedDCBytes, errUpdated := json.Marshal(updatedDc.Spec.Template)
		if errCurrent != nil || errUpdated != nil {
			return errors.Wrapf(err, "unable to unmarshal dc")
		}
		klog.V(3).Infof("going to wait for new deployment roll out because updatedDc Spec.Template: %v doesn't match currentDc Spec.Template: %v", string(updatedDCBytes), string(currentDCBytes))

	}

	// We use the currentDC + 1 for the next revision.. We do NOT use the updated DC (see above code)
	// as the "Update" function will not update the Status.LatestVersion quick enough... so we wait until
	// the current revision + 1 is available.
	desiredRevision := currentDC.Status.LatestVersion + 1

	// Watch / wait for deploymentconfig to update annotations
	// importing "component" results in an import loop, so we do *not* use the constants here.
	_, err = c.WaitAndGetDC(name, desiredRevision, OcUpdateTimeout, waitCond)
	if err != nil {
		return errors.Wrapf(err, "unable to wait for DeploymentConfig %s to update", name)
	}

	// update the owner references for the new storage
	for _, storage := range ucp.StorageToBeMounted {
		err := c.GetKubeClient().GetAndUpdateStorageOwnerReference(storage, GenerateOwnerReference(updatedDc))
		if err != nil {
			return errors.Wrapf(err, "unable to update owner reference of storage")
		}
	}

	return nil
}

// copies volumes and volume mounts from currentDC to dc, excluding the supervisord related ones
func copyVolumesAndVolumeMounts(dc appsv1.DeploymentConfig, currentDC *appsv1.DeploymentConfig, matchingContainer corev1.Container) {
	// Append the existing VolumeMounts to the new DC. We use "range" and find the correct container rather than
	// using .spec.Containers[0] *in case* the template ever changes and a new container has been added.
	for index, container := range dc.Spec.Template.Spec.Containers {
		// Find the container
		if container.Name == matchingContainer.Name {

			// create a map of volume mount names for faster searching later
			dcVolumeMountsMap := make(map[string]bool)
			for _, volumeMount := range container.VolumeMounts {
				dcVolumeMountsMap[volumeMount.Name] = true
			}

			// Loop through all the volumes
			for _, volume := range matchingContainer.VolumeMounts {
				// If it's the supervisord volume, ignore it.
				if volume.Name == SupervisordVolumeName {
					continue
				} else {
					// check if we are appending the same volume mount again or not
					if _, ok := dcVolumeMountsMap[volume.Name]; !ok {
						dc.Spec.Template.Spec.Containers[index].VolumeMounts = append(dc.Spec.Template.Spec.Containers[index].VolumeMounts, volume)
					}
				}
			}

			// Break out since we've succeeded in updating the container we were looking for
			break
		}
	}

	// create a map of volume names for faster searching later
	dcVolumeMap := make(map[string]bool)
	for _, volume := range dc.Spec.Template.Spec.Volumes {
		dcVolumeMap[volume.Name] = true
	}

	// Now the same with Volumes, again, ignoring the supervisord volume.
	for _, volume := range currentDC.Spec.Template.Spec.Volumes {
		if volume.Name == SupervisordVolumeName {
			continue
		} else {
			// check if we are appending the same volume again or not
			if _, ok := dcVolumeMap[volume.Name]; !ok {
				dc.Spec.Template.Spec.Volumes = append(dc.Spec.Template.Spec.Volumes, volume)
			}
		}
	}
}

// UpdateDCToGit replaces / updates the current DeplomentConfig with the appropriate
// generated image from BuildConfig as well as the correct DeploymentConfig triggers for Git.
func (c *Client) UpdateDCToGit(ucp UpdateComponentParams, isDeleteSupervisordVolumes bool) (err error) {

	// Find the container (don't want to use .Spec.Containers[0] in case the user has modified the DC...)
	existingCmpContainer, err := FindContainer(ucp.ExistingDC.Spec.Template.Spec.Containers, ucp.CommonObjectMeta.Name)
	if err != nil {
		return errors.Wrapf(err, "Unable to find container %s", ucp.CommonObjectMeta.Name)
	}

	// Fail if blank
	if ucp.ImageMeta.Name == "" {
		return errors.New("UpdateDCToGit imageName cannot be blank")
	}

	dc := generateGitDeploymentConfig(ucp.CommonObjectMeta, ucp.ImageMeta.Name, ucp.ImageMeta.Ports, ucp.EnvVars, &ucp.ResourceLimits)

	if isDeleteSupervisordVolumes {
		// Patch the current DC
		err = c.PatchCurrentDC(
			dc,
			removeTracesOfSupervisordFromDC,
			existingCmpContainer,
			ucp,
			true,
		)

		if err != nil {
			return errors.Wrapf(err, "unable to update the current DeploymentConfig %s", ucp.CommonObjectMeta.Name)
		}

		// Cleanup after the supervisor
		err = c.GetKubeClient().DeletePVC(getAppRootVolumeName(ucp.CommonObjectMeta.Name))
		if err != nil {
			return errors.Wrapf(err, "unable to delete S2I data PVC from %s", ucp.CommonObjectMeta.Name)
		}
	} else {
		err = c.PatchCurrentDC(
			dc,
			nil,
			existingCmpContainer,
			ucp,
			true,
		)
	}

	if err != nil {
		return errors.Wrapf(err, "unable to update the current DeploymentConfig %s", ucp.CommonObjectMeta.Name)
	}

	return nil
}

// UpdateDCToSupervisor updates the current DeploymentConfig to a SupervisorD configuration.
// Parameters:
//	commonObjectMeta: dc meta object
//	componentImageType: type of builder image
//	isToLocal: bool used to indicate if component is to be updated to local in which case a source backup dir will be injected into component env
//  isCreatePVC bool used to indicate if a new supervisorD PVC should be created during the update
// Returns:
//	errors if any or nil
func (c *Client) UpdateDCToSupervisor(ucp UpdateComponentParams, isToLocal bool, createPVC bool) error {

	existingCmpContainer, err := FindContainer(ucp.ExistingDC.Spec.Template.Spec.Containers, ucp.CommonObjectMeta.Name)
	if err != nil {
		return errors.Wrapf(err, "Unable to find container %s", ucp.CommonObjectMeta.Name)
	}

	// Retrieve the namespace of the corresponding component image
	imageStream, err := c.GetImageStream(ucp.ImageMeta.Namespace, ucp.ImageMeta.Name, ucp.ImageMeta.Tag)
	if err != nil {
		return errors.Wrap(err, "unable to get image stream for CreateBuildConfig")
	}
	ucp.ImageMeta.Namespace = imageStream.ObjectMeta.Namespace

	imageStreamImage, err := c.GetImageStreamImage(imageStream, ucp.ImageMeta.Tag)
	if err != nil {
		return errors.Wrap(err, "unable to bootstrap supervisord")
	}

	s2iPaths, err := getS2IMetaInfoFromBuilderImg(imageStreamImage)
	if err != nil {
		return errors.Wrap(err, "unable to bootstrap supervisord")
	}

	cmpContainer := ucp.ExistingDC.Spec.Template.Spec.Containers[0]

	// Append s2i related parameters extracted above to env
	inputEnvs := injectS2IPaths(ucp.EnvVars, s2iPaths)

	if isToLocal {
		inputEnvs = uniqueAppendOrOverwriteEnvVars(
			inputEnvs,
			corev1.EnvVar{
				Name:  EnvS2ISrcBackupDir,
				Value: s2iPaths.SrcBackupPath,
			},
		)
	} else {
		inputEnvs = deleteEnvVars(inputEnvs, EnvS2ISrcBackupDir)
	}

	var dc appsv1.DeploymentConfig
	// if createPVC is true then we need to create a supervisorD volume and generate a new deployment config
	// needed for update from git to local/binary components
	// if false, we just update the current deployment config
	if createPVC {
		// Generate the SupervisorD Config
		dc = generateSupervisordDeploymentConfig(
			ucp.CommonObjectMeta,
			ucp.ImageMeta,
			inputEnvs,
			cmpContainer.EnvFrom,
			&ucp.ResourceLimits,
		)
		addInitVolumesToDC(&dc, ucp.CommonObjectMeta.Name, s2iPaths.DeploymentDir)

		ownerReference := GenerateOwnerReference(ucp.ExistingDC)

		// Setup PVC
		_, err = c.CreatePVC(getAppRootVolumeName(ucp.CommonObjectMeta.Name), "1Gi", ucp.CommonObjectMeta.Labels, ownerReference)
		if err != nil {
			return errors.Wrapf(err, "unable to create PVC for %s", ucp.CommonObjectMeta.Name)
		}
	} else {
		dc = updateSupervisorDeploymentConfig(
			SupervisorDUpdateParams{
				ucp.ExistingDC.DeepCopy(), ucp.CommonObjectMeta,
				ucp.ImageMeta,
				inputEnvs,
				cmpContainer.EnvFrom,
				&ucp.ResourceLimits,
			},
		)
	}

	// Patch the current DC with the new one
	err = c.PatchCurrentDC(
		dc,
		nil,
		existingCmpContainer,
		ucp,
		false,
	)
	if err != nil {
		return errors.Wrapf(err, "unable to update the current DeploymentConfig %s", ucp.CommonObjectMeta.Name)
	}

	return nil
}

func addInitVolumesToDC(dc *appsv1.DeploymentConfig, dcName string, deploymentDir string) {

	// Add the appropriate bootstrap volumes for SupervisorD
	addBootstrapVolumeCopyInitContainer(dc, dcName)
	addBootstrapSupervisordInitContainer(dc, dcName)
	addBootstrapVolume(dc, dcName)
	addBootstrapVolumeMount(dc, dcName)
	// only use the deployment Directory volume mount if its being used and
	// its not a sub directory of src_or_bin_path
	if deploymentDir != "" && !isSubDir(DefaultAppRootDir, deploymentDir) {
		addDeploymentDirVolumeMount(dc, deploymentDir)
	}
}

// removeTracesOfSupervisordFromDC takes a DeploymentConfig and removes any traces of the supervisord from it
// so it removes things like supervisord volumes, volumes mounts and init containers
func removeTracesOfSupervisordFromDC(dc *appsv1.DeploymentConfig) error {
	dcName := dc.Name

	err := removeVolumeFromDC(getAppRootVolumeName(dcName), dc)
	if err != nil {
		return err
	}

	err = removeVolumeMountsFromDC(getAppRootVolumeName(dcName), dc)
	if err != nil {
		return err
	}

	// remove the one bootstrapped init container
	for i, container := range dc.Spec.Template.Spec.InitContainers {
		if container.Name == "copy-files-to-volume" {
			dc.Spec.Template.Spec.InitContainers = append(dc.Spec.Template.Spec.InitContainers[:i], dc.Spec.Template.Spec.InitContainers[i+1:]...)
		}
	}

	return nil
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
	// Delete DeploymentConfig
	klog.V(3).Info("Deleting DeploymentConfigs")
	err := c.appsClient.DeploymentConfigs(c.Namespace).DeleteCollection(context.TODO(), metav1.DeleteOptions{PropagationPolicy: &deletionPolicy}, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		errorList = append(errorList, "unable to delete deploymentconfig")
	}
	// Delete BuildConfig
	klog.V(3).Info("Deleting BuildConfigs")
	err = c.buildClient.BuildConfigs(c.Namespace).DeleteCollection(context.TODO(), metav1.DeleteOptions{}, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		errorList = append(errorList, "unable to delete buildconfig")
	}
	// Delete ImageStream
	klog.V(3).Info("Deleting ImageStreams")
	err = c.imageClient.ImageStreams(c.Namespace).DeleteCollection(context.TODO(), metav1.DeleteOptions{}, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		errorList = append(errorList, "unable to delete imagestream")
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

	watcher, err := c.appsClient.DeploymentConfigs(c.Namespace).Watch(context.TODO(), metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return err
	}
	defer watcher.Stop()
	eventCh := watcher.ResultChan()

	for {
		select {
		case event, ok := <-eventCh:
			_, typeOk := event.Object.(*appsv1.DeploymentConfig)
			if !ok || !typeOk {
				return errors.New("Unable to watch deployment config")
			}
			if event.Type == watch.Deleted {
				klog.V(3).Infof("WaitForComponentDeletion, Event Received:Deleted")
				return nil
			} else if event.Type == watch.Error {
				klog.V(3).Infof("WaitForComponentDeletion, Event Received:Deleted ")
				return errors.New("Unable to watch deployment config")
			}
		case <-time.After(waitForComponentDeletionTimeout):
			klog.V(3).Infof("WaitForComponentDeletion, Timeout")
			return errors.New("Time out waiting for component to get deleted")
		}
	}
}

// LinkSecret links a secret to the DeploymentConfig of a component
func (c *Client) LinkSecret(secretName, componentName, applicationName string) error {

	var dcPatchProvider = func(dc *appsv1.DeploymentConfig) (string, error) {
		if len(dc.Spec.Template.Spec.Containers[0].EnvFrom) > 0 {
			// we always add the link as the first value in the envFrom array. That way we don't need to know the existing value
			return fmt.Sprintf(`[{ "op": "add", "path": "/spec/template/spec/containers/0/envFrom/0", "value": {"secretRef": {"name": "%s"}} }]`, secretName), nil
		}

		//in this case we need to add the full envFrom value
		return fmt.Sprintf(`[{ "op": "add", "path": "/spec/template/spec/containers/0/envFrom", "value": [{"secretRef": {"name": "%s"}}] }]`, secretName), nil
	}

	dcName, err := util.NamespaceOpenShiftObject(componentName, applicationName)
	if err != nil {
		return err
	}

	return c.patchDC(dcName, dcPatchProvider)
}

// UnlinkSecret unlinks a secret to the DeploymentConfig of a component
func (c *Client) UnlinkSecret(secretName, componentName, applicationName string) error {
	// Remove the Secret from the container
	var dcPatchProvider = func(dc *appsv1.DeploymentConfig) (string, error) {
		indexForRemoval := -1
		for i, env := range dc.Spec.Template.Spec.Containers[0].EnvFrom {
			if env.SecretRef.Name == secretName {
				indexForRemoval = i
				break
			}
		}

		if indexForRemoval == -1 {
			return "", fmt.Errorf("DeploymentConfig does not contain a link to %s", secretName)
		}

		return fmt.Sprintf(`[{"op": "remove", "path": "/spec/template/spec/containers/0/envFrom/%d"}]`, indexForRemoval), nil
	}

	dcName, err := util.NamespaceOpenShiftObject(componentName, applicationName)
	if err != nil {
		return err
	}

	return c.patchDC(dcName, dcPatchProvider)
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

func (c *Client) IsOpenshift4() bool {
	resource, err := c.GetKubeClient().IsResourceSupported("config.openshift.io", "v1", "clusterversions")
	if err != nil {
		return false
	}
	return resource
}

func (c *Client) GetKubeClient() *kclient.Client {
	return c.kubeClient
}

func (c *Client) SetKubeClient(client *kclient.Client) {
	c.kubeClient = client
}

// FindContainer finds the container
func FindContainer(containers []corev1.Container, name string) (corev1.Container, error) {

	if name == "" {
		return corev1.Container{}, errors.New("Invalid parameter for FindContainer, unable to find a blank container")
	}

	for _, container := range containers {
		if container.Name == name {
			return container, nil
		}
	}

	return corev1.Container{}, errors.New("Unable to find container")
}

// PropagateDeletes deletes the watch detected deleted files from remote component pod from each of the paths in passed s2iPaths
// Parameters:
//	targetPodName: Name of component pod
//	delSrcRelPaths: Paths to be deleted on the remote pod relative to component source base path ex: Component src: /abc/src, file deleted: abc/src/foo.lang => relative path: foo.lang
//	s2iPaths: Slice of all s2i paths -- deployment dir, destination dir, working dir, etc..
func (c *Client) PropagateDeletes(targetPodName string, delSrcRelPaths []string, s2iPaths []string) error {
	reader, writer := io.Pipe()
	var rmPaths []string
	if len(s2iPaths) == 0 || len(delSrcRelPaths) == 0 {
		return fmt.Errorf("Failed to propagate deletions: s2iPaths: %+v and delSrcRelPaths: %+v", s2iPaths, delSrcRelPaths)
	}
	for _, s2iPath := range s2iPaths {
		for _, delRelPath := range delSrcRelPaths {
			// since the paths inside the container are linux oriented
			// so we convert the paths accordingly
			rmPaths = append(rmPaths, filepath.ToSlash(filepath.Join(s2iPath, delRelPath)))
		}
	}
	klog.V(3).Infof("s2ipaths marked for deletion are %+v", rmPaths)
	cmdArr := []string{"rm", "-rf"}
	cmdArr = append(cmdArr, rmPaths...)

	err := c.GetKubeClient().ExecCMDInContainer("", targetPodName, cmdArr, writer, writer, reader, false)
	if err != nil {
		return err
	}
	return err
}

func injectS2IPaths(existingVars []corev1.EnvVar, s2iPaths S2IPaths) []corev1.EnvVar {
	return uniqueAppendOrOverwriteEnvVars(
		existingVars,
		corev1.EnvVar{
			Name:  EnvS2IScriptsURL,
			Value: s2iPaths.ScriptsPath,
		},
		corev1.EnvVar{
			Name:  EnvS2IScriptsProtocol,
			Value: s2iPaths.ScriptsPathProtocol,
		},
		corev1.EnvVar{
			Name:  EnvS2ISrcOrBinPath,
			Value: s2iPaths.SrcOrBinPath,
		},
		corev1.EnvVar{
			Name:  EnvS2IDeploymentDir,
			Value: s2iPaths.DeploymentDir,
		},
		corev1.EnvVar{
			Name:  EnvS2IWorkingDir,
			Value: s2iPaths.WorkingDir,
		},
		corev1.EnvVar{
			Name:  EnvS2IBuilderImageName,
			Value: s2iPaths.BuilderImgName,
		},
	)

}

func isSubDir(baseDir, otherDir string) bool {
	cleanedBaseDir := filepath.Clean(baseDir)
	cleanedOtherDir := filepath.Clean(otherDir)
	if cleanedBaseDir == cleanedOtherDir {
		return true
	}
	//matches, _ := filepath.Match(fmt.Sprintf("%s/*", cleanedBaseDir), cleanedOtherDir)
	matches, _ := filepath.Match(filepath.Join(cleanedBaseDir, "*"), cleanedOtherDir)
	return matches
}

// GenerateOwnerReference generates an ownerReference which can then be set as
// owner for various OpenShift objects and ensure that when the owner object is
// deleted from the cluster, all other objects are automatically removed by
// OpenShift garbage collector
func GenerateOwnerReference(dc *appsv1.DeploymentConfig) metav1.OwnerReference {

	ownerReference := metav1.OwnerReference{
		APIVersion: "apps.openshift.io/v1",
		Kind:       "DeploymentConfig",
		Name:       dc.Name,
		UID:        dc.UID,
	}

	return ownerReference
}
