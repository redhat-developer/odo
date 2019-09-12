package catalog

import (
	"fmt"
	"strings"

	"github.com/golang/glog"
	imagev1 "github.com/openshift/api/image/v1"
	"github.com/openshift/odo/pkg/occlient"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// List lists all the available component types
func List(client *occlient.Client) (CatalogImageList, error) {

	catalogList, err := getDefaultBuilderImages(client)
	if err != nil {
		return CatalogImageList{}, errors.Wrap(err, "unable to get image streams")
	}

	if len(catalogList) == 0 {
		return CatalogImageList{}, errors.New("unable to retrieve any catalog images from the OpenShift cluster")
	}

	return CatalogImageList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CatalogList",
			APIVersion: "odo.openshift.io/v1alpha1",
		},
		Items: catalogList,
	}, nil
}

// Search searches for the component
func Search(client *occlient.Client, name string) ([]string, error) {
	var result []string
	componentList, err := List(client)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list components")
	}

	// do a partial search in all the components
	for _, component := range componentList.Items {
		// we only show components that contain the search term and that have at least non-hidden tag
		// since a component with all hidden tags is not shown in the odo catalog list components either
		if strings.Contains(component.ObjectMeta.Name, name) && len(component.Spec.NonHiddenTags) > 0 {
			result = append(result, component.ObjectMeta.Name)
		}
	}

	return result, nil
}

// Exists returns true if the given component type and the version are valid, false if not
func Exists(client *occlient.Client, componentType string, componentVersion string) (bool, error) {
	imageStream, err := client.GetImageStream("", componentType, componentVersion)
	if err != nil {
		return false, errors.Wrapf(err, "unable to get from catalog")
	}
	if imageStream == nil {
		return false, nil
	}
	return true, nil
}

// getDefaultBuilderImages returns the default builder images available in the
// openshift and the current namespaces
func getDefaultBuilderImages(client *occlient.Client) ([]Catalog, error) {

	var imageStreams []imagev1.ImageStream
	currentNamespace := client.GetCurrentProjectName()

	// Fetch imagestreams from default openshift namespace
	openshiftNSImageStreams, openshiftNSISFetchError := client.GetImageStreams(occlient.OpenShiftNameSpace)
	if openshiftNSISFetchError != nil {
		// Tolerate the error as it might only be a partial failure
		// We may get the imagestreams from other Namespaces
		//err = errors.Wrapf(openshiftNSISFetchError, "unable to get Image Streams from namespace %s", occlient.OpenShiftNameSpace)
		// log it for debugging purposes
		glog.V(4).Infof("Unable to get Image Streams from namespace %s. Error %s", occlient.OpenShiftNameSpace, openshiftNSISFetchError.Error())
	}

	// Fetch imagestreams from current namespace
	currentNSImageStreams, currentNSISFetchError := client.GetImageStreams(currentNamespace)
	// If failure to fetch imagestreams from current namespace, log the failure for debugging purposes
	if currentNSISFetchError != nil {
		// Tolerate the error as it is totally a valid scenario to not have any imagestreams in current namespace
		// log it for debugging purposes
		glog.V(4).Infof("Unable to get Image Streams from namespace %s. Error %s", currentNamespace, currentNSISFetchError.Error())
	}

	// If failure fetching imagestreams from both namespaces, error out
	if openshiftNSISFetchError != nil && currentNSISFetchError != nil {
		return nil, errors.Wrapf(
			fmt.Errorf("%s.\n%s", openshiftNSISFetchError, currentNSISFetchError),
			"Failed to fetch imagestreams from both openshift and %s namespaces.\nPlease ensure that a builder imagestream of required version for the component exists in either openshift or %s namespaces",
			currentNamespace,
			currentNamespace,
		)
	}

	// Resultant imagestreams is list of imagestreams from current and openshift namespaces
	imageStreams = append(imageStreams, openshiftNSImageStreams...)
	imageStreams = append(imageStreams, currentNSImageStreams...)

	// create a map from name (builder image name + tag) to the ImageStreamTag
	// we need this in order to filter out hidden tags
	imageStreamTagMap := make(map[string]imagev1.ImageStreamTag)

	currentNSImageStreamTags, currentNSImageStreamTagsErr := client.GetImageStreamTags(currentNamespace)
	openshiftNSImageStreamTags, openshiftNSImageStreamTagsErr := client.GetImageStreamTags(occlient.OpenShiftNameSpace)

	// If failure fetching imagestreamtags from both namespaces, error out
	if currentNSImageStreamTagsErr != nil && openshiftNSImageStreamTagsErr != nil {
		return nil, errors.Wrapf(
			fmt.Errorf("%s.\n%s", currentNSImageStreamTagsErr, openshiftNSImageStreamTagsErr),
			"Failed to fetch imagestreamtags from both openshift and %s namespaces.\nPlease ensure that a builder imagestream of required version for the component exists in either openshift or %s namespaces",
			currentNamespace,
			currentNamespace,
		)
	}

	// create a map from name to ImageStreamTag out of all the ImageStreamTag objects we collect
	var imageStreamTags []imagev1.ImageStreamTag
	imageStreamTags = append(imageStreamTags, currentNSImageStreamTags...)
	imageStreamTags = append(imageStreamTags, openshiftNSImageStreamTags...)
	for _, imageStreamTag := range imageStreamTags {
		imageStreamTagMap[imageStreamTag.Name] = imageStreamTag
	}

	builderImages := getBuildersFromImageStreams(imageStreams, imageStreamTagMap)

	return builderImages, nil
}

