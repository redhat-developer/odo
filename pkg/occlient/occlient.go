package occlient

import (
	taro "archive/tar"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/redhat-developer/odo/pkg/util"

	"github.com/fatih/color"
	"github.com/golang/glog"
	dockerapiv10 "github.com/openshift/api/image/docker10"
	"github.com/pkg/errors"

	servicecatalogclienset "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	appsschema "github.com/openshift/client-go/apps/clientset/versioned/scheme"
	appsclientset "github.com/openshift/client-go/apps/clientset/versioned/typed/apps/v1"
	buildschema "github.com/openshift/client-go/build/clientset/versioned/scheme"
	buildclientset "github.com/openshift/client-go/build/clientset/versioned/typed/build/v1"
	imageclientset "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"
	projectclientset "github.com/openshift/client-go/project/clientset/versioned/typed/project/v1"
	routeclientset "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	userclientset "github.com/openshift/client-go/user/clientset/versioned/typed/user/v1"

	scv1beta1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	appsv1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	imagev1 "github.com/openshift/api/image/v1"
	projectv1 "github.com/openshift/api/project/v1"
	routev1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/client-go/util/retry"
)

const (
	ocRequestTimeout   = 1 * time.Second
	OpenShiftNameSpace = "openshift"

	// The length of the string to be generated for names of resources
	nameLength = 5
	// git repository that will be used for bootstraping
	bootstrapperURI = "https://github.com/kadel/bootstrap-supervisored-s2i"
	bootstrapperRef = "v0.0.2"
)

// errorMsg is the message for user when invalid configuration error occurs
const errorMsg = `
Please login to your server: 

oc login https://mycluster.mydomain.com
`

type Client struct {
	kubeClient           kubernetes.Interface
	imageClient          imageclientset.ImageV1Interface
	appsClient           appsclientset.AppsV1Interface
	buildClient          buildclientset.BuildV1Interface
	projectClient        projectclientset.ProjectV1Interface
	serviceCatalogClient servicecatalogclienset.ServicecatalogV1beta1Interface
	routeClient          routeclientset.RouteV1Interface
	userClient           userclientset.UserV1Interface
	kubeConfig           clientcmd.ClientConfig
	namespace            string
}

func New(connectionCheck bool) (*Client, error) {
	var client Client

	// initialize client-go clients
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	client.kubeConfig = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	config, err := client.kubeConfig.ClientConfig()
	if err != nil {
		return nil, errors.New(err.Error() + errorMsg)
	}

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	client.kubeClient = kubeClient

	imageClient, err := imageclientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	client.imageClient = imageClient

	appsClient, err := appsclientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	client.appsClient = appsClient

	buildClient, err := buildclientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	client.buildClient = buildClient

	serviceCatalogClient, err := servicecatalogclienset.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	client.serviceCatalogClient = serviceCatalogClient

	projectClient, err := projectclientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	client.projectClient = projectClient

	routeClient, err := routeclientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	client.routeClient = routeClient

	userClient, err := userclientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	client.userClient = userClient

	namespace, _, err := client.kubeConfig.Namespace()
	if err != nil {
		return nil, err
	}
	client.namespace = namespace

	// Skip this if connectionCheck is false
	if !connectionCheck {
		if !isServerUp(config.Host) {
			return nil, errors.New("Unable to connect to OpenShift cluster, is it down?")
		}
		if !client.isLoggedIn() {
			return nil, errors.New("Please log in to the cluster")
		}
	}
	return &client, nil
}

// parseImageName parse image reference
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

// imageWithMetadata mutates the given image. It parses raw DockerImageManifest data stored in the image and
// fills its DockerImageMetadata and other fields.
// Copied from v3.7 github.com/openshift/origin/pkg/image/apis/image/v1/helpers.go
func imageWithMetadata(image *imagev1.Image) error {
	// Check if the metadata are already filled in for this image.
	meta, hasMetadata := image.DockerImageMetadata.Object.(*dockerapiv10.DockerImage)
	if hasMetadata && meta.Size > 0 {
		return nil
	}

	version := image.DockerImageMetadataVersion
	if len(version) == 0 {
		version = "1.0"
	}

	obj := &dockerapiv10.DockerImage{}
	if len(image.DockerImageMetadata.Raw) != 0 {
		if err := json.Unmarshal(image.DockerImageMetadata.Raw, obj); err != nil {
			return err
		}
		image.DockerImageMetadata.Object = obj
	}

	image.DockerImageMetadataVersion = version

	return nil
}

// isLoggedIn checks whether user is logged in or not and returns boolean output
func (c *Client) isLoggedIn() bool {
	// ~ indicates current user
	// Reference: https://github.com/openshift/origin/blob/master/pkg/oc/cli/cmd/whoami.go#L55
	output, err := c.userClient.Users().Get("~", metav1.GetOptions{})
	glog.V(4).Infof("isLoggedIn err:  %#v \n output: %#v", err, output.Name)
	if err != nil {
		glog.V(4).Info(errors.Wrap(err, "error running command"))
		glog.V(4).Infof("Output is: %v", output)
		return false
	}
	return true
}

// isServerUp returns true if server is up and running
func isServerUp(server string) bool {
	u, err := url.Parse(server)
	if err != nil {
		glog.V(4).Info(errors.Wrap(err, "unable to parse url"))
		return false
	}

	glog.V(4).Infof("Trying to connect to server %v", u.Host)
	_, connectionError := net.DialTimeout("tcp", u.Host, time.Duration(ocRequestTimeout))
	if connectionError != nil {
		glog.V(4).Info(errors.Wrap(connectionError, "unable to connect to server"))
		return false
	}
	glog.V(4).Infof("Server %v is up", server)
	return true
}

func (c *Client) GetCurrentProjectName() string {
	return c.namespace
}

// GetProjectNames return list of existing projects that user has access to.
func (c *Client) GetProjectNames() ([]string, error) {
	projects, err := c.projectClient.Projects().List(metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "unable to list projects")
	}

	var projectNames []string
	for _, p := range projects.Items {
		projectNames = append(projectNames, p.Name)
	}
	return projectNames, nil
}

func (c *Client) CreateNewProject(name string) error {
	projectRequest := &projectv1.ProjectRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	_, err := c.projectClient.ProjectRequests().Create(projectRequest)
	if err != nil {
		return errors.Wrapf(err, "unable to create new project %s", name)
	}
	return nil
}

func (c *Client) SetCurrentProject(project string) error {
	rawConfig, err := c.kubeConfig.RawConfig()
	if err != nil {
		return errors.Wrapf(err, "unable to switch to %s project", project)
	}

	rawConfig.Contexts[rawConfig.CurrentContext].Namespace = project

	err = clientcmd.ModifyConfig(clientcmd.NewDefaultClientConfigLoadingRules(), rawConfig, true)
	if err != nil {
		return errors.Wrapf(err, "unable to switch to %s project", project)
	}
	return nil
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

// getExposedPortsFromISI parse ImageStreamImage definition and return all exposed ports in form of ContainerPorts structs
func getExposedPortsFromISI(image *imagev1.ImageStreamImage) ([]corev1.ContainerPort, error) {
	// file DockerImageMetadata
	imageWithMetadata(&image.Image)

	var ports []corev1.ContainerPort

	for exposedPort := range image.Image.DockerImageMetadata.Object.(*dockerapiv10.DockerImage).ContainerConfig.ExposedPorts {
		splits := strings.Split(exposedPort, "/")
		if len(splits) != 2 {
			return nil, fmt.Errorf("invalid port %s", exposedPort)
		}

		portNumberI64, err := strconv.ParseInt(splits[0], 10, 32)
		if err != nil {
			return nil, errors.Wrapf(err, "invalid port number %s", splits[0])
		}
		portNumber := int32(portNumberI64)

		var portProto corev1.Protocol
		switch strings.ToUpper(splits[1]) {
		case "TCP":
			portProto = corev1.ProtocolTCP
		case "UDP":
			portProto = corev1.ProtocolUDP
		default:
			return nil, fmt.Errorf("invalid port protocol %s", splits[1])
		}

		port := corev1.ContainerPort{
			Name:          fmt.Sprintf("%d-%s", portNumber, strings.ToLower(string(portProto))),
			ContainerPort: portNumber,
			Protocol:      portProto,
		}

		ports = append(ports, port)
	}

	return ports, nil
}

// GetImageStreams returns the Image Stream objects in the given namespace
func (c *Client) GetImageStreams(namespace string) ([]imagev1.ImageStream, error) {
	imageStreamList, err := c.imageClient.ImageStreams(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "unable to list imagestreams")
	}
	return imageStreamList.Items, nil
}

