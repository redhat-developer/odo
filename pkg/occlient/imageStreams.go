package occlient

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	dockerapiv10 "github.com/openshift/api/image/docker10"
	imagev1 "github.com/openshift/api/image/v1"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

// IsImageStreamSupported checks if imagestream resource type is present on the cluster
func (c *Client) IsImageStreamSupported() (bool, error) {

	return c.GetKubeClient().IsResourceSupported("image.openshift.io", "v1", "imagestreams")
}

// GetImageStream returns the imagestream using image details like imageNS, imageName and imageTag
// imageNS can be empty in which case, this function searches currentNamespace on priority. If
// imagestream of required tag not found in current namespace, then searches openshift namespace.
// If not found, error out. If imageNS is not empty string, then, the requested imageNS only is searched
// for requested imagestream
func (c *Client) GetImageStream(imageNS string, imageName string, imageTag string) (*imagev1.ImageStream, error) {
	var err error
	var imageStream *imagev1.ImageStream
	currentProjectName := c.GetCurrentProjectName()
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
		currentNSImageStream, e := c.imageClient.ImageStreams(currentProjectName).Get(context.TODO(), imageName, metav1.GetOptions{})
		if e != nil {
			err = errors.Wrapf(e, "no match found for : %s in namespace %s", imageName, currentProjectName)
		} else {
			if isTagInImageStream(*currentNSImageStream, imageTag) {
				return currentNSImageStream, nil
			}
		}

		// If not in current namespace, try finding imagestream from openshift namespace
		openshiftNSImageStream, e := c.imageClient.ImageStreams(OpenShiftNameSpace).Get(context.TODO(), imageName, metav1.GetOptions{})
		if e != nil {
			// The image is not available in current Namespace.
			err = errors.Wrapf(e, "no match found for : %s in namespace %s", imageName, OpenShiftNameSpace)
		} else {
			if isTagInImageStream(*openshiftNSImageStream, imageTag) {
				return openshiftNSImageStream, nil
			}
		}
		if e != nil && err != nil {
			return nil, fmt.Errorf("component type %q not found", imageName)
		}

		// Required tag not in openshift and current namespaces
		return nil, fmt.Errorf("image stream %s with tag %s not found in openshift and %s namespaces", imageName, imageTag, currentProjectName)

	}

	// Fetch imagestream from requested namespace
	imageStream, err = c.imageClient.ImageStreams(imageNS).Get(context.TODO(), imageName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(
			err, "no match found for %s in namespace %s", imageName, imageNS,
		)
	}
	if !isTagInImageStream(*imageStream, imageTag) {
		return nil, fmt.Errorf("image stream %s with tag %s not found in %s namespaces", imageName, imageTag, currentProjectName)
	}

	return imageStream, nil
}

// GetImageStreamImage returns image and error if any, corresponding to the passed imagestream and image tag
func (c *Client) GetImageStreamImage(imageStream *imagev1.ImageStream, imageTag string) (*imagev1.ImageStreamImage, error) {
	imageNS := imageStream.ObjectMeta.Namespace
	imageName := imageStream.ObjectMeta.Name

	for _, tag := range imageStream.Status.Tags {
		// look for matching tag
		if tag.Tag == imageTag {
			klog.V(3).Infof("Found exact image tag match for %s:%s", imageName, imageTag)

			if len(tag.Items) > 0 {
				tagDigest := tag.Items[0].Image
				imageStreamImageName := fmt.Sprintf("%s@%s", imageName, tagDigest)

				// look for imageStreamImage for given tag (reference by digest)
				imageStreamImage, err := c.imageClient.ImageStreamImages(imageNS).Get(context.TODO(), imageStreamImageName, metav1.GetOptions{})
				if err != nil {
					return nil, errors.Wrapf(err, "unable to find ImageStreamImage with  %s digest", imageStreamImageName)
				}
				return imageStreamImage, nil
			}

			return nil, fmt.Errorf("unable to find tag %s for image %s", imageTag, imageName)

		}
	}

	// return error since its an unhandled case if code reaches here
	return nil, fmt.Errorf("unable to find tag %s for image %s", imageTag, imageName)
}

// GetImageStreamTags returns all the ImageStreamTag objects in the given namespace
func (c *Client) GetImageStreamTags(namespace string) ([]imagev1.ImageStreamTag, error) {
	imageStreamTagList, err := c.imageClient.ImageStreamTags(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "unable to list imagestreamtags")
	}
	return imageStreamTagList.Items, nil
}

// GetPortsFromBuilderImage returns list of available port from given builder image of given component type
func (c *Client) GetPortsFromBuilderImage(componentType string) ([]string, error) {
	// checking port through builder image
	imageNS, imageName, imageTag, _, err := ParseImageName(componentType)
	if err != nil {
		return []string{}, err
	}
	imageStream, err := c.GetImageStream(imageNS, imageName, imageTag)
	if err != nil {
		return []string{}, err
	}
	imageStreamImage, err := c.GetImageStreamImage(imageStream, imageTag)
	if err != nil {
		return []string{}, err
	}
	containerPorts, err := getExposedPortsFromISI(imageStreamImage)
	if err != nil {
		return []string{}, err
	}
	var portList []string
	for _, po := range containerPorts {
		port := fmt.Sprint(po.ContainerPort) + "/" + string(po.Protocol)
		portList = append(portList, port)
	}
	if len(portList) == 0 {
		return []string{}, fmt.Errorf("given component type doesn't expose any ports, please use --port flag to specify a port")
	}
	return portList, nil
}

