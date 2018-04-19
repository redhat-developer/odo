package occlient

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	servicecatalogclienset "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	appsclientset "github.com/openshift/client-go/apps/clientset/versioned/typed/apps/v1"
	buildschema "github.com/openshift/client-go/build/clientset/versioned/scheme"
	buildclientset "github.com/openshift/client-go/build/clientset/versioned/typed/build/v1"
	imageclientset "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"
	projectclientset "github.com/openshift/client-go/project/clientset/versioned/typed/project/v1"
	routeclientset "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"

	appsv1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	imagev1 "github.com/openshift/api/image/v1"
	projectv1 "github.com/openshift/api/project/v1"
	routev1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	scv1beta1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"

	"github.com/openshift/source-to-image/pkg/tar"
	s2ifs "github.com/openshift/source-to-image/pkg/util/fs"

	dockerapiv10 "github.com/openshift/api/image/docker10"
	"github.com/redhat-developer/odo/pkg/util"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/util/retry"
	"path/filepath"
)

const (
	ocRequestTimeout   = 1 * time.Second
	OpenShiftNameSpace = "openshift"

	// The length of the string to be generated for names of resources
	nameLength = 5
)

type Client struct {
	ocpath               string
	kubeClient           *kubernetes.Clientset
	imageClient          *imageclientset.ImageV1Client
	appsClient           *appsclientset.AppsV1Client
	buildClient          *buildclientset.BuildV1Client
	projectClient        *projectclientset.ProjectV1Client
	serviceCatalogClient *servicecatalogclienset.ServicecatalogV1beta1Client
	routeClient          *routeclientset.RouteV1Client
	kubeConfig           clientcmd.ClientConfig
	namespace            string
}

func New() (*Client, error) {
	var client Client

	// initialize client-go clients
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	client.kubeConfig = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	config, err := client.kubeConfig.ClientConfig()
	if err != nil {
		return nil, err
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

	namespace, _, err := client.kubeConfig.Namespace()
	if err != nil {
		return nil, err
	}
	client.namespace = namespace

	// The following should go away once we're done with complete migration to
	// client-go
	ocpath, err := getOcBinary()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get oc binary")
	}
	client.ocpath = ocpath

	if !isServerUp(client.ocpath) {
		return nil, errors.New("Unable to connect to OpenShift cluster, is it down?")
	}
	if !isLoggedIn(client.ocpath) {
		return nil, errors.New("Please log in to the cluster")
	}

	return &client, nil
}

// parseImageName parse image reference
// returns (imageName, tag, digest, error)
// if image is referenced by tag (name:tag)  than digest is ""
// if image is referenced by digest (name@digest) than  tag is ""
func parseImageName(image string) (string, string, string, error) {
	digestParts := strings.Split(image, "@")
	if len(digestParts) == 2 {
		// image is references digest
		return digestParts[0], "", digestParts[1], nil
	} else if len(digestParts) == 1 {
		tagParts := strings.Split(image, ":")
		if len(tagParts) == 2 {
			// image references tag
			return tagParts[0], tagParts[1], "", nil
		} else if len(tagParts) == 1 {
			return tagParts[0], "latest", "", nil
		}
	}
	return "", "", "", fmt.Errorf("invalid image reference %s", image)

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

// getOcBinary returns full path to oc binary
// first it looks for env variable KUBECTL_PLUGINS_CALLER (run as oc plugin)
// than looks for env variable OC_BIN (set manualy by user)
// at last it tries to find oc in default PATH
func getOcBinary() (string, error) {
	log.Debug("getOcBinary - searching for oc binary")

	var ocPath string

	envKubectlPluginCaller := os.Getenv("KUBECTL_PLUGINS_CALLER")
	envOcBin := os.Getenv("OC_BIN")

	log.Debugf("envKubectlPluginCaller = %s\n", envKubectlPluginCaller)
	log.Debugf("envOcBin = %s\n", envOcBin)

	if len(envKubectlPluginCaller) > 0 {
		log.Debug("using path from KUBECTL_PLUGINS_CALLER")
		ocPath = envKubectlPluginCaller
	} else if len(envOcBin) > 0 {
		log.Debug("using path from OC_BIN")
		ocPath = envOcBin
	} else {
		path, err := exec.LookPath("oc")
		if err != nil {
			log.Debug("oc binary not found in PATH")
			return "", err
		}
		log.Debug("using oc from PATH")
		ocPath = path
	}
	log.Debug("using oc from %s", ocPath)

	if _, err := os.Stat(ocPath); err != nil {
		return "", err
	}

	return ocPath, nil
}

type OcCommand struct {
	args   []string
	data   *string
	format string
}

// runOcCommands executes oc
// args - command line arguments to be passed to oc ('-o json' is added by default if data is not nil)
// data - is a pointer to a string, if set than data is given to command to stdin ('-f -' is added to args as default)
func (c *Client) runOcComamnd(command *OcCommand) ([]byte, error) {
	cmd := exec.Command(c.ocpath, command.args...)

	// if data is not set assume that it is get command
	if len(command.format) > 0 {
		cmd.Args = append(cmd.Args, "-o", command.format)
	}
	if command.data != nil {
		// data is given, assume this is create or apply command
		// that takes data from stdin
		cmd.Args = append(cmd.Args, "-f", "-")

		// Read from stdin
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return nil, err
		}

		// Write to stdin
		go func() {
			defer stdin.Close()
			_, err := io.WriteString(stdin, *command.data)
			if err != nil {
				fmt.Printf("can't write to stdin %v\n", err)
			}
		}()
	}

	log.Debugf("running oc command with arguments: %s\n", strings.Join(cmd.Args, " "))

	output, err := cmd.CombinedOutput()
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return nil, errors.Wrapf(err, "command: %v failed to run:\n%v", cmd.Args, string(output))
		}
		return nil, errors.Wrap(err, "unable to get combined output")
	}

	return output, nil
}