// GetImageStreamsNames returns the names of the image streams in a given
// namespace
func (c *Client) GetImageStreamsNames(namespace string) ([]string, error) {
	imageStreams, err := c.GetImageStreams(namespace)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get image streams")
	}

	var names []string
	for _, imageStream := range imageStreams {
		names = append(names, imageStream.Name)
	}
	return names, nil
}

// isTagInImageStream takes a imagestream and a tag and checks if the tag is present in the imagestream's status attribute
func isTagInImageStream(is imagev1.ImageStream, imageTag string) bool {
	// Loop through the tags in the imagestream's status attribute
	for _, tag := range is.Status.Tags {
		// look for a matching tag
		if tag.Tag == imageTag {
			// Return true if found
			return true
		}
	}
	// Return false if not found.
	return false
}

// GetImageNS returns the imagestream using image details like imageNS, imageName and imageTag
// imageNS can be empty in which case, this function searches currentNamespace on priority. If
// imagestream of required tag not found in current namespace, then searches openshift namespace.
// If not found, error out. If imageNS is not empty string, then, the requested imageNS only is searched
// for requested imagestream
func (c *Client) GetImageStream(imageNS string, imageName string, imageTag string) (*imagev1.ImageStream, error) {
	var err error
	var imageStream *imagev1.ImageStream
	var currentProjectName string
	currentProjectName = c.GetCurrentProjectName()
	/*
		If User has not chosen image NS then,
			1. Use image from current NS if available
			2. If not 1, use default openshift NS
			3. If not 2, return errors from both 1 and 2
		else
			Use user chosen namespace
			If image doesn't exist in user chosen namespace,
				error out
			else
				Proceed
	*/
	// User has not passed any particular ImageStream
	if imageNS == "" {

		// First try finding imagestream from current namespace
		currentNSImageStream, e := c.imageClient.ImageStreams(currentProjectName).Get(imageName, metav1.GetOptions{})
		if e != nil {
			err = errors.Wrapf(e, "no match found for : %s in namespace %s", imageName, currentProjectName)
		} else {
			if isTagInImageStream(*currentNSImageStream, imageTag) {
				return currentNSImageStream, nil
			}
		}

		// If not in current namespace, try finding imagestream from openshift namespace
		openshiftNSImageStream, e := c.imageClient.ImageStreams(OpenShiftNameSpace).Get(imageName, metav1.GetOptions{})
		if e != nil {
			// The image is not available in current Namespace.
			err = errors.Wrapf(e, "%s\n.no match found for : %s in namespace %s", err.Error(), imageName, OpenShiftNameSpace)
		} else {
			if isTagInImageStream(*openshiftNSImageStream, imageTag) {
				return openshiftNSImageStream, nil
			}
		}
		if e != nil && err != nil {
			// Imagestream not found in openshift and current namespaces
			return nil, err
		}

		// Required tag not in openshift and current namespaces
		return nil, fmt.Errorf("image stream %s with tag %s not found in openshift and %s namespaces", imageName, imageTag, currentProjectName)

	} else {

		// Fetch imagestream from requested namespace
		imageStream, err = c.imageClient.ImageStreams(imageNS).Get(imageName, metav1.GetOptions{})
		if err != nil {
			return nil, errors.Wrapf(
				err, "no match found for %s in namespace %s", imageName, imageNS,
			)
		}
		if !isTagInImageStream(*imageStream, imageTag) {
			return nil, fmt.Errorf("image stream %s with tag %s not found in %s namespaces", imageName, imageTag, currentProjectName)
		}
	}

	return imageStream, nil
}

// GetExposedPorts retruns image namespace and list of ContainerPorts that are exposed by given image
func (c *Client) GetExposedPorts(imageStream *imagev1.ImageStream, imageTag string) ([]corev1.ContainerPort, error) {
	var containerPorts []corev1.ContainerPort

	glog.V(4).Infof("Checking for exact match of builderImage with ImageStream")
	imageNS := imageStream.ObjectMeta.Namespace
	imageName := imageStream.ObjectMeta.Name

	tagFound := false

	for _, tag := range imageStream.Status.Tags {
		// look for matching tag
		if tag.Tag == imageTag {
			tagFound = true
			glog.V(4).Infof("Found exact image tag match for %s:%s", imageName, imageTag)

			if len(tag.Items) > 0 {
				tagDigest := tag.Items[0].Image
				imageStreamImageName := fmt.Sprintf("%s@%s", imageName, tagDigest)

				// look for imageStreamImage for given tag (reference by digest)
				imageStreamImage, err := c.imageClient.ImageStreamImages(imageNS).Get(imageStreamImageName, metav1.GetOptions{})
				if err != nil {
					return nil, errors.Wrapf(err, "unable to find ImageStreamImage with  %s digest", imageStreamImageName)
				}

				// get ports that are exported by image
				containerPorts, err = getExposedPortsFromISI(imageStreamImage)
				if err != nil {
					return nil, errors.Wrapf(err, "unable to get exported ports from %s:%s image", imageName, imageTag)
				}
			} else {
				return nil, fmt.Errorf("unable to find tag %s for image %s", imageTag, imageName)
			}
		}
	}

	if !tagFound {
		return nil, fmt.Errorf("unable to find tag %s for image %s", imageTag, imageName)
	}

	return containerPorts, nil
}

func getAppRootVolumeName(dcName string) string {
	return fmt.Sprintf("%s-s2idata", dcName)
}

// NewAppS2I create new application using S2I
// gitUrl is the url of the git repo
// inputPorts is the array containing the string port values
func (c *Client) NewAppS2I(name string, builderImage string, gitUrl string, labels map[string]string, annotations map[string]string, inputPorts []string) error {

	imageNS, imageName, imageTag, _, err := ParseImageName(builderImage)
	if err != nil {
		return errors.Wrap(err, "unable to parse image name")
	}
	imageStream, err := c.GetImageStream(imageNS, imageName, imageTag)
	if err != nil {
		return errors.Wrap(err, "Unable to create new app")
	}
	/*
	 Set imageNS to the namespace of above fetched imagestream because, the namespace passed here can potentially be emptystring
	 in which case, GetImageStream function resolves to correct namespace in accordance with priorities in GetImageStream
	*/
	imageNS = imageStream.ObjectMeta.Namespace

	var containerPorts []corev1.ContainerPort
	if len(inputPorts) == 0 {
		containerPorts, err = c.GetExposedPorts(imageStream, imageTag)
		if err != nil {
			return errors.Wrapf(err, "unable to get exposed ports for %s:%s", imageName, imageTag)
		}
	} else {
		if err != nil {
			return errors.Wrapf(err, "unable to create s2i app for %s", name)
		}
		imageNS = imageStream.ObjectMeta.Namespace
		containerPorts, err = getContainerPortsFromStrings(inputPorts)
		if err != nil {
			return errors.Wrapf(err, "unable to get container ports from %v", inputPorts)
		}
	}

	// ObjectMetadata are the same for all generated objects
	commonObjectMeta := metav1.ObjectMeta{
		Name:        name,
		Labels:      labels,
		Annotations: annotations,
	}

	// generate and create ImageStream
	is := imagev1.ImageStream{
		ObjectMeta: commonObjectMeta,
	}
	_, err = c.imageClient.ImageStreams(c.namespace).Create(&is)
	if err != nil {
		return errors.Wrapf(err, "unable to create ImageStream for %s", name)
	}

	// if gitUrl is not set, error out
	if gitUrl == "" {
		return errors.New("unable to create buildSource with empty gitUrl")
	}

	buildSource := buildv1.BuildSource{
		Git: &buildv1.GitBuildSource{
			URI: gitUrl,
		},
		Type: buildv1.BuildSourceGit,
	}

	bc := buildv1.BuildConfig{
		ObjectMeta: commonObjectMeta,
		Spec: buildv1.BuildConfigSpec{
			CommonSpec: buildv1.CommonSpec{
				Output: buildv1.BuildOutput{
					To: &corev1.ObjectReference{
						Kind: "ImageStreamTag",
						Name: name + ":latest",
					},
				},
				Source: buildSource,
				Strategy: buildv1.BuildStrategy{
					SourceStrategy: &buildv1.SourceBuildStrategy{
						From: corev1.ObjectReference{
							Kind:      "ImageStreamTag",
							Name:      imageName + ":" + imageTag,
							Namespace: imageNS,
						},
					},
				},
			},
		},
	}
	_, err = c.buildClient.BuildConfigs(c.namespace).Create(&bc)
	if err != nil {
		return errors.Wrapf(err, "unable to create BuildConfig for %s", name)
	}

	// generate  and create DeploymentConfig
	dc := appsv1.DeploymentConfig{
		ObjectMeta: commonObjectMeta,
		Spec: appsv1.DeploymentConfigSpec{
			Replicas: 1,
			Selector: map[string]string{
				"deploymentconfig": name,
			},
			Template: &corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"deploymentconfig": name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image: bc.Spec.Output.To.Name,
							Name:  name,
							Ports: containerPorts,
						},
					},
				},
			},
			Triggers: []appsv1.DeploymentTriggerPolicy{
				{
					Type: "ConfigChange",
				},
				{
					Type: "ImageChange",
					ImageChangeParams: &appsv1.DeploymentTriggerImageChangeParams{
						Automatic: true,
						ContainerNames: []string{
							name,
						},
						From: corev1.ObjectReference{
							Kind: "ImageStreamTag",
							Name: bc.Spec.Output.To.Name,
						},
					},
				},
			},
		},
	}
	_, err = c.appsClient.DeploymentConfigs(c.namespace).Create(&dc)
	if err != nil {
		return errors.Wrapf(err, "unable to create DeploymentConfig for %s", name)
	}

	err = c.CreateService(commonObjectMeta, dc.Spec.Template.Spec.Containers[0].Ports)
	if err != nil {
		return errors.Wrapf(err, "unable to create Service for %s", name)
	}

	return nil

}

