package catalog

import (
	"fmt"
	"strings"

	"github.com/golang/glog"
	imagev1 "github.com/openshift/api/image/v1"
	"github.com/pkg/errors"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/occlient"
)

type CatalogImage struct {
	Name      string
	Namespace string
	Tags      []string
}

// List lists all the available component types
func List(client *occlient.Client) ([]CatalogImage, error) {
	catalogList, err := getDefaultBuilderImages(client)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get image streams")
	}

	if len(catalogList) == 0 {
		return nil, errors.New("unable to retrieve any catalog images from the OpenShift cluster")
	}

	return catalogList, nil
}

// Search searches for the component
func Search(client *occlient.Client, name string) ([]string, error) {
	var result []string
	componentList, err := List(client)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list components")
	}

	// do a partial search in all the components
	for _, component := range componentList {
		if strings.Contains(component.Name, name) {
			result = append(result, component.Name)
		}
	}

	return result, nil
}

// Exists returns true if the given component type is valid, false if not
func Exists(client *occlient.Client, componentType string) (bool, error) {

	s := log.Spinner("Checking component")
	defer s.End(false)
	catalogList, err := List(client)
	if err != nil {
		return false, errors.Wrapf(err, "unable to list catalog")
	}

	for _, supported := range catalogList {
		if componentType == supported.Name || componentType == fmt.Sprintf("%s/%s", supported.Namespace, supported.Name) {
			s.End(true)
			return true, nil
		}
	}
	return false, nil
}

// VersionExists checks if that version exists, returns true if the given version exists, false if not
func VersionExists(client *occlient.Client, componentType string, componentVersion string) (bool, error) {

	// Loading status
	s := log.Spinner("Checking component version")
	defer s.End(false)

	// Retrieve the catalogList
	catalogList, err := List(client)
	if err != nil {
		return false, errors.Wrapf(err, "unable to list catalog")
	}

	// Find the component and then return true if the version has been found
	for _, supported := range catalogList {
		if componentType == supported.Name || componentType == fmt.Sprintf("%s/%s", supported.Namespace, supported.Name) {
			// Now check to see if that version matches that components tag
			for _, tag := range supported.Tags {
				if componentVersion == tag {
					s.End(true)
					return true, nil
				}
			}
		}
	}

	// Else return false if nothing is found
	return false, nil
}

// getDefaultBuilderImages returns the default builder images available in the
// openshift and the current namespaces
func getDefaultBuilderImages(client *occlient.Client) ([]CatalogImage, error) {

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
	var builderImages []CatalogImage

	// Get builder images from the available imagestreams
	for _, imageStream := range imageStreams {
		var allTags []string
		buildImage := false
		hidden := false
		for _, tag := range imageStream.Spec.Tags {

			allTags = append(allTags, tag.Name)

			// Check to see if it is a "builder" image
			if _, ok := tag.Annotations["tags"]; ok {
				for _, t := range strings.Split(tag.Annotations["tags"], ",") {
					// if there is a "hidden" tag then the image stream is deprecated
					if t == "hidden" {

						hidden = true
						break
					}
					// If the tag has "builder" then we will add the image to the list
					if t == "builder" {
						buildImage = true
					}
				}
			}

			if hidden {
				break
			}

		}
		// Append to the list of images if a "builder" tag was found
		if buildImage && !hidden {
			builderImages = append(builderImages, CatalogImage{Name: imageStream.Name, Namespace: imageStream.Namespace, Tags: allTags})
		}

	}

	glog.V(4).Infof("Found builder images: %v", builderImages)
	return builderImages, nil
}