func isLoggedIn(ocpath string) bool {
	cmd := exec.Command(ocpath, "whoami")
	output, err := cmd.CombinedOutput()
	log.Debugf("isLoggedIn err:  %#v \n output: %#v", err, string(output))
	if err != nil {
		log.Debug(errors.Wrap(err, "error running command"))
		log.Debugf("Output is: %v", output)
		return false
	}
	return true
}

func isServerUp(ocpath string) bool {
	cmd := exec.Command(ocpath, "whoami", "--show-server")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Debug(errors.Wrap(err, "error running command"))
		return false
	}

	server := strings.TrimSpace(string(output))
	u, err := url.Parse(server)
	if err != nil {
		log.Debug(errors.Wrap(err, "unable to parse url"))
		return false
	}

	log.Debugf("Trying to connect to server %v - %v", u.Host)
	_, connectionError := net.DialTimeout("tcp", u.Host, time.Duration(ocRequestTimeout))
	if connectionError != nil {
		log.Debug(errors.Wrap(connectionError, "unable to connect to server"))
		return false
	}
	log.Debugf("Server %v is up", server)
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

// getExposedPorts parse ImageStreamImage definition and return all exposed ports in form of ContainerPorts structs
func getExposedPorts(image *imagev1.ImageStreamImage) ([]corev1.ContainerPort, error) {
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

// NewAppS2I create new application using S2I
// if gitUrl is ""  than it creates binary build otherwise uses gitUrl as buildSource
func (c *Client) NewAppS2I(name string, builderImage string, gitUrl string, labels map[string]string, annotations map[string]string) error {

	imageName, imageTag, _, err := parseImageName(builderImage)
	if err != nil {
		return errors.Wrap(err, "unable to create new s2i git build ")
	}

	var containerPorts []corev1.ContainerPort

	log.Debugf("Checking for exact match of builderImage with ImageStream")
	imageStream, err := c.imageClient.ImageStreams(OpenShiftNameSpace).Get(imageName, metav1.GetOptions{})
	if err != nil {
		log.Debugf("No exact match found: %s", err.Error())
		return errors.Wrapf(err, "unable to find matching builder image %s", imageName)
	} else {
		tagFound := false
		for _, tag := range imageStream.Status.Tags {
			// look for matching tag
			if tag.Tag == imageTag {
				tagFound = true
				log.Debugf("Found exact image tag match for %s:%s", imageName, imageTag)
				// ImageStream holds tag history
				// first item is the latest one
				tagDigest := tag.Items[0].Image
				// look for imageStreamImage for given tag (reference by digest)
				imageStreamImageName := fmt.Sprintf("%s@%s", imageName, tagDigest)
				imageStreamImage, err := c.imageClient.ImageStreamImages("openshift").Get(imageStreamImageName, metav1.GetOptions{})
				if err != nil {
					return errors.Wrapf(err, "unable to find ImageStreamImage with  %s digest", imageStreamImageName)
				}
				// get ports that are exported by image
				containerPorts, err = getExposedPorts(imageStreamImage)
				if err != nil {
					return errors.Wrapf(err, "unable to get exported ports from % image", builderImage)
				}
			}
		}
		if !tagFound {
			return errors.Wrapf(err, "unable to find tag %s for image", imageTag, imageName)
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
		Type:   buildv1.BuildSourceBinary,
		Binary: &buildv1.BinaryBuildSource{},
	}
	// if gitUrl set change buildSource to git and use given repo
	if gitUrl != "" {
		buildSource = buildv1.BuildSource{
			Git: &buildv1.GitBuildSource{
				URI: gitUrl,
			},
			Type: buildv1.BuildSourceGit,
		}
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
							Namespace: OpenShiftNameSpace,
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

	// generate and create Service
	var svcPorts []corev1.ServicePort
	for _, containerPort := range dc.Spec.Template.Spec.Containers[0].Ports {
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
				"deploymentconfig": name,
			},
		},
	}
	_, err = c.kubeClient.CoreV1().Services(c.namespace).Create(&svc)
	if err != nil {
		return errors.Wrapf(err, "unable to create Service for %s", name)
	}

	return nil

}

// UpdateBuildConfig updates the BuildConfig file
// buildConfigName is the name of the BuildConfig file to be updated
// projectName is the name of the project
// gitUrl equals to the git URL of the source and is equals to "" if the source is of type dir or binary
// annotations contains the annotations for the BuildConfig file
func (c *Client) UpdateBuildConfig(buildConfigName string, projectName string, gitUrl string, annotations map[string]string) error {
	// generate BuildConfig
	buildSource := buildv1.BuildSource{
		Type:   buildv1.BuildSourceBinary,
		Binary: &buildv1.BinaryBuildSource{},
	}
	// if gitUrl set change buildSource to git and use given repo
	if gitUrl != "" {
		buildSource = buildv1.BuildSource{
			Git: &buildv1.GitBuildSource{
				URI: gitUrl,
			},
			Type: buildv1.BuildSourceGit,
		}
	}
	buildConfig, err := c.GetBuildConfig(buildConfigName, projectName)
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

// StartBinaryBuild starts new build and streams dir as source for build
// asFile indicates a build will start with a single file specified in the path otherwise it assumes path is a directory
func (c *Client) StartBinaryBuild(name string, path string, asFile bool) error {
	var r io.Reader

	buildRequest := buildv1.BinaryBuildRequestOptions{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}

	if !asFile {
		pr, pw := io.Pipe()
		go func() {
			w := gzip.NewWriter(pw)
			if err := tar.New(s2ifs.NewFileSystem()).CreateTarStream(path, false, w); err != nil {
				pw.CloseWithError(err)
			} else {
				w.Close()
				pw.CloseWithError(io.EOF)
			}
		}()
		r = pr
	} else {
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		r = f
		buildRequest.AsFile = filepath.Base(path)
	}

	result := &buildv1.Build{}
	// this should be  buildClient.BuildConfigs(namespace).Instantiate
	// but there is no way to pass data using that call.
	err := c.buildClient.RESTClient().Post().
		Namespace(c.namespace).
		Resource("buildconfigs").
		Name(name).
		SubResource("instantiatebinary").
		Body(r).
		VersionedParams(&buildRequest, buildschema.ParameterCodec).
		Do().
		Into(result)

	if err != nil {
		return errors.Wrapf(err, "unable to start build %s", name)
	}
	log.Debugf("Build %s from %s directory triggered.", name, path)

	err = c.FollowBuildLog(result.Name)
	if err != nil {
		return errors.Wrapf(err, "unable to start build %s", name)
	}

	return nil
}

// StartBuild starts new build as it is
func (c *Client) StartBuild(name string) error {
	buildRequest := buildv1.BuildRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	result, err := c.buildClient.BuildConfigs(c.namespace).Instantiate(name, &buildRequest)
	if err != nil {
		return errors.Wrapf(err, "unable to start build %s", name)
	}
	log.Debugf("Build %s triggered.", name)

	err = c.FollowBuildLog(result.Name)
	if err != nil {
		return errors.Wrapf(err, "unable to start build %s", name)
	}

	return nil
}

// FollowBuildLog stream build log to stdout
func (c *Client) FollowBuildLog(buildName string) error {
	buildLogOpetions := buildv1.BuildLogOptions{
		Follow: true,
		NoWait: false,
	}

	rd, err := c.buildClient.RESTClient().Get().
		Namespace(c.namespace).
		Resource("builds").
		Name(buildName).
		SubResource("log").
		VersionedParams(&buildLogOpetions, buildschema.ParameterCodec).
		Stream()

	if err != nil {
		return errors.Wrapf(err, "unable get build log %s", buildName)
	}
	defer rd.Close()

	// Set the colour of the stdout output..
	color.Set(color.FgYellow)
	defer color.Unset()

	stdout := color.Output

	if _, err = io.Copy(stdout, rd); err != nil {
		return errors.Wrapf(err, "error streaming logs for %s", buildName)
	}

	return nil
}

// Delete calls oc delete
// kind is always required (can be set to 'all')
// name can be omitted if labels are set, in that case set name to ''
// if you want to delete object just by its name set labels to nil
func (c *Client) Delete(kind string, name string, labels map[string]string) (string, error) {

	args := []string{
		"delete",
		kind,
	}

	if len(name) > 0 {
		args = append(args, name)
	}

	if labels != nil {
		var labelsString []string
		for key, value := range labels {
			labelsString = append(labelsString, fmt.Sprintf("%s=%s", key, value))
		}
		args = append(args, "--selector")
		args = append(args, strings.Join(labelsString, ","))
	}

	output, err := c.runOcComamnd(&OcCommand{args: args})
	if err != nil {
		return "", err
	}

	return string(output[:]), nil

}

func (c *Client) DeleteProject(name string) error {
	_, err := c.runOcComamnd(&OcCommand{
		args: []string{"delete", "project", name},
	})
	if err != nil {
		return errors.Wrap(err, "unable to delete project")
	}
	return nil
}

// GetLabelValues get label values of given label from objects in project that are matching selector
// returns slice of uniq label values
func (c *Client) GetLabelValues(project string, label string, selector string) ([]string, error) {
	// get all object that have given label
	// and show just label values separated by ,
	args := []string{
		"get", "all",
		"--selector", selector,
		"--namespace", project,
		"-o", "go-template={{range .items}}{{range $key, $value := .metadata.labels}}{{if eq $key \"" + label + "\"}}{{$value}},{{end}}{{end}}{{end}}",
	}

	output, err := c.runOcComamnd(&OcCommand{args: args})
	if err != nil {
		return nil, err
	}

	var values []string
	// deduplicate label values
	for _, val := range strings.Split(string(output), ",") {
		val = strings.TrimSpace(val)
		if val != "" {
			// check if this is the first time we see this value
			found := false
			for _, existing := range values {
				if val == existing {
					found = true
				}
			}
			if !found {
				values = append(values, val)
			}
		}
	}

	return values, nil
}

// GetBuildConfig get BuildConfig by its name
func (c *Client) GetBuildConfig(name string, project string) (*buildv1.BuildConfig, error) {
	log.Debugf("Getting BuildConfig: %s", name)
	bc, err := c.buildClient.BuildConfigs(project).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get BuildConfig %s", name)
	}
	return bc, nil
}

// GetClusterServiceClasses queries the service service catalog to get all the
// currently available cluster service classes
func (c *Client) GetClusterServiceClasses() ([]scv1beta1.ClusterServiceClass, error) {
	classList, err := c.serviceCatalogClient.ClusterServiceClasses().List(metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "unable to list cluster service classes")
	}
	return classList.Items, nil
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
		log.Debugf("unable to get image streams in the namespace: %v", namespace)
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
		log.Debugf("unable to get cluster service classes' external names")
	}

	for _, class := range clusterServiceClasses {
		if class == name {
			return true
		}
	}

	return false
}