// BootstrapSupervisoredS2I uses s2i to inject Supervisor into builder image.
// Supervisor keeps pod running (runs as pid1), so you it is possible to trigger assembly script inside running pod,
// and than restart application using Supervisor without need to restart whole container.
// inputPorts is the array containing the string port values
func (c *Client) BootstrapSupervisoredS2I(name string, builderImage string, labels map[string]string, annotations map[string]string, inputPorts []string) error {
	imageNS, imageName, imageTag, _, err := ParseImageName(builderImage)

	if err != nil {
		return errors.Wrap(err, "unable to create new s2i git build ")
	}
	imageStream, err := c.GetImageStream(imageNS, imageName, imageTag)
	if err != nil {
		return errors.Wrap(err, "Failed to bootstrap supervisored")
	}
	/*
	 Set imageNS to the namespace of above fetched imagestream because, the namespace passed here can potentially be emptystring
	 in which case, GetImageStream function resolves to correct namespace in accordance with priorities in GetImageStream
	*/
	imageNS = imageStream.ObjectMeta.Namespace

	var containerPorts []corev1.ContainerPort
	if len(inputPorts) == 0 {
		containerPorts, err = c.GetExposedPorts(imageStream, imageTag)
		if err != nil {
			return errors.Wrapf(err, "unable to get exposed ports for %s:%s", imageName, imageTag)
		}
	} else {
		if err != nil {
			return errors.Wrapf(err, "unable to bootstrap s2i supervisored for %s", name)
		}
		containerPorts, err = getContainerPortsFromStrings(inputPorts)
		if err != nil {
			return errors.Wrapf(err, "unable to get container ports from %v", inputPorts)
		}
	}

	// ObjectMetadata are the same for all generated objects
	commonObjectMeta := metav1.ObjectMeta{
		Name:        name,
		Labels:      labels,
		Annotations: annotations,
	}

	// generate and create ImageStream
	is := imagev1.ImageStream{
		ObjectMeta: commonObjectMeta,
	}
	_, err = c.imageClient.ImageStreams(c.namespace).Create(&is)
	if err != nil {
		return errors.Wrapf(err, "unable to create ImageStream for %s", name)
	}

	// generate BuildConfig
	buildSource := buildv1.BuildSource{
		Git: &buildv1.GitBuildSource{
			URI: bootstrapperURI,
			Ref: bootstrapperRef,
		},
		Type: buildv1.BuildSourceGit,
	}

	bc := buildv1.BuildConfig{
		ObjectMeta: commonObjectMeta,
		Spec: buildv1.BuildConfigSpec{
			CommonSpec: buildv1.CommonSpec{
				Output: buildv1.BuildOutput{
					To: &corev1.ObjectReference{
						Kind: "ImageStreamTag",
						Name: name + ":latest",
					},
				},
				Source: buildSource,
				Strategy: buildv1.BuildStrategy{
					SourceStrategy: &buildv1.SourceBuildStrategy{
						From: corev1.ObjectReference{
							Kind:      "ImageStreamTag",
							Name:      imageName + ":" + imageTag,
							Namespace: imageNS,
						},
					},
				},
			},
		},
	}
	_, err = c.buildClient.BuildConfigs(c.namespace).Create(&bc)
	if err != nil {
		return errors.Wrapf(err, "unable to create BuildConfig for %s", name)
	}

	// generate  and create DeploymentConfig
	dc := appsv1.DeploymentConfig{
		ObjectMeta: commonObjectMeta,
		Spec: appsv1.DeploymentConfigSpec{
			Replicas: 1,
			Selector: map[string]string{
				"deploymentconfig": name,
			},
			Template: &corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"deploymentconfig": name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image: bc.Spec.Output.To.Name,
							Name:  name,
							Ports: containerPorts,
						},
					},
				},
			},
			Triggers: []appsv1.DeploymentTriggerPolicy{
				{
					Type: "ConfigChange",
				},
				{
					Type: "ImageChange",
					ImageChangeParams: &appsv1.DeploymentTriggerImageChangeParams{
						Automatic: true,
						ContainerNames: []string{
							name,
							"copy-files-to-volume",
						},
						From: corev1.ObjectReference{
							Kind: "ImageStreamTag",
							Name: bc.Spec.Output.To.Name,
						},
					},
				},
			},
		},
	}

	addBootstrapInitContainer(&dc, name)
	addBootstrapVolume(&dc, name)
	addBootstrapVolumeMount(&dc, name)

	_, err = c.appsClient.DeploymentConfigs(c.namespace).Create(&dc)
	if err != nil {
		return errors.Wrapf(err, "unable to create DeploymentConfig for %s", name)
	}

	err = c.CreateService(commonObjectMeta, dc.Spec.Template.Spec.Containers[0].Ports)
	if err != nil {
		return errors.Wrapf(err, "unable to create Service for %s", name)
	}

	_, err = c.CreatePVC(getAppRootVolumeName(name), "1Gi", labels)
	if err != nil {
		return errors.Wrapf(err, "unable to create PVC for %s", name)
	}

	return nil
}

// CreateService generates and creates the service
// commonObjectMeta is the ObjectMeta for the service
// dc is the deploymentConfig to get the container ports
func (c *Client) CreateService(commonObjectMeta metav1.ObjectMeta, containerPorts []corev1.ContainerPort) error {
	// generate and create Service
	var svcPorts []corev1.ServicePort
	for _, containerPort := range containerPorts {
		svcPort := corev1.ServicePort{

			Name:       containerPort.Name,
			Port:       containerPort.ContainerPort,
			Protocol:   containerPort.Protocol,
			TargetPort: intstr.FromInt(int(containerPort.ContainerPort)),
		}
		svcPorts = append(svcPorts, svcPort)
	}
	svc := corev1.Service{
		ObjectMeta: commonObjectMeta,
		Spec: corev1.ServiceSpec{
			Ports: svcPorts,
			Selector: map[string]string{
				"deploymentconfig": commonObjectMeta.Name,
			},
		},
	}
	_, err := c.kubeClient.CoreV1().Services(c.namespace).Create(&svc)
	if err != nil {
		return errors.Wrapf(err, "unable to create Service for %s", commonObjectMeta.Name)
	}
	return nil
}

// AddBootstrapInitContainer adds the bootstrap init container to the deployment config
// dc is the deployment config to be updated
// dcName is the name of the deployment config
func addBootstrapInitContainer(dc *appsv1.DeploymentConfig, dcName string) {
	dc.Spec.Template.Spec.InitContainers = append(dc.Spec.Template.Spec.InitContainers,
		corev1.Container{
			Name:  "copy-files-to-volume",
			Image: dc.Spec.Template.Spec.Containers[0].Image,
			Command: []string{
				"copy-files-to-volume",
				"/opt/app-root",
				"/mnt/app-root"},
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      getAppRootVolumeName(dcName),
					MountPath: "/mnt",
				},
			},
		})
}