// ListImageStreams returns the Image Stream objects in the given namespace
func (c *Client) ListImageStreams(namespace string) ([]imagev1.ImageStream, error) {
	imageStreamList, err := c.imageClient.ImageStreams(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "unable to list imagestreams")
	}
	return imageStreamList.Items, nil
}

// ListImageStreamsNames returns the names of the image streams in a given
// namespace
func (c *Client) ListImageStreamsNames(namespace string) ([]string, error) {
	imageStreams, err := c.ListImageStreams(namespace)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get image streams")
	}

	var names []string
	for _, imageStream := range imageStreams {
		names = append(names, imageStream.Name)
	}
	return names, nil
}

// getExposedPortsFromISI parse ImageStreamImage definition and return all exposed ports in form of ContainerPorts structs
func getExposedPortsFromISI(image *imagev1.ImageStreamImage) ([]corev1.ContainerPort, error) {
	// file DockerImageMetadata
	err := imageWithMetadata(&image.Image)
	if err != nil {
		return nil, err
	}

	var ports []corev1.ContainerPort

	var exposedPorts = image.Image.DockerImageMetadata.Object.(*dockerapiv10.DockerImage).ContainerConfig.ExposedPorts

	if image.Image.DockerImageMetadata.Object.(*dockerapiv10.DockerImage).Config != nil {
		if exposedPorts == nil {
			exposedPorts = make(map[string]struct{})
		}

		// add ports from Config
		for exposedPort := range image.Image.DockerImageMetadata.Object.(*dockerapiv10.DockerImage).Config.ExposedPorts {
			var emptyStruct struct{}
			exposedPorts[exposedPort] = emptyStruct
		}
	}

	for exposedPort := range exposedPorts {
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

// getS2IMetaInfoFromBuilderImg returns script path protocol, S2I scripts path, S2I source or binary expected path, S2I deployment dir and errors(if any) from the passed builder image
func getS2IMetaInfoFromBuilderImg(builderImage *imagev1.ImageStreamImage) (S2IPaths, error) {

	// Define structs for internal un-marshalling of imagestreamimage to extract label from it
	type ContainerConfig struct {
		Labels     map[string]string `json:"Labels"`
		WorkingDir string            `json:"WorkingDir"`
	}
	type DockerImageMetaDataRaw struct {
		ContainerConfig ContainerConfig `json:"ContainerConfig"`
		Config          ContainerConfig `json:"Config"`
	}

	var dimdr DockerImageMetaDataRaw

	// The label $S2IScriptsURLLabel needs to be extracted from builderImage#Image#DockerImageMetadata#Raw which is byte array
	dimdrByteArr := (*builderImage).Image.DockerImageMetadata.Raw

	// Unmarshal the byte array into the struct for ease of access of required fields
	err := json.Unmarshal(dimdrByteArr, &dimdr)
	if err != nil {
		return S2IPaths{}, errors.Wrap(err, "unable to bootstrap supervisord")
	}

	labels := make(map[string]string)

	// Put labels from both ContainerConfig and Config into one map
	// if there is the same label in both, the Config has priority
	for k, v := range dimdr.ContainerConfig.Labels {
		labels[k] = v
	}
	for k, v := range dimdr.Config.Labels {
		labels[k] = v
	}

	// If by any chance, labels attribute is nil(although ideally not the case for builder images), return
	if len(labels) == 0 {
		klog.V(3).Infof("No Labels found in %+v in builder image %+v", dimdr, builderImage)
		return S2IPaths{}, nil
	}

	// Extract the label containing S2I scripts URL
	s2iScriptsURL := labels[S2IScriptsURLLabel]
	s2iSrcOrBinPath := labels[S2ISrcOrBinLabel]
	s2iBuilderImgName := labels[S2IBuilderImageName]

	if s2iSrcOrBinPath == "" {
		// In cases like nodejs builder image, where there is no concept of binary and sources are directly run, use destination as source
		// s2iSrcOrBinPath = getS2ILabelValue(dimdr.Config.Labels, S2IDeploymentsDir)
		s2iSrcOrBinPath = DefaultS2ISrcOrBinPath
	}

	s2iDestinationDir := getS2ILabelValue(labels, S2IDeploymentsDir)
	// The URL is a combination of protocol and the path to script details of which can be found @
	// https://github.com/openshift/source-to-image/blob/master/docs/builder_image.md#s2i-scripts
	// Extract them out into protocol and path separately to minimise the task in
	// https://github.com/openshift/odo-init-image/blob/master/assemble-and-restart when custom handling
	// for each of the protocols is added
	s2iScriptsProtocol := ""
	s2iScriptsPath := ""

	switch {
	case strings.HasPrefix(s2iScriptsURL, "image://"):
		s2iScriptsProtocol = "image://"
		s2iScriptsPath = strings.TrimPrefix(s2iScriptsURL, "image://")
	case strings.HasPrefix(s2iScriptsURL, "file://"):
		s2iScriptsProtocol = "file://"
		s2iScriptsPath = strings.TrimPrefix(s2iScriptsURL, "file://")
	case strings.HasPrefix(s2iScriptsURL, "http(s)://"):
		s2iScriptsProtocol = "http(s)://"
		s2iScriptsPath = s2iScriptsURL
	default:
		return S2IPaths{}, fmt.Errorf("Unknown scripts url %s", s2iScriptsURL)
	}
	return S2IPaths{
		ScriptsPathProtocol: s2iScriptsProtocol,
		ScriptsPath:         s2iScriptsPath,
		SrcOrBinPath:        s2iSrcOrBinPath,
		DeploymentDir:       s2iDestinationDir,
		WorkingDir:          dimdr.Config.WorkingDir,
		SrcBackupPath:       DefaultS2ISrcBackupDir,
		BuilderImgName:      s2iBuilderImgName,
	}, nil
}
