package catalog

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"sync"

	"github.com/openshift/odo/pkg/preference"
	"github.com/zalando/go-keyring"

	imagev1 "github.com/openshift/api/image/v1"
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/occlient"

	registryUtil "github.com/openshift/odo/pkg/odo/cli/registry/util"
	"github.com/openshift/odo/pkg/util"
	olm "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

const (
	apiVersion = "odo.dev/v1alpha1"
)

var supportedImages = map[string]bool{
	"redhat-openjdk-18/openjdk18-openshift:latest": true,
	"openjdk/openjdk-11-rhel8:latest":              true,
	"openjdk/openjdk-11-rhel7:latest":              true,
	"ubi8/openjdk-11:latest":                       true,
	"centos/nodejs-10-centos7:latest":              true,
	"centos/nodejs-12-centos7:latest":              true,
	"rhscl/nodejs-10-rhel7:latest":                 true,
	"rhscl/nodejs-12-rhel7:latest":                 true,
	"rhoar-nodejs/nodejs-10:latest":                true,
	"ubi8/nodejs-12:latest":                        true,
}

// GetDevfileRegistries gets devfile registries from preference file,
// if registry name is specified return the specific registry, otherwise return all registries
func GetDevfileRegistries(registryName string) ([]Registry, error) {
	var devfileRegistries []Registry

	cfg, err := preference.New()
	if err != nil {
		return nil, err
	}

	hasName := len(registryName) != 0
	if cfg.OdoSettings.RegistryList != nil {
		registryList := *cfg.OdoSettings.RegistryList
		// Loop backwards here to ensure the registry display order is correct (display latest newly added registry firstly)
		for i := len(registryList) - 1; i >= 0; i-- {
			registry := registryList[i]
			if hasName {
				if registryName == registry.Name {
					reg := Registry{
						Name:   registry.Name,
						URL:    registry.URL,
						Secure: registry.Secure,
					}
					devfileRegistries = append(devfileRegistries, reg)
					return devfileRegistries, nil
				}
			} else {
				reg := Registry{
					Name:   registry.Name,
					URL:    registry.URL,
					Secure: registry.Secure,
				}
				devfileRegistries = append(devfileRegistries, reg)
			}
		}
	} else {
		return nil, nil
	}

	return devfileRegistries, nil
}

// convertURL converts GitHub regular URL to GitHub raw URL, do nothing if the URL is not GitHub URL
// For example:
// GitHub regular URL: https://github.com/elsony/devfile-registry/tree/johnmcollier-crw
// GitHub raw URL: https://raw.githubusercontent.com/elsony/devfile-registry/johnmcollier-crw
func convertURL(URL string) (string, error) {
	url, err := url.Parse(URL)
	if err != nil {
		return "", err
	}

	if strings.Contains(url.Host, "github") && !strings.Contains(url.Host, "raw") {
		// Convert path part of the URL
		URLSlice := strings.Split(URL, "/")
		if len(URLSlice) > 2 && URLSlice[len(URLSlice)-2] == "tree" {
			// GitHub raw URL doesn't have "tree" structure in the URL, need to remove it
			URL = strings.Replace(URL, "/tree", "", 1)
		} else {
			// Add "master" branch for GitHub raw URL by default if branch is not specified
			URL = URL + "/master"
		}

		// Convert host part of the URL
		if url.Host == "github.com" {
			URL = strings.Replace(URL, "github.com", "raw.githubusercontent.com", 1)
		} else {
			URL = strings.Replace(URL, url.Host, "raw."+url.Host, 1)
		}
	}

	return URL, nil
}

const indexPath = "/devfiles/index.json"