// addBootstrapVolume adds the bootstrap volume to the deployment config
// dc is the deployment config to be updated
// dcName is the name of the deployment config
func addBootstrapVolume(dc *appsv1.DeploymentConfig, dcName string) {
	dc.Spec.Template.Spec.Volumes = append(dc.Spec.Template.Spec.Volumes, corev1.Volume{
		Name: getAppRootVolumeName(dcName),
		VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: getAppRootVolumeName(dcName),
			},
		},
	})
}

// addBootstrapVolumeMount mounts the bootstrap volume to the deployment config
// dc is the deployment config to be updated
// dcName is the name of the deployment config
func addBootstrapVolumeMount(dc *appsv1.DeploymentConfig, dcName string) {
	for i := range dc.Spec.Template.Spec.Containers {
		dc.Spec.Template.Spec.Containers[i].VolumeMounts = append(dc.Spec.Template.Spec.Containers[i].VolumeMounts, corev1.VolumeMount{
			Name:      getAppRootVolumeName(dcName),
			MountPath: "/opt/app-root",
			SubPath:   "app-root",
		})
	}
}

// UpdateBuildConfig updates the BuildConfig file
// buildConfigName is the name of the BuildConfig file to be updated
// projectName is the name of the project
// gitUrl equals to the git URL of the source and is equals to "" if the source is of type dir or binary
// annotations contains the annotations for the BuildConfig file
func (c *Client) UpdateBuildConfig(buildConfigName string, projectName string, gitUrl string, annotations map[string]string) error {
	// generate BuildConfig
	buildSource := buildv1.BuildSource{}

	// if gitUrl set change buildSource to git and use given repo
	if gitUrl != "" {
		buildSource = buildv1.BuildSource{
			Git: &buildv1.GitBuildSource{
				URI: gitUrl,
			},
			Type: buildv1.BuildSourceGit,
		}
	} else {
		buildSource = buildv1.BuildSource{
			Git: &buildv1.GitBuildSource{
				URI: bootstrapperURI,
				Ref: bootstrapperRef,
			},
			Type: buildv1.BuildSourceGit,
		}
	}
	buildConfig, err := c.GetBuildConfigFromName(buildConfigName, projectName)
	if err != nil {
		return errors.Wrap(err, "unable to get the BuildConfig file")
	}
	buildConfig.Spec.Source = buildSource
	buildConfig.Annotations = annotations
	_, err = c.buildClient.BuildConfigs(c.namespace).Update(buildConfig)
	if err != nil {
		return errors.Wrap(err, "unable to update the component")
	}
	return nil
}

// UpdateDCAnnotations updates the DeploymentConfig file
// dcName is the name of the DeploymentConfig file to be updated
// annotations contains the annotations for the DeploymentConfig file
func (c *Client) UpdateDCAnnotations(dcName string, annotations map[string]string) error {
	dc, err := c.GetDeploymentConfigFromName(dcName, c.namespace)
	if err != nil {
		return errors.Wrapf(err, "unable to get DeploymentConfig %s", dcName)
	}

	dc.Annotations = annotations
	_, err = c.appsClient.DeploymentConfigs(c.namespace).Update(dc)
	if err != nil {
		return errors.Wrapf(err, "unable to uDeploymentConfig config %s", dcName)
	}
	return nil
}

// SetupForSupervisor adds the supervisor to the deployment config
// dcName is the name of the deployment config to be updated
// projectName is the name of the project
// annotations are the updated annotations for the new deployment config
// labels are the labels of the PVC created while setting up the supervisor
func (c *Client) SetupForSupervisor(dcName string, projectName string, annotations map[string]string, labels map[string]string) error {
	dc, err := c.GetDeploymentConfigFromName(dcName, projectName)
	if err != nil {
		return errors.Wrapf(err, "unable to get DeploymentConfig %s", dcName)
	}

	dc.Annotations = annotations

	addBootstrapInitContainer(dc, dcName)

	addBootstrapVolume(dc, dcName)

	addBootstrapVolumeMount(dc, dcName)

	dc, err = c.appsClient.DeploymentConfigs(c.namespace).Update(dc)
	if err != nil {
		return errors.Wrapf(err, "unable to uDeploymentConfig config %s", dcName)
	}
	_, err = c.CreatePVC(getAppRootVolumeName(dcName), "1Gi", labels)
	if err != nil {
		return errors.Wrapf(err, "unable to create PVC for %s", dcName)
	}
	return nil
}

// CleanupAfterSupervisor removes the supervisor from the deployment config
// dcName is the name of the deployment config to be updated
// projectName is the name of the project
// annotations are the updated annotations for the new deployment config
func (c *Client) CleanupAfterSupervisor(dcName string, projectName string, annotations map[string]string) error {
	dc, err := c.GetDeploymentConfigFromName(dcName, projectName)
	if err != nil {
		return errors.Wrapf(err, "unable to get DeploymentConfig %s ", dcName)
	}

	dc.Annotations = annotations

	found := removeVolumeFromDC(getAppRootVolumeName(dcName), dc)
	if !found {
		return errors.Wrapf(err, "unable to find volume in the dc")
	}
	found = removeVolumeMountFromDC(getAppRootVolumeName(dcName), dc)
	if !found {
		return errors.Wrapf(err, "unable to find volume in the dc")
	}

	// remove the one bootstrapped init container
	for i, container := range dc.Spec.Template.Spec.InitContainers {
		if container.Name == "copy-files-to-volume" {
			dc.Spec.Template.Spec.InitContainers = append(dc.Spec.Template.Spec.InitContainers[:i], dc.Spec.Template.Spec.InitContainers[i+1:]...)
		}
	}

	dc, err = c.appsClient.DeploymentConfigs(c.namespace).Update(dc)
	if err != nil {
		return errors.Wrapf(err, "unable to update deployment config %s", dcName)
	}

	err = c.DeletePVC(getAppRootVolumeName(dcName))
	if err != nil {
		return errors.Wrapf(err, "unable to delete S2I data PVC from %s", dcName)
	}
	return nil
}

// GetLatestBuildName gets the name of the latest build
// buildConfigName is the name of the buildConfig for which we are fetching the build name
// returns the name of the latest build or the error
func (c *Client) GetLatestBuildName(buildConfigName string) (string, error) {
	buildConfig, err := c.buildClient.BuildConfigs(c.namespace).Get(buildConfigName, metav1.GetOptions{})
	if err != nil {
		return "", errors.Wrap(err, "unable to get the latest build name")
	}
	return fmt.Sprintf("%s-%d", buildConfigName, buildConfig.Status.LastVersion), nil
}

// StartBuild starts new build as it is, returns name of the build stat was started
func (c *Client) StartBuild(name string) (string, error) {
	buildRequest := buildv1.BuildRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	result, err := c.buildClient.BuildConfigs(c.namespace).Instantiate(name, &buildRequest)
	if err != nil {
		return "", errors.Wrapf(err, "unable to instantiate BuildConfig for %s", name)
	}
	glog.V(4).Infof("Build %s for BuildConfig %s triggered.", name, result.Name)

	return result.Name, nil
}

// WaitForBuildToFinish block and waits for build to finish. Returns error if build failed or was canceled.
func (c *Client) WaitForBuildToFinish(buildName string) error {
	glog.V(4).Infof("Waiting for %s  build to finish", buildName)

	w, err := c.buildClient.Builds(c.namespace).Watch(metav1.ListOptions{
		FieldSelector: fields.Set{"metadata.name": buildName}.AsSelector().String(),
	})
	if err != nil {
		return errors.Wrapf(err, "unable to watch build")
	}
	defer w.Stop()
	for {
		val, ok := <-w.ResultChan()
		if !ok {
			break
		}
		if e, ok := val.Object.(*buildv1.Build); ok {
			glog.V(4).Infof("Status of %s build is %s", e.Name, e.Status.Phase)
			switch e.Status.Phase {
			case buildv1.BuildPhaseComplete:
				glog.V(4).Infof("Build %s completed.", e.Name)
				return nil
			case buildv1.BuildPhaseFailed, buildv1.BuildPhaseCancelled, buildv1.BuildPhaseError:
				return errors.Errorf("build %s status %s", e.Name, e.Status.Phase)
			}
		}
	}
	return nil
}

