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

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	appsclientset "github.com/openshift/client-go/apps/clientset/versioned/typed/apps/v1"
	buildschema "github.com/openshift/client-go/build/clientset/versioned/scheme"
	buildclientset "github.com/openshift/client-go/build/clientset/versioned/typed/build/v1"
	imageclientset "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"

	appsv1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	imagev1 "github.com/openshift/api/image/v1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/openshift/source-to-image/pkg/tar"
	s2ifs "github.com/openshift/source-to-image/pkg/util/fs"

	dockerapiv10 "github.com/openshift/api/image/docker10"
)

const ocRequestTimeout = 1 * time.Second

type Client struct {
	ocpath      string
	kubeClient  *kubernetes.Clientset
	imageClient *imageclientset.ImageV1Client
	appsClient  *appsclientset.AppsV1Client
	buildClient *buildclientset.BuildV1Client
	namespace   string
}

func New() (*Client, error) {
	var client Client

	// initialize client-go clients
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	config, err := kubeConfig.ClientConfig()
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

	namespace, _, err := kubeConfig.Namespace()
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
		return nil, errors.New("server is down")
	}
	if !isLoggedIn(client.ocpath) {
		return nil, errors.New("please log in to the cluster")
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
		// data is given, assume this is crate or apply command
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

func (c *Client) GetCurrentProjectName() (string, error) {
	// We need to run `oc project` because it returns an error when project does
	// not exist, while `oc project -q` does not return an error, it simply
	// returns the project name
	_, err := c.runOcComamnd(&OcCommand{
		args: []string{"project"},
	})
	if err != nil {
		return "", errors.Wrap(err, "unable to get current project info")
	}

	output, err := c.runOcComamnd(&OcCommand{
		args: []string{"project", "-q"},
	})
	if err != nil {
		return "", errors.Wrap(err, "unable to get current project name")
	}

	return strings.TrimSpace(string(output)), nil
}

func (c *Client) GetProjects() (string, error) {
	output, err := c.runOcComamnd(&OcCommand{
		args:   []string{"get", "project"},
		format: "custom-columns=NAME:.metadata.name",
	})
	if err != nil {
		return "", errors.Wrap(err, "unable to get projects")
	}
	return strings.Join(strings.Split(string(output), "\n")[1:], "\n"), nil
}

func (c *Client) CreateNewProject(name string) error {
	_, err := c.runOcComamnd(&OcCommand{
		args: []string{"new-project", name},
	})
	if err != nil {
		return errors.Wrap(err, "unable to create new project")
	}
	return nil
}

func (c *Client) SetCurrentProject(project string) error {
	_, err := c.runOcComamnd(&OcCommand{
		args: []string{"project", project},
	})
	if err != nil {
		return errors.Wrap(err, "unable to set current project to "+project)
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

// NewAppS2I create new application using S2I
// if gitUrl is ""  than it creates binary build otherwise uses gitUrl as buildSource
func (c *Client) NewAppS2I(name string, builderImage string, gitUrl string, labels map[string]string) (string, error) {
	openShiftNameSpace := "openshift"

	imageName, imageTag, _, err := parseImageName(builderImage)
	if err != nil {
		return "", errors.Wrap(err, "unable to create new s2i git build ")
	}

	var containerPorts []corev1.ContainerPort

	log.Debugf("Checking for exact match of builderImage with ImageStream")
	imageStream, err := c.imageClient.ImageStreams(openShiftNameSpace).Get(imageName, metav1.GetOptions{})
	if err != nil {
		log.Debugf("No exact match found: %s", err.Error())
		// TODO: search elsewhere
		return "", errors.Wrapf(err, "unable to find matching builder image %s", imageName)
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
					return "", errors.Wrapf(err, "unable to find ImageStreamImage with  %s digest", imageStreamImageName)
				}
				// get ports that are exported by image
				containerPorts, err = getExposedPorts(imageStreamImage)
				if err != nil {
					return "", errors.Wrapf(err, "unable to get exported ports from % image", builderImage)
				}
			}
		}
		if !tagFound {
			return "", errors.Wrapf(err, "unable to find tag %s for image", imageTag, imageName)
		}

	}

	// generate and create ImageStream
	is := imagev1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
	}
	_, err = c.imageClient.ImageStreams(c.namespace).Create(&is)
	if err != nil {
		panic(err.Error())
	}

	// generate BuildConfig
	buildSource := buildv1.BuildSource{
		Type:   "Binary",
		Binary: &buildv1.BinaryBuildSource{},
	}
	// if gitUrl set change buildSource to git and use given repo
	if gitUrl != "" {
		buildSource = buildv1.BuildSource{
			Git: &buildv1.GitBuildSource{
				URI: gitUrl,
			},
			Type: "Git",
		}
	}

	bc := buildv1.BuildConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
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
							Namespace: openShiftNameSpace,
						},
					},
				},
			},
			Triggers: []buildv1.BuildTriggerPolicy{
				{
					Type: "ConfigChange",
				},
				{
					Type:        "ImageChange",
					ImageChange: &buildv1.ImageChangeTrigger{},
				},
			},
		},
	}
	_, err = c.buildClient.BuildConfigs(c.namespace).Create(&bc)
	if err != nil {
		panic(err.Error())
	}

	// generate  and create DeploymentConfig
	dc := appsv1.DeploymentConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
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
		panic(err.Error())
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
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Spec: corev1.ServiceSpec{
			Ports: svcPorts,
			Selector: map[string]string{
				"deploymentconfig": name,
			},
		},
	}
	_, err = c.kubeClient.CoreV1().Services(c.namespace).Create(&svc)
	if err != nil {
		panic(err.Error())
	}

	return "", nil

}

func (c *Client) StartBuild(name string, dir string) (string, error) {
	var r io.Reader
	pr, pw := io.Pipe()
	go func() {
		w := gzip.NewWriter(pw)
		if err := tar.New(s2ifs.NewFileSystem()).CreateTarStream(dir, false, w); err != nil {
			pw.CloseWithError(err)
		} else {
			w.Close()
			pw.CloseWithError(io.EOF)
		}
	}()
	r = pr

	buildRequest := buildv1.BuildRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
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
		return "", errors.Wrapf(err, "unable to start build %s", name)
	}
	log.Debug("Build %s from %s directory triggered.", name, dir)

	err = c.FollowBuildLog(result.Name)
	if err != nil {
		return "", errors.Wrapf(err, "unable to start build %s", name)
	}

	return "", nil
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

	stdout := os.Stdout

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

type VolumeConfig struct {
	Name             *string
	Size             *string
	DeploymentConfig *string
	Path             *string
}

type VolumeOpertaions struct {
	Add    bool
	Remove bool
	List   bool
}

func (c *Client) SetVolumes(config *VolumeConfig, operations *VolumeOpertaions) (string, error) {
	args := []string{
		"set",
		"volumes",
		"dc", *config.DeploymentConfig,
		"--type", "pvc",
	}
	if config.Name != nil {
		args = append(args, "--name", *config.Name)
	}
	if config.Size != nil {
		args = append(args, "--claim-size", *config.Size)
	}
	if config.Path != nil {
		args = append(args, "--mount-path", *config.Path)
	}
	if operations.Add {
		args = append(args, "--add")
	}
	if operations.Remove {
		args = append(args, "--remove", "--confirm")
	}
	if operations.List {
		args = append(args, "--all")
	}
	output, err := c.runOcComamnd(&OcCommand{
		args: args,
	})
	if err != nil {
		return "", errors.Wrap(err, "unable to perform volume operations")
	}
	return string(output), nil
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