// SliceSupportedTags splits the tags in to fully supported and unsupported tags
func SliceSupportedTags(catalogImage Catalog) ([]string, []string) {

	var supTag, unSupTag []string
	tagMap := createImageTagMap(catalogImage.Spec.ImageStreamRef.Spec.Tags)
	for _, tag := range catalogImage.Spec.NonHiddenTags {
		imageName := tagMap[tag]
		if isSupportedImage(imageName) {
			supTag = append(supTag, tag)
		} else {
			unSupTag = append(unSupTag, tag)
		}
	}
	return supTag, unSupTag
}

// createImageTagMap takes a list of image TagReferences and creates a map of type tag name => image name e.g. 1.11 => openshift/nodejs-11
func createImageTagMap(tagRefs []imagev1.TagReference) map[string]string {
	tagMap := make(map[string]string)
	for _, tagRef := range tagRefs {
		imageName := tagRef.From.Name
		if tagRef.From.Kind == "DockerImage" {
			// we get the image name from the repo url e.g. registry.redhat.com/openshift/nodejs:10 will give openshift/nodejs:10
			urlImageName := strings.SplitN(imageName, "/", 2)[1]
			// here we remove the tag and digest
			ns, img, _, _, _ := occlient.ParseImageName(urlImageName)
			imageName = ns + "/" + img
		} else if tagRef.From.Kind == "ImageStreamTag" {
			tagList := strings.Split(imageName, ":")
			tag := tagList[len(tagList)-1]
			// if the kind is a image stream tag that means its pointing to an existing dockerImage or image stream image
			// we just look it up from the tapMap we already have
			imageName = tagMap[tag]
		}
		tagMap[tagRef.Name] = imageName
	}
	return tagMap
}

// isSupportedImages returns if the image is supported or not. the supported images have been provided here
// https://github.com/openshift/odo-init-image/blob/master/language-scripts/image-mappings.json
func isSupportedImage(imgName string) bool {
	supportedImages := []string{
		"redhat-openjdk-18/openjdk18-openshift",
		"openjdk/openjdk-11-rhel8",
		"openjdk/openjdk-11-rhel7",
		"rhscl/nodejs-8-rhel7",
		"rhoar-nodejs/nodejs-8",
		"rhoar-nodejs/nodejs-10",
	}
	for _, supImage := range supportedImages {
		if supImage == imgName {
			return true
		}
	}
	return false
}

// getBuildersFromImageStreams returns all the builder Images from the image streams provided and also hides the builder images
// which have hidden annotation attached to it
func getBuildersFromImageStreams(imageStreams []imagev1.ImageStream, imageStreamTagMap map[string]imagev1.ImageStreamTag) []Catalog {
	var builderImages []Catalog
	// Get builder images from the available imagestreams
	for _, imageStream := range imageStreams {
		var allTags []string
		var hiddenTags []string
		buildImage := false

		for _, tagReference := range imageStream.Spec.Tags {
			allTags = append(allTags, tagReference.Name)
			// Check to see if it is a "builder" image
			if _, ok := tagReference.Annotations["tags"]; ok {
				for _, t := range strings.Split(tagReference.Annotations["tags"], ",") {
					// If the tagReference has "builder" then we will add the image to the list
					if t == "builder" {
						buildImage = true
					}
				}
			}

		}

		// Append to the list of images if a "builder" tag was found
		if buildImage {
			// We need to gauge the ImageStreamTag of each potential builder image, because it might contain
			// the 'hidden' tag. If so, this builder image is deprecated and should not be offered to the user
			// as candidate
			for _, tag := range allTags {
				imageStreamTag := imageStreamTagMap[imageStream.Name+":"+tag]
				if _, ok := imageStreamTag.Annotations["tags"]; ok {
					for _, t := range strings.Split(imageStreamTag.Annotations["tags"], ",") {
						// If the tagReference has "builder" then we will add the image to the list
						if t == "hidden" {
							glog.V(5).Infof("Tag: %v of builder: %v is marked as hidden and therefore will be excluded", tag, imageStream.Name)
							hiddenTags = append(hiddenTags, tag)
						}
					}
				}

			}

			catalogImage := Catalog{
				TypeMeta: metav1.TypeMeta{
					Kind: "Catalog",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: imageStream.Name,
				},
				Spec: CatalogSpec{
					Namespace:      imageStream.Namespace,
					AllTags:        allTags,
					NonHiddenTags:  getAllNonHiddenTags(allTags, hiddenTags),
					ImageStreamRef: imageStream,
				},
			}
			builderImages = append(builderImages, catalogImage)
			glog.V(5).Infof("Found builder image: %#v", catalogImage)
		}

	}
	return builderImages
}

func getAllNonHiddenTags(allTags []string, hiddenTags []string) []string {
	result := make([]string, 0, len(allTags))
	for _, t1 := range allTags {
		found := false
		for _, t2 := range hiddenTags {
			if t1 == t2 {
				found = true
				break
			}
		}

		if !found {
			result = append(result, t1)
		}
	}
	return result
}