// WaitAndGetPod block and waits until pod matching selector is in in Running state
func (c *Client) WaitAndGetPod(selector string) (*corev1.Pod, error) {
	glog.V(4).Infof("Waiting for %s pod", selector)

	w, err := c.kubeClient.CoreV1().Pods(c.namespace).Watch(metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to watch pod")
	}
	defer w.Stop()
	for {
		val, ok := <-w.ResultChan()
		if !ok {
			break
		}
		if e, ok := val.Object.(*corev1.Pod); ok {
			glog.V(4).Infof("Status of %s pod is %s", e.Name, e.Status.Phase)
			switch e.Status.Phase {
			case corev1.PodRunning:
				glog.V(4).Infof("Pod %s is running.", e.Name)
				return e, nil
			case corev1.PodFailed, corev1.PodUnknown:
				return nil, errors.Errorf("pod %s status %s", e.Name, e.Status.Phase)
			}
		}
	}
	return nil, errors.Errorf("unknown error while waiting for pod matchin '%s' selector", selector)
}

// FollowBuildLog stream build log to stdout
func (c *Client) FollowBuildLog(buildName string, stdout io.Writer) error {
	buildLogOptions := buildv1.BuildLogOptions{
		Follow: true,
		NoWait: false,
	}

	rd, err := c.buildClient.RESTClient().Get().
		Namespace(c.namespace).
		Resource("builds").
		Name(buildName).
		SubResource("log").
		VersionedParams(&buildLogOptions, buildschema.ParameterCodec).
		Stream()

	if err != nil {
		return errors.Wrapf(err, "unable get build log %s", buildName)
	}
	defer rd.Close()

	// Set the colour of the stdout output..
	color.Set(color.FgYellow)
	defer color.Unset()

	if _, err = io.Copy(stdout, rd); err != nil {
		return errors.Wrapf(err, "error streaming logs for %s", buildName)
	}

	return nil
}

// Display DeploymentConfig log to stdout
func (c *Client) DisplayDeploymentConfigLog(deploymentConfigName string, followLog bool, stdout io.Writer) error {

	// Set standard log options
	deploymentLogOptions := appsv1.DeploymentLogOptions{Follow: false, NoWait: true}

	// If the log is being followed, set it to follow / don't wait
	if followLog {
		// TODO: https://github.com/kubernetes/kubernetes/pull/60696
		// Unable to set to 0, until openshift/client-go updates their Kubernetes vendoring to 1.11.0
		// Set to 1 for now.
		tailLines := int64(1)
		deploymentLogOptions = appsv1.DeploymentLogOptions{Follow: true, NoWait: false, Previous: false, TailLines: &tailLines}
	}

	// RESTClient call to OpenShift
	rd, err := c.appsClient.RESTClient().Get().
		Namespace(c.namespace).
		Name(deploymentConfigName).
		Resource("deploymentconfigs").
		SubResource("log").
		VersionedParams(&deploymentLogOptions, appsschema.ParameterCodec).
		Stream()
	if err != nil {
		return errors.Wrapf(err, "unable get deploymentconfigs log %s", deploymentConfigName)
	}
	if rd == nil {
		return errors.New("unable to retrieve DeploymentConfig from OpenShift, does your component exist?")
	}
	defer rd.Close()

	// Copy to stdout (in yellow)
	color.Set(color.FgYellow)
	defer color.Unset()

	// If we are going to followLog, we'll be copying it to stdout
	// else, we copy it to a buffer
	if followLog {

		if _, err = io.Copy(stdout, rd); err != nil {
			return errors.Wrapf(err, "error followLoging logs for %s", deploymentConfigName)
		}

	} else {

		// Copy to buffer (we aren't going to be followLoging the logs..)
		buf := new(bytes.Buffer)
		_, err = io.Copy(buf, rd)
		if err != nil {
			return errors.Wrapf(err, "unable to copy followLog to buffer")
		}

		// Copy to stdout
		if _, err = io.Copy(stdout, buf); err != nil {
			return errors.Wrapf(err, "error copying logs to stdout")
		}

	}

	return nil
}

// Delete takes labels as a input and based on it, deletes respective resource
func (c *Client) Delete(labels map[string]string) error {
	// convert labels to selector
	selector := util.ConvertLabelsToSelector(labels)
	glog.V(4).Infof("Selectors used for deletion: %s", selector)

	var errorList []string
	// Delete DeploymentConfig
	glog.V(4).Info("Deleting DeploymentConfigs")
	err := c.appsClient.DeploymentConfigs(c.namespace).DeleteCollection(&metav1.DeleteOptions{}, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		errorList = append(errorList, "unable to delete deploymentconfig")
	}
	// Delete Route
	glog.V(4).Info("Deleting Routes")
	err = c.routeClient.Routes(c.namespace).DeleteCollection(&metav1.DeleteOptions{}, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		errorList = append(errorList, "unable to delete route")
	}
	// Delete BuildConfig
	glog.V(4).Info("Deleting BuildConfigs")
	err = c.buildClient.BuildConfigs(c.namespace).DeleteCollection(&metav1.DeleteOptions{}, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		errorList = append(errorList, "unable to delete buildconfig")
	}
	// Delete ImageStream
	glog.V(4).Info("Deleting ImageStreams")
	err = c.imageClient.ImageStreams(c.namespace).DeleteCollection(&metav1.DeleteOptions{}, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		errorList = append(errorList, "unable to delete imagestream")
	}
	// Delete Services
	glog.V(4).Info("Deleting Services")
	svcList, err := c.kubeClient.CoreV1().Services(c.namespace).List(metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		errorList = append(errorList, "unable to list services")
	}
	for _, svc := range svcList.Items {
		err = c.kubeClient.CoreV1().Services(c.namespace).Delete(svc.Name, &metav1.DeleteOptions{})
		if err != nil {
			errorList = append(errorList, "unable to delete service")
		}
	}
	// PersistentVolumeClaim
	glog.V(4).Infof("Deleting PersistentVolumeClaims")
	err = c.kubeClient.CoreV1().PersistentVolumeClaims(c.namespace).DeleteCollection(&metav1.DeleteOptions{}, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		errorList = append(errorList, "unable to delete volume")
	}

	// Error string
	errString := strings.Join(errorList, ",")
	if len(errString) != 0 {
		return errors.New(errString)
	}
	return nil

}

// DeleteServiceInstance takes labels as a input and based on it, deletes respective service instance
func (c *Client) DeleteServiceInstance(labels map[string]string) error {
	glog.V(4).Infof("Deleting Service Instance")

	// convert labels to selector
	selector := util.ConvertLabelsToSelector(labels)
	glog.V(4).Infof("Selectors used for deletion: %s", selector)

	// Listing out serviceInstance because `DeleteCollection` method don't work on serviceInstance
	svcCatList, err := c.GetServiceInstanceList(c.namespace, selector)
	if err != nil {
		return errors.Wrap(err, "unable to list service instance")
	}

	// Iterating over serviceInstance List and deleting one by one
	for _, svc := range svcCatList {
		err = c.serviceCatalogClient.ServiceInstances(c.namespace).Delete(svc.Name, &metav1.DeleteOptions{})
		if err != nil {
			return errors.Wrap(err, "unable to delete serviceInstance")
		}
	}

	return nil
}

// DeleteProject deletes given project
func (c *Client) DeleteProject(name string) error {
	err := c.projectClient.Projects().Delete(name, &metav1.DeleteOptions{})
	if err != nil {
		return errors.Wrap(err, "unable to delete project")
	}
	return nil
}

// GetLabelValues get label values of given label from objects in project that are matching selector
// returns slice of unique label values
func (c *Client) GetLabelValues(project string, label string, selector string) ([]string, error) {
	// List DeploymentConfig according to selectors
	dcList, err := c.appsClient.DeploymentConfigs(project).List(metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return nil, errors.Wrap(err, "unable to list DeploymentConfigs")
	}
	var values []string
	for _, elem := range dcList.Items {
		for key, val := range elem.Labels {
			if key == label {
				values = append(values, val)
			}
		}
	}

	return values, nil
}

// GetServiceInstanceList returns list service instances
func (c *Client) GetServiceInstanceList(namespace string, selector string) ([]scv1beta1.ServiceInstance, error) {
	// List ServiceInstance according to given selectors
	svcList, err := c.serviceCatalogClient.ServiceInstances(namespace).List(metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return nil, errors.Wrap(err, "unable to list ServiceInstances")
	}

	return svcList.Items, nil
}