// getRegistryDevfiles retrieves the registry's index devfile entries
func getRegistryDevfiles(registry Registry) ([]DevfileComponentType, error) {
	var devfileIndex []DevfileIndexEntry

	URL, err := convertURL(registry.URL)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to convert URL %s", registry.URL)
	}
	registry.URL = URL
	indexLink := registry.URL + indexPath
	request := util.HTTPRequestParams{
		URL: indexLink,
	}
	if registryUtil.IsSecure(registry.Name) {
		token, err := keyring.Get(fmt.Sprintf("%s%s", util.CredentialPrefix, registry.Name), registryUtil.RegistryUser)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to get secure registry credential from keyring")
		}
		request.Token = token
	}

	cfg, err := preference.New()
	if err != nil {
		return nil, err
	}

	jsonBytes, err := util.HTTPGetRequest(request, cfg.GetRegistryCacheTime())
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to download the devfile index.json from %s", indexLink)
	}

	err = json.Unmarshal(jsonBytes, &devfileIndex)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to unmarshal the devfile index.json from %s", indexLink)
	}

	var registryDevfiles []DevfileComponentType

	for _, devfileIndexEntry := range devfileIndex {
		stackDevfile := DevfileComponentType{
			Name:        devfileIndexEntry.Name,
			DisplayName: devfileIndexEntry.DisplayName,
			Description: devfileIndexEntry.Description,
			Link:        devfileIndexEntry.Links.Link,
			Registry:    registry,
		}
		registryDevfiles = append(registryDevfiles, stackDevfile)
	}

	return registryDevfiles, nil
}

// ListDevfileComponents lists all the available devfile components
func ListDevfileComponents(registryName string) (DevfileComponentTypeList, error) {
	catalogDevfileList := &DevfileComponentTypeList{}
	var err error

	// TODO: consider caching registry information for better performance since it should be fairly stable over time
	// Get devfile registries
	catalogDevfileList.DevfileRegistries, err = GetDevfileRegistries(registryName)
	if err != nil {
		return *catalogDevfileList, err
	}
	if catalogDevfileList.DevfileRegistries == nil {
		return *catalogDevfileList, nil
	}

	// first retrieve the indices for each registry, concurrently
	devfileIndicesMutex := &sync.Mutex{}
	retrieveRegistryIndices := util.NewConcurrentTasks(len(catalogDevfileList.DevfileRegistries))

	// The 2D slice index is the priority of the registry (highest priority has highest index)
	// and the element is the devfile slice that belongs to the registry
	registrySlice := make([][]DevfileComponentType, len(catalogDevfileList.DevfileRegistries))
	for regPriority, reg := range catalogDevfileList.DevfileRegistries {
		// Load the devfile registry index.json
		registry := reg                 // Needed to prevent the lambda from capturing the value
		registryPriority := regPriority // Needed to prevent the lambda from capturing the value
		retrieveRegistryIndices.Add(util.ConcurrentTask{ToRun: func(errChannel chan error) {
			registryDevfiles, err := getRegistryDevfiles(registry)
			if err != nil {
				log.Warningf("Registry %s is not set up properly with error: %v, please check the registry URL and credential (refer `odo registry update --help`)\n", registry.Name, err)
				return
			}

			devfileIndicesMutex.Lock()
			registrySlice[registryPriority] = registryDevfiles
			devfileIndicesMutex.Unlock()
		}})
	}
	if err := retrieveRegistryIndices.Run(); err != nil {
		return *catalogDevfileList, err
	}

	for _, registryDevfiles := range registrySlice {
		catalogDevfileList.Items = append(catalogDevfileList.Items, registryDevfiles...)
	}

	return *catalogDevfileList, nil
}

// ListComponents lists all the available component types
func ListComponents(client *occlient.Client) (ComponentTypeList, error) {

	catalogList, err := getDefaultBuilderImages(client)
	if err != nil {
		return ComponentTypeList{}, errors.Wrap(err, "unable to get image streams")
	}

	if len(catalogList) == 0 {
		return ComponentTypeList{}, errors.New("unable to retrieve any catalog images from the OpenShift cluster")
	}

	return ComponentTypeList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "List",
			APIVersion: apiVersion,
		},
		Items: catalogList,
	}, nil

}

// SearchComponent searches for the component
func SearchComponent(client *occlient.Client, name string) ([]string, error) {
	var result []string
	componentList, err := ListComponents(client)
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

// ComponentExists returns true if the given component type and the version are valid, false if not
func ComponentExists(client *occlient.Client, componentType string, componentVersion string) (bool, error) {
	imageStream, err := client.GetImageStream("", componentType, componentVersion)
	if err != nil {
		return false, errors.Wrapf(err, "unable to get from catalog")
	}
	if imageStream == nil {
		return false, nil
	}
	return true, nil
}

// ListSvcCatServices lists all the available services provided by Service Catalog
func ListSvcCatServices(client *occlient.Client) (ServiceTypeList, error) {

	clusterServiceClasses, err := getClusterCatalogServices(client)
	if err != nil {
		return ServiceTypeList{}, errors.Wrapf(err, "unable to get cluster serviceClassExternalName")
	}

	// Sorting service classes alphabetically
	// Reference: https://golang.org/pkg/sort/#example_Slice
	sort.Slice(clusterServiceClasses, func(i, j int) bool {
		return clusterServiceClasses[i].Name < clusterServiceClasses[j].Name
	})

	return ServiceTypeList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "List",
			APIVersion: apiVersion,
		},
		Items: clusterServiceClasses,
	}, nil
}