// CreateRoute creates a route object for the given service and with the given
// labels
func (c *Client) CreateRoute(service string, labels map[string]string) (*routev1.Route, error) {
	route := &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:   service,
			Labels: labels,
		},
		Spec: routev1.RouteSpec{
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: service,
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
	return errors.Wrap(c.routeClient.Routes(c.namespace).Delete(name, &metav1.DeleteOptions{}), "unable to delete route")
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

	dc.Spec.Template.Spec.Volumes = append(dc.Spec.Template.Spec.Volumes, corev1.Volume{
		Name: volumeName,
		VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: pvc,
			},
		},
	})

	dc.Spec.Template.Spec.Containers[0].VolumeMounts = append(dc.Spec.Template.Spec.Containers[0].VolumeMounts, corev1.VolumeMount{
		Name:      volumeName,
		MountPath: path,
	},
	)

	log.Debugf("Updating DeploymentConfig: %v", dc)
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
		dc, err := c.GetDeploymentConfigFromName(dcName)
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
		log.Debugf("Found volume: %v in Deployment Config: %v", volumeName, dc.Name)

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
func (c *Client) GetDeploymentConfigFromName(name string) (*appsv1.DeploymentConfig, error) {
	return c.appsClient.DeploymentConfigs(c.namespace).Get(name, metav1.GetOptions{})
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
	numDC := len(pods.Items)
	if numDC == 0 {
		return nil, fmt.Errorf("no Pod was found for the selector: %v", selector)
	} else if numDC > 1 {
		return nil, fmt.Errorf("multiple Pods exist for the selector: %v. Only one must be present", selector)
	}

	return &pods.Items[0], nil
}

// SyncPath copies local directory to directory in running Pod.
func (c *Client) SyncPath(localPath string, targetPodName string, targetPath string) (string, error) {
	// TODO: do this without using 'oc' binary
	args := []string{
		"rsync",
		localPath,
		fmt.Sprintf("%s:%s", targetPodName, targetPath),
		"--exclude", ".git",
		"--no-perms",
	}

	output, err := c.runOcComamnd(&OcCommand{args: args})
	if err != nil {
		return "", err
	}

	log.Debugf("command output:\n %s \n", string(output[:]))
	return string(output[:]), nil
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