// GetBuildConfigFromName get BuildConfig by its name
func (c *Client) GetBuildConfigFromName(name string, project string) (*buildv1.BuildConfig, error) {
	glog.V(4).Infof("Getting BuildConfig: %s", name)
	bc, err := c.buildClient.BuildConfigs(project).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get BuildConfig %s", name)
	}
	return bc, nil
}

// GetClusterServiceClasses queries the service service catalog to get available clusterServiceClasses
func (c *Client) GetClusterServiceClasses() ([]scv1beta1.ClusterServiceClass, error) {
	classList, err := c.serviceCatalogClient.ClusterServiceClasses().List(metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "unable to list cluster service classes")
	}
	return classList.Items, nil
}

// CreateServiceInstance creates service instance from service catalog
func (c *Client) CreateServiceInstance(componentName string, componentType string, labels map[string]string) error {
	// Creating Service Instance
	_, err := c.serviceCatalogClient.ServiceInstances(c.namespace).Create(
		&scv1beta1.ServiceInstance{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ServiceInstance",
				APIVersion: "servicecatalog.k8s.io/v1beta1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Finalizers: []string{"kubernetes-incubator/service-catalog"},
				Name:       componentName,
				Namespace:  c.namespace,
				Labels:     labels,
			},
			Spec: scv1beta1.ServiceInstanceSpec{
				PlanReference: scv1beta1.PlanReference{
					ClusterServiceClassExternalName: componentType,
				},
			},
			Status: scv1beta1.ServiceInstanceStatus{},
		})

	if err != nil {
		return errors.Wrap(err, "unable to create service instance")
	}
	return nil
}

// GetClusterServiceClassExternalNames returns the names of all the cluster service
// classes in the cluster
func (c *Client) GetClusterServiceClassExternalNames() ([]string, error) {
	var classNames []string

	classes, err := c.GetClusterServiceClasses()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get cluster service classes")
	}

	for _, class := range classes {
		classNames = append(classNames, class.Spec.ExternalName)
	}
	return classNames, nil
}

// imageStreamExists returns true if the given image stream exists in the given
// namespace
func (c *Client) imageStreamExists(name string, namespace string) bool {
	imageStreams, err := c.GetImageStreamsNames(namespace)
	if err != nil {
		glog.V(4).Infof("unable to get image streams in the namespace: %v", namespace)
		return false
	}

	for _, is := range imageStreams {
		if is == name {
			return true
		}
	}
	return false
}

// clusterServiceClassExists returns true if the given external name of the
// cluster service class exists in the cluster, and false otherwise
func (c *Client) clusterServiceClassExists(name string) bool {
	clusterServiceClasses, err := c.GetClusterServiceClassExternalNames()
	if err != nil {
		glog.V(4).Infof("unable to get cluster service classes' external names")
	}

	for _, class := range clusterServiceClasses {
		if class == name {
			return true
		}
	}

	return false
}

// CreateRoute creates a route object for the given service and with the given labels
// serviceName is the name of the service for the target reference
// portNumber is the target port of the route
func (c *Client) CreateRoute(name string, serviceName string, portNumber intstr.IntOrString, labels map[string]string) (*routev1.Route, error) {
	route := &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Spec: routev1.RouteSpec{
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: serviceName,
			},
			Port: &routev1.RoutePort{
				TargetPort: portNumber,
			},
		},
	}
	r, err := c.routeClient.Routes(c.namespace).Create(route)
	if err != nil {
		return nil, errors.Wrap(err, "error creating route")
	}
	return r, nil
}

// DeleteRoute deleted the given route
func (c *Client) DeleteRoute(name string) error {
	err := c.routeClient.Routes(c.namespace).Delete(name, &metav1.DeleteOptions{})
	if err != nil {
		return errors.Wrap(err, "unable to delete route")
	}
	return nil
}

// ListRoutes lists all the routes based on the given label selector
func (c *Client) ListRoutes(labelSelector string) ([]routev1.Route, error) {
	routeList, err := c.routeClient.Routes(c.namespace).List(metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get route list")
	}

	return routeList.Items, nil
}

// ListRouteNames lists all the names of the routes based on the given label
// selector
func (c *Client) ListRouteNames(labelSelector string) ([]string, error) {
	routes, err := c.ListRoutes(labelSelector)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list routes")
	}

	var routeNames []string
	for _, r := range routes {
		routeNames = append(routeNames, r.Name)
	}

	return routeNames, nil
}

// CreatePVC creates a PVC resource in the cluster with the given name, size and
// labels
func (c *Client) CreatePVC(name string, size string, labels map[string]string) (*corev1.PersistentVolumeClaim, error) {
	quantity, err := resource.ParseQuantity(size)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to parse size: %v", size)
	}

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: quantity,
				},
			},
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
		},
	}

	createdPvc, err := c.kubeClient.CoreV1().PersistentVolumeClaims(c.namespace).Create(pvc)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create PVC")
	}
	return createdPvc, nil
}

// DeletePVC deletes the given PVC by name
func (c *Client) DeletePVC(name string) error {
	return c.kubeClient.CoreV1().PersistentVolumeClaims(c.namespace).Delete(name, nil)
}

// generateVolumeNameFromPVC generates a random volume name based on the name
// of the given PVC
func generateVolumeNameFromPVC(pvc string) string {
	return fmt.Sprintf("%v-%v-volume", pvc, util.GenerateRandomString(nameLength))
}

// AddPVCToDeploymentConfig adds the given PVC to the given Deployment Config
// at the given path
func (c *Client) AddPVCToDeploymentConfig(dc *appsv1.DeploymentConfig, pvc string, path string) error {
	volumeName := generateVolumeNameFromPVC(pvc)

	// Validating dc.Spec.Template is present before dereferencing
	if dc.Spec.Template == nil {
		return fmt.Errorf("TemplatePodSpec in %s DeploymentConfig is empty", dc.Name)
	}
	dc.Spec.Template.Spec.Volumes = append(dc.Spec.Template.Spec.Volumes, corev1.Volume{
		Name: volumeName,
		VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: pvc,
			},
		},
	})

	// Validating dc.Spec.Template.Spec.Containers[] is present before dereferencing
	if len(dc.Spec.Template.Spec.Containers) == 0 {
		return fmt.Errorf("DeploymentConfig %s doesn't have any Containers defined", dc.Name)
	}
	dc.Spec.Template.Spec.Containers[0].VolumeMounts = append(dc.Spec.Template.Spec.Containers[0].VolumeMounts, corev1.VolumeMount{
		Name:      volumeName,
		MountPath: path,
	},
	)

	glog.V(4).Infof("Updating DeploymentConfig: %v", dc)
	_, err := c.appsClient.DeploymentConfigs(c.namespace).Update(dc)
	if err != nil {
		return errors.Wrapf(err, "failed to update DeploymentConfig: %v", dc)
	}
	return nil
}

// removeVolumeFromDC removes the volume from the given Deployment Config and
// returns true. If the given volume is not found, it returns false.
func removeVolumeFromDC(vol string, dc *appsv1.DeploymentConfig) bool {
	found := false
	for i, volume := range dc.Spec.Template.Spec.Volumes {
		if volume.Name == vol {
			found = true
			dc.Spec.Template.Spec.Volumes = append(dc.Spec.Template.Spec.Volumes[:i], dc.Spec.Template.Spec.Volumes[i+1:]...)
		}
	}
	return found
}

// removeVolumeMountFromDC removes the volumeMount from all the given containers
// in the given Deployment Config and return true. If the given volumeMount is
// not found, it returns false
func removeVolumeMountFromDC(vm string, dc *appsv1.DeploymentConfig) bool {
	found := false
	for i, container := range dc.Spec.Template.Spec.Containers {
		for j, volumeMount := range container.VolumeMounts {
			if volumeMount.Name == vm {
				found = true
				dc.Spec.Template.Spec.Containers[i].VolumeMounts = append(dc.Spec.Template.Spec.Containers[i].VolumeMounts[:j], dc.Spec.Template.Spec.Containers[i].VolumeMounts[j+1:]...)
			}
		}
	}
	return found
}