// ListOperatorServices fetches a list of Operators from the cluster and
// returns only those Operators which are successfully installed on the cluster
func ListOperatorServices(client *kclient.Client) (*olm.ClusterServiceVersionList, error) {
	var csvList olm.ClusterServiceVersionList

	allCsvs, err := client.ListClusterServiceVersions()
	if err != nil {
		return &csvList, err
	}

	// now let's filter only those csvs which are successfully installed
	csvList.TypeMeta = allCsvs.TypeMeta
	csvList.ListMeta = allCsvs.ListMeta
	for _, csv := range allCsvs.Items {
		if csv.Status.Phase == "Succeeded" {
			csvList.Items = append(csvList.Items, csv)
		}
	}

	return &csvList, nil
}

// SearchService searches for the services
func SearchService(client *occlient.Client, name string) (ServiceTypeList, error) {
	var result []ServiceType
	serviceList, err := ListSvcCatServices(client)
	if err != nil {
		return ServiceTypeList{}, errors.Wrap(err, "unable to list services")
	}

	// do a partial search in all the services
	for _, service := range serviceList.Items {
		if strings.Contains(service.ObjectMeta.Name, name) {
			result = append(result, service)
		}
	}

	return ServiceTypeList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "List",
			APIVersion: apiVersion,
		},
		Items: result,
	}, nil
}

// getClusterCatalogServices returns the names of all the cluster service
// classes in the cluster
func getClusterCatalogServices(client *occlient.Client) ([]ServiceType, error) {
	var classNames []ServiceType

	classes, err := client.GetKubeClient().ListClusterServiceClasses()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get cluster service classes")
	}

	planListItems, err := client.GetKubeClient().ListClusterServicePlans()
	if err != nil {
		return nil, errors.Wrap(err, "Unable to get service plans")
	}
	for _, class := range classes {

		var planList []string
		for _, plan := range planListItems {
			if plan.Spec.ClusterServiceClassRef.Name == class.Spec.ExternalID {
				planList = append(planList, plan.Spec.ExternalName)
			}
		}
		classNames = append(classNames, ServiceType{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ServiceType",
				APIVersion: apiVersion,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: class.Spec.ExternalName,
			},
			Spec: ServiceSpec{
				Hidden:   occlient.HasTag(class.Spec.Tags, "hidden"),
				PlanList: planList,
			},
		})
	}
	return classNames, nil
}

