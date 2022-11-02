package component

import (
	"errors"
	"fmt"
	"strings"

	devfile "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	devfilefs "github.com/devfile/library/pkg/testingutil/filesystem"

	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/libdevfile"
)

// ValidateResourcesExist validates if the Kubernetes inlined components are installed on the cluster
func ValidateResourcesExist(client kclient.ClientInterface, devfileObj parser.DevfileObj, k8sComponents []devfile.Component, context string) error {
	if len(k8sComponents) == 0 {
		return nil
	}

	var unsupportedResources []string
	for _, c := range k8sComponents {
		kindErr, err := ValidateResourceExist(client, devfileObj, c, context)
		if err != nil {
			if kindErr != "" {
				unsupportedResources = append(unsupportedResources, kindErr)
			} else {
				return err
			}
		}
	}

	if len(unsupportedResources) > 0 {
		// tell the user about all the unsupported resources in one message
		return fmt.Errorf("following resource(s) in the devfile are not supported by your cluster; please install corresponding Operator(s) before doing \"odo dev\": %s", strings.Join(unsupportedResources, ", "))
	}
	return nil
}

// ValidateResourceExist validates if a Kubernetes inlined component is installed on the cluster
func ValidateResourceExist(client kclient.ClientInterface, devfileObj parser.DevfileObj, k8sComponent devfile.Component, context string) (kindErr string, err error) {
	// get the string representation of the YAML definition of a CRD
	u, err := libdevfile.GetK8sComponentAsUnstructured(devfileObj, k8sComponent.Name, context, devfilefs.DefaultFs{})
	if err != nil {
		return "", err
	}

	_, err = client.GetRestMappingFromUnstructured(u)
	if err != nil && u.GetKind() != "ServiceBinding" {
		// getting a RestMapping would fail if there are no matches for the Kind field on the cluster;
		// but if it's a "ServiceBinding" resource, we don't add it to unsupported list because odo can create links
		// without having SBO installed
		return u.GetKind(), errors.New("resource not supported")
	}
	return "", nil
}