// RemoveVolumeFromDeploymentConfig removes the volume associated with the
// given PVC from the Deployment Config. Both, the volume entry and the
// volume mount entry in the containers, are deleted.
func (c *Client) RemoveVolumeFromDeploymentConfig(pvc string, dcName string) error {

	retryErr := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		dc, err := c.GetDeploymentConfigFromName(dcName, c.namespace)
		if err != nil {
			return errors.Wrapf(err, "unable to get Deployment Config: %v", dcName)
		}

		volumeNames := c.getVolumeNamesFromPVC(pvc, dc)
		numVolumes := len(volumeNames)
		if numVolumes == 0 {
			return fmt.Errorf("no volume found for PVC %v in DC %v, expected one", pvc, dc.Name)
		} else if numVolumes > 1 {
			return fmt.Errorf("found more than one volume for PVC %v in DC %v, expected one", pvc, dc.Name)
		}
		volumeName := volumeNames[0]

		// Remove volume if volume exists in Deployment Config
		if !removeVolumeFromDC(volumeName, dc) {
			return fmt.Errorf("could not find volume '%v' in Deployment Config '%v'", volumeName, dc.Name)
		}
		glog.V(4).Infof("Found volume: %v in Deployment Config: %v", volumeName, dc.Name)

		// Remove volume mount if volume mount exists
		if !removeVolumeMountFromDC(volumeName, dc) {
			return fmt.Errorf("could not find volumeMount: %v in Deployment Config: %v", volumeName, dc)
		}

		_, updateErr := c.appsClient.DeploymentConfigs(c.namespace).Update(dc)
		return updateErr
	})
	if retryErr != nil {
		return errors.Wrapf(retryErr, "updating Deployment Config %v failed", dcName)
	}
	return nil
}

// getVolumeNamesFromPVC returns the name of the volume associated with the given
// PVC in the given Deployment Config
func (c *Client) getVolumeNamesFromPVC(pvc string, dc *appsv1.DeploymentConfig) []string {
	var volumes []string
	for _, volume := range dc.Spec.Template.Spec.Volumes {
		if volume.PersistentVolumeClaim.ClaimName == pvc {
			volumes = append(volumes, volume.Name)
		}
	}
	return volumes
}

// GetDeploymentConfigsFromSelector returns an array of Deployment Config
// resources which match the given selector
func (c *Client) GetDeploymentConfigsFromSelector(selector string) ([]appsv1.DeploymentConfig, error) {
	dcList, err := c.appsClient.DeploymentConfigs(c.namespace).List(metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to list DeploymentConfigs")
	}
	return dcList.Items, nil
}

// GetServicesFromSelector returns an array of Service resources which match the
// given selector
func (c *Client) GetServicesFromSelector(selector string) ([]corev1.Service, error) {
	serviceList, err := c.kubeClient.CoreV1().Services(c.namespace).List(metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to list Services")
	}
	return serviceList.Items, nil
}

// GetDeploymentConfigFromName returns the Deployment Config resource given
// the Deployment Config name
func (c *Client) GetDeploymentConfigFromName(name string, project string) (*appsv1.DeploymentConfig, error) {
	glog.V(4).Infof("Getting DeploymentConfig: %s", name)
	deploymentConfig, err := c.appsClient.DeploymentConfigs(project).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get DeploymentConfig %s", name)
	}
	return deploymentConfig, nil

}

// GetPVCsFromSelector returns the PVCs based on the given selector
func (c *Client) GetPVCsFromSelector(selector string) ([]corev1.PersistentVolumeClaim, error) {
	pvcList, err := c.kubeClient.CoreV1().PersistentVolumeClaims(c.namespace).List(metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get PVCs for selector: %v", selector)
	}

	return pvcList.Items, nil
}

// GetPVCNamesFromSelector returns the PVC names for the given selector
func (c *Client) GetPVCNamesFromSelector(selector string) ([]string, error) {
	pvcs, err := c.GetPVCsFromSelector(selector)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get PVCs from selector")
	}

	var names []string
	for _, pvc := range pvcs {
		names = append(names, pvc.Name)
	}

	return names, nil
}

// GetOneDeploymentConfigFromSelector returns the Deployment Config object associated
// with the given selector.
// An error is thrown when exactly one Deployment Config is not found for the
// selector.
func (c *Client) GetOneDeploymentConfigFromSelector(selector string) (*appsv1.DeploymentConfig, error) {
	deploymentConfigs, err := c.GetDeploymentConfigsFromSelector(selector)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get DeploymentConfigs for the selector: %v", selector)
	}

	numDC := len(deploymentConfigs)
	if numDC == 0 {
		return nil, fmt.Errorf("no Deployment Config was found for the selector: %v", selector)
	} else if numDC > 1 {
		return nil, fmt.Errorf("multiple Deployment Configs exist for the selector: %v. Only one must be present", selector)
	}

	return &deploymentConfigs[0], nil
}

// GetOnePodFromSelector returns the Pod  object associated with the given selector.
// An error is thrown when exactly one Pod is not found.
func (c *Client) GetOnePodFromSelector(selector string) (*corev1.Pod, error) {

	pods, err := c.kubeClient.CoreV1().Pods(c.namespace).List(metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get Pod for the selector: %v", selector)
	}
	numPods := len(pods.Items)
	if numPods == 0 {
		return nil, fmt.Errorf("no Pod was found for the selector: %v", selector)
	} else if numPods > 1 {
		return nil, fmt.Errorf("multiple Pods exist for the selector: %v. Only one must be present", selector)
	}

	return &pods.Items[0], nil
}

// CopyFile copies localPath directory or list of files in copyFiles list to the directory in running Pod.
// copyFiles is list of changed files captured during `odo watch` as well as binary file path
// During copying binary components, localPath represent base directory path to binary and copyFiles contains path of binary
// During copying local source components, localPath represent base directory path whereas copyFiles is empty
// During `odo watch`, localPath represent base directory path whereas copyFiles contains list of changed Files
func (c *Client) CopyFile(localPath string, targetPodName string, targetPath string, copyFiles []string) error {
	dest := path.Join(targetPath, filepath.Base(localPath))
	reader, writer := io.Pipe()
	// inspired from https://github.com/kubernetes/kubernetes/blob/master/pkg/kubectl/cmd/cp.go#L235
	go func() {
		defer writer.Close()
		err := makeTar(localPath, dest, writer, copyFiles)
		if err != nil {
			glog.Errorf("Error while creating tar: %#v", err)
			os.Exit(-1)
		}

	}()

	// cmdArr will run inside container
	cmdArr := []string{"tar", "xf", "-", "-C", targetPath, "--strip", "1"}

	err := c.ExecCMDInContainer(targetPodName, cmdArr, writer, writer, reader, false)
	if err != nil {
		return err
	}
	return nil
}

// checkFileExist check if given file exists or not
func checkFileExist(fileName string) bool {
	_, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

// makeTar function is copied from https://github.com/kubernetes/kubernetes/blob/master/pkg/kubectl/cmd/cp.go#L309
// srcPath is ignored if files is set
func makeTar(srcPath, destPath string, writer io.Writer, files []string) error {
	// TODO: use compression here?
	tarWriter := taro.NewWriter(writer)
	defer tarWriter.Close()
	srcPath = path.Clean(srcPath)
	destPath = path.Clean(destPath)

	if len(files) != 0 {
		//watchTar
		for _, fileName := range files {
			if checkFileExist(fileName) {
				err := tar(tarWriter, fileName, path.Base(destPath))
				if err != nil {
					return err
				}

			}
		}
	} else {
		return recursiveTar(path.Dir(srcPath), path.Base(srcPath), path.Dir(destPath), path.Base(destPath), tarWriter)
	}

	return nil
}

// Tar will be used to tar files using odo watch
// inspired from https://gist.github.com/jonmorehouse/9060515
func tar(tw *taro.Writer, fileName string, destFile string) error {
	stat, _ := os.Lstat(fileName)

	// now lets create the header as needed for this file within the tarball
	hdr, err := taro.FileInfoHeader(stat, fileName)
	if err != nil {
		return err
	}
	splitFileName := strings.Split(fileName, destFile)[1]

	// hdr.Name can have only '/' as path separator, next line makes sure there is no '\'
	// in hdr.Name on Windows by replacing '\' to '/' in splitFileName. destFile is
	// a result of path.Base() call and never have '\' in it.
	hdr.Name = destFile + strings.Replace(splitFileName, "\\", "/", -1)
	// write the header to the tarball archive
	err = tw.WriteHeader(hdr)
	if err != nil {
		return err
	}

	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	// copy the file data to the tarball
	_, err = io.Copy(tw, file)
	if err != nil {
		return err
	}

	return nil
}

// recursiveTar function is copied from https://github.com/kubernetes/kubernetes/blob/master/pkg/kubectl/cmd/cp.go#L319
func recursiveTar(srcBase, srcFile, destBase, destFile string, tw *taro.Writer) error {
	filepath := path.Join(srcBase, srcFile)
	stat, err := os.Lstat(filepath)
	if err != nil {
		return err
	}
	if stat.IsDir() {
		files, err := ioutil.ReadDir(filepath)
		if err != nil {
			return err
		}
		if len(files) == 0 {
			//case empty directory
			hdr, _ := taro.FileInfoHeader(stat, filepath)
			hdr.Name = destFile
			if err := tw.WriteHeader(hdr); err != nil {
				return err
			}
		}
		for _, f := range files {
			if err := recursiveTar(srcBase, path.Join(srcFile, f.Name()), destBase, path.Join(destFile, f.Name()), tw); err != nil {
				return err
			}
		}
		return nil
	} else if stat.Mode()&os.ModeSymlink != 0 {
		//case soft link
		hdr, _ := taro.FileInfoHeader(stat, filepath)
		target, err := os.Readlink(filepath)
		if err != nil {
			return err
		}

		hdr.Linkname = target
		hdr.Name = destFile
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
	} else {
		//case regular file or other file type like pipe
		hdr, err := taro.FileInfoHeader(stat, filepath)
		if err != nil {
			return err
		}
		hdr.Name = destFile

		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}

		f, err := os.Open(filepath)
		if err != nil {
			return err
		}
		defer f.Close()

		if _, err := io.Copy(tw, f); err != nil {
			return err
		}
		return f.Close()
	}
	return nil
}