// getDefaultBuilderImages returns the default builder images available in the
// openshift and the current namespaces
func getDefaultBuilderImages(client *occlient.Client) ([]ComponentType, error) {

	var imageStreams []imagev1.ImageStream
	currentNamespace := client.GetCurrentProjectName()

	// Fetch imagestreams from default openshift namespace
	openshiftNSImageStreams, openshiftNSISFetchError := client.ListImageStreams(occlient.OpenShiftNameSpace)
	if openshiftNSISFetchError != nil {
		// Tolerate the error as it might only be a partial failure
		// We may get the imagestreams from other Namespaces
		//err = errors.Wrapf(openshiftNSISFetchError, "unable to get Image Streams from namespace %s", occlient.OpenShiftNameSpace)
		// log it for debugging purposes
		klog.V(4).Infof("Unable to get Image Streams from namespace %s. Error %s", occlient.OpenShiftNameSpace, openshiftNSISFetchError.Error())
	}

	// Fetch imagestreams from current namespace
	currentNSImageStreams, currentNSISFetchError := client.ListImageStreams(currentNamespace)
	// If failure to fetch imagestreams from current namespace, log the failure for debugging purposes
	if currentNSISFetchError != nil {
		// Tolerate the error as it is totally a valid scenario to not have any imagestreams in current namespace
		// log it for debugging purposes
		klog.V(4).Infof("Unable to get Image Streams from namespace %s. Error %s", currentNamespace, currentNSISFetchError.Error())
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
func SliceSupportedTags(component ComponentType) ([]string, []string) {

	// this makes sure that json marshal shows these lists as [] instead of null
	supTag, unSupTag := []string{}, []string{}
	tagMap := createImageTagMap(component.Spec.ImageStreamTags)

	for _, tag := range component.Spec.NonHiddenTags {
		imageName := tagMap[tag]
		if isSupportedImage(imageName) {
			supTag = append(supTag, tag)
		} else {
			unSupTag = append(unSupTag, tag)
		}
	}
	return supTag, unSupTag
}

// IsComponentTypeSupported takes the componentType e.g. java:8 and return true if
// it is fully supported i.e. debug mode and more.
func IsComponentTypeSupported(client *occlient.Client, componentType string) (bool, error) {
	_, componentType, _, componentVersion := util.ParseComponentImageName(componentType)

	imageStream, err := client.GetImageStream("", componentType, componentVersion)
	if err != nil {
		return false, err
	}
	tagMap := createImageTagMap(imageStream.Spec.Tags)

	return isSupportedImage(tagMap[componentVersion]), nil
}

// createImageTagMap takes a list of image TagReferences and creates a map of type tag name => image name e.g. 1.11 => openshift/nodejs-11
func createImageTagMap(tagRefs []imagev1.TagReference) map[string]string {
	tagMap := make(map[string]string)
	for _, tagRef := range tagRefs {
		imageName := tagRef.From.Name
		if tagRef.From.Kind == "DockerImage" {
			// we get the image name from the repo url e.g. registry.redhat.com/openshift/nodejs:10 will give openshift/nodejs:10
			imageNameParts := strings.SplitN(imageName, "/", 2)

			var urlImageName string
			// this means the docker image url might just be something like nodejs:10, no namespace or registry info
			if len(imageNameParts) == 1 {
				urlImageName = imageNameParts[0]
				// else block executes when there is a registry information attached in the docker image url
			} else {
				// we dont want the registry url portion
				urlImageName = imageNameParts[1]
			}
			// here we remove the tag and digest
			ns, img, tag, _, _ := occlient.ParseImageName(urlImageName)
			imageName = ns + "/" + img + ":" + tag
			tagMap[tagRef.Name] = imageName
		}
	}

	for _, tagRef := range tagRefs {
		if tagRef.From.Kind == "ImageStreamTag" {
			imageName := tagRef.From.Name
			tagList := strings.Split(imageName, ":")
			tag := tagList[len(tagList)-1]
			// if the kind is a image stream tag that means its pointing to an existing dockerImage or image stream image
			// we just look it up from the tapMap we already have
			imageName = tagMap[tag]
			tagMap[tagRef.Name] = imageName
		}

	}
	return tagMap
}

// isSupportedImages returns if the image is supported or not. the supported images have been provided here
// https://github.com/openshift/odo-init-image/blob/master/language-scripts/image-mappings.json
func isSupportedImage(imgName string) bool {
	return supportedImages[imgName]
}

// getBuildersFromImageStreams returns all the builder Images from the image streams provided and also hides the builder images
// which have hidden annotation attached to it
func getBuildersFromImageStreams(imageStreams []imagev1.ImageStream, imageStreamTagMap map[string]imagev1.ImageStreamTag) []ComponentType {
	var builderImages []ComponentType
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
							klog.V(5).Infof("Tag: %v of builder: %v is marked as hidden and therefore will be excluded", tag, imageStream.Name)
							hiddenTags = append(hiddenTags, tag)
						}
					}
				}

			}

			catalogImage := ComponentType{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ComponentType",
					APIVersion: apiVersion,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      imageStream.Name,
					Namespace: imageStream.Namespace,
				},
				Spec: ComponentSpec{
					AllTags:         allTags,
					NonHiddenTags:   getAllNonHiddenTags(allTags, hiddenTags),
					ImageStreamTags: imageStream.Spec.Tags,
				},
			}
			builderImages = append(builderImages, catalogImage)
			klog.V(5).Infof("Found builder image: %#v", catalogImage)
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
