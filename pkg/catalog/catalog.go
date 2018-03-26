package catalog

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/redhat-developer/ocdev/pkg/occlient"
	log "github.com/sirupsen/logrus"
)

// List lists all the available component types
func List(client *occlient.Client) ([]string, error) {
	var catalogList []string
	imageStreams, err := getDefaultBuilderImages(client)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get image streams")
	}
	catalogList = append(catalogList, imageStreams...)

	// TODO: uncomment when component create supports template creation
	//clusterServiceClasses, err := client.GetClusterServiceClassExternalNames()
	//if err != nil {
	//	return nil, errors.Wrap(err, "unable to get cluster service classes")
	//}
	//catalogList = append(catalogList, clusterServiceClasses...)

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
		if strings.Contains(component, name) {
			result = append(result, component)
		}
	}

	return result, nil
}

// getDefaultBuilderImages returns the default builder images available in the
// openshift namespace
func getDefaultBuilderImages(client *occlient.Client) ([]string, error) {
	imageStreams, err := client.GetImageStreams(occlient.OpenShiftNameSpace)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get Image Streams")
	}

	var builderImages []string
	// Get builder images from the available imagestreams
outer:
	for _, imageStream := range imageStreams {
		for _, tag := range imageStream.Spec.Tags {
			if _, ok := tag.Annotations["tags"]; ok {
				for _, t := range strings.Split(tag.Annotations["tags"], ",") {
					if t == "builder" {
						builderImages = append(builderImages, imageStream.Name)
						continue outer
					}
				}
			}
		}
	}
	log.Debugf("Found builder images: %v", builderImages)
	return builderImages, nil
}