// GetOneServiceFromSelector returns the Service object associated with the
// given selector.
// An error is thrown when exactly one Service is not found for the selector
func (c *Client) GetOneServiceFromSelector(selector string) (*corev1.Service, error) {
	services, err := c.GetServicesFromSelector(selector)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get services for the selector: %v", selector)
	}

	numServices := len(services)
	if numServices == 0 {
		return nil, fmt.Errorf("no Service was found for the selector: %v", selector)
	} else if numServices > 1 {
		return nil, fmt.Errorf("multiple Services exist for the selector: %v. Only one must be present", selector)
	}

	return &services[0], nil
}

// AddEnvironmentVariablesToDeploymentConfig adds the given environment
// variables to the only container in the Deployment Config and updates in the
// cluster
func (c *Client) AddEnvironmentVariablesToDeploymentConfig(envs []corev1.EnvVar, dc *appsv1.DeploymentConfig) error {
	numContainers := len(dc.Spec.Template.Spec.Containers)
	if numContainers != 1 {
		return fmt.Errorf("expected exactly one container in Deployment Config %v, got %v", dc.Name, numContainers)
	}

	dc.Spec.Template.Spec.Containers[0].Env = append(dc.Spec.Template.Spec.Containers[0].Env, envs...)

	_, err := c.appsClient.DeploymentConfigs(c.namespace).Update(dc)
	if err != nil {
		return errors.Wrapf(err, "unable to update Deployment Config %v", dc.Name)
	}
	return nil
}

// serverInfo contains the fields that contain the server's information like
// address, OpenShift and Kubernetes versions
type serverInfo struct {
	Address           string
	OpenShiftVersion  string
	KubernetesVersion string
}

// GetServerVersion will fetch the Server Host, OpenShift and Kubernetes Version
// It will be shown on the execution of odo version command
func (c *Client) GetServerVersion() (*serverInfo, error) {
	var info serverInfo

	// This will fetch the information about Server Address
	config, err := c.kubeConfig.ClientConfig()
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get server's address")
	}
	info.Address = config.Host

	// This will fetch the information about OpenShift Version
	rawOpenShiftVersion, err := c.kubeClient.CoreV1().RESTClient().Get().AbsPath("/version/openshift").Do().Raw()
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get OpenShift Version")
	}
	var openShiftVersion version.Info
	if err := json.Unmarshal(rawOpenShiftVersion, &openShiftVersion); err != nil {
		return nil, errors.Wrapf(err, "unable to unmarshal OpenShift version %v", string(rawOpenShiftVersion))
	}
	info.OpenShiftVersion = openShiftVersion.GitVersion

	// This will fetch the information about Kubernetes Version
	rawKubernetesVersion, err := c.kubeClient.CoreV1().RESTClient().Get().AbsPath("/version").Do().Raw()
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

// ExecCMDInContainer execute command in first container of a pod
func (c *Client) ExecCMDInContainer(podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {

	req := c.kubeClient.CoreV1().RESTClient().
		Post().
		Namespace(c.namespace).
		Resource("pods").
		Name(podName).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Command: cmd,
			Stdin:   stdin != nil,
			Stdout:  stdout != nil,
			Stderr:  stderr != nil,
			TTY:     tty,
		}, scheme.ParameterCodec)

	config, err := c.kubeConfig.ClientConfig()
	if err != nil {
		return errors.Wrapf(err, "unable to get Kubernetes client config")
	}

	// Connect to url (constructed from req) using SPDY (HTTP/2) protocol which allows bidirectional streams.
	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return errors.Wrapf(err, "unable execute command via SPDY")
	}
	// initialize the transport of the standard shell streams
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
		Tty:    tty,
	})
	if err != nil {
		return errors.Wrapf(err, "error while streaming command")
	}

	return nil
}

// GetVolumeMountsFromDC returns a list of all volume mounts in the given DC
func (c *Client) GetVolumeMountsFromDC(dc *appsv1.DeploymentConfig) []corev1.VolumeMount {
	var volumeMounts []corev1.VolumeMount
	for _, container := range dc.Spec.Template.Spec.Containers {
		volumeMounts = append(volumeMounts, container.VolumeMounts...)
	}
	return volumeMounts
}

// GetPVCNameFromVolumeMountName returns the PVC associated with the given volume
// An empty string is returned if the volume is not found
func (c *Client) GetPVCNameFromVolumeMountName(volumeMountName string, dc *appsv1.DeploymentConfig) string {
	for _, volume := range dc.Spec.Template.Spec.Volumes {
		if volume.Name == volumeMountName {
			return volume.PersistentVolumeClaim.ClaimName
		}
	}
	return ""
}

// GetPVCFromName returns the PVC of the given name
func (c *Client) GetPVCFromName(pvcName string) (*corev1.PersistentVolumeClaim, error) {
	return c.kubeClient.CoreV1().PersistentVolumeClaims(c.namespace).Get(pvcName, metav1.GetOptions{})
}

// UpdatePVCLabels updates the given PVC with the given labels
func (c *Client) UpdatePVCLabels(pvc *corev1.PersistentVolumeClaim, labels map[string]string) error {
	pvc.Labels = labels
	_, err := c.kubeClient.CoreV1().PersistentVolumeClaims(c.namespace).Update(pvc)
	if err != nil {
		return errors.Wrap(err, "unable to remove storage label from PVC")
	}
	return nil
}

// getContainerPortsFromStrings generates ContainerPort values from the array of string port values
// ports is the array containing the string port values
func getContainerPortsFromStrings(ports []string) ([]corev1.ContainerPort, error) {
	var containerPorts []corev1.ContainerPort
	for _, port := range ports {
		splits := strings.Split(port, "/")
		if len(splits) < 1 || len(splits) > 2 {
			return nil, errors.Errorf("unable to parse the port string %s", port)
		}

		portNumberI64, err := strconv.ParseInt(splits[0], 10, 32)
		if err != nil {
			return nil, errors.Wrapf(err, "invalid port number %s", splits[0])
		}
		portNumber := int32(portNumberI64)

		var portProto corev1.Protocol
		if len(splits) == 2 {
			switch strings.ToUpper(splits[1]) {
			case "TCP":
				portProto = corev1.ProtocolTCP
			case "UDP":
				portProto = corev1.ProtocolUDP
			default:
				return nil, fmt.Errorf("invalid port protocol %s", splits[1])
			}
		} else {
			portProto = corev1.ProtocolTCP
		}

		port := corev1.ContainerPort{
			Name:          fmt.Sprintf("%d-%s", portNumber, strings.ToLower(string(portProto))),
			ContainerPort: portNumber,
			Protocol:      portProto,
		}
		containerPorts = append(containerPorts, port)
	}
	return containerPorts, nil
}
