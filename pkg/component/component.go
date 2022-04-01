package component

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/api/v2/pkg/devfile"
	"github.com/devfile/library/pkg/devfile/parser"
	dfutil "github.com/devfile/library/pkg/util"

	odolabels "github.com/redhat-developer/odo/pkg/component/labels"
	"github.com/redhat-developer/odo/pkg/kclient"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog"
)

const NotAvailable = "Not available"

// GetComponentTypeFromDevfileMetadata returns component type from the devfile metadata;
// it could either be projectType or language, if neither of them are set, return 'Not available'
func GetComponentTypeFromDevfileMetadata(metadata devfile.DevfileMetadata) string {
	var componentType string
	if metadata.ProjectType != "" {
		componentType = metadata.ProjectType
	} else if metadata.Language != "" {
		componentType = metadata.Language
	} else {
		componentType = NotAvailable
	}
	return componentType
}

// GatherName parses the Devfile and retrieves an appropriate name in two ways.
// 1. If metadata.name exists, we use it
// 2. If metadata.name does NOT exist, we use the folder name where the devfile.yaml is located
func GatherName(devObj parser.DevfileObj, devfilePath string) (string, error) {

	metadata := devObj.Data.GetMetadata()

	klog.V(4).Infof("metadata.Name: %s", metadata.Name)

	// 1. Use metadata.name if it exists
	if metadata.Name != "" {

		// Remove any suffix's that end with `-`. This is because many Devfile's use the original v1 Devfile pattern of
		// having names such as "foo-bar-" in order to prepend container names such as "foo-bar-container1"
		return strings.TrimSuffix(metadata.Name, "-"), nil
	}

	// 2. Use the folder name as a last resort if nothing else exists
	sourcePath, err := dfutil.GetAbsPath(devfilePath)
	if err != nil {
		return "", fmt.Errorf("unable to get source path: %w", err)
	}
	klog.V(4).Infof("Source path: %s", sourcePath)
	klog.V(4).Infof("devfile dir: %s", filepath.Dir(sourcePath))

	return filepath.Base(filepath.Dir(sourcePath)), nil
}

// Exists checks whether a component with the given name exists in the current application or not
// componentName is the component name to perform check for
// The first returned parameter is a bool indicating if a component with the given name already exists or not
// The second returned parameter is the error that might occurs while execution
func Exists(client kclient.ClientInterface, componentName, applicationName string) (bool, error) {
	deploymentName, err := dfutil.NamespaceOpenShiftObject(componentName, applicationName)
	if err != nil {
		return false, fmt.Errorf("unable to create namespaced name: %w", err)
	}
	deployment, _ := client.GetDeploymentByName(deploymentName)
	if deployment != nil {
		return true, nil
	}
	return false, nil
}

// GetOnePod gets a pod using the component and app name
func GetOnePod(client kclient.ClientInterface, componentName string, appName string) (*corev1.Pod, error) {
	return client.GetOnePodFromSelector(odolabels.GetSelector(componentName, appName, odolabels.ComponentDevMode))
}

// ComponentExists checks whether a deployment by the given name exists in the given app
func ComponentExists(client kclient.ClientInterface, name string, app string) (bool, error) {
	deployment, err := client.GetOneDeployment(name, app)
	if _, ok := err.(*kclient.DeploymentNotFoundError); ok {
		klog.V(2).Infof("Deployment %s not found for belonging to the %s app ", name, app)
		return false, nil
	}
	return deployment != nil, err
}

// Log returns log from component
func Log(client kclient.ClientInterface, componentName string, appName string, follow bool, command v1alpha2.Command) (io.ReadCloser, error) {

	pod, err := GetOnePod(client, componentName, appName)
	if err != nil {
		return nil, fmt.Errorf("the component %s doesn't exist on the cluster", componentName)
	}

	if pod.Status.Phase != corev1.PodRunning {
		return nil, fmt.Errorf("unable to show logs, component is not in running state. current status=%v", pod.Status.Phase)
	}

	containerName := command.Exec.Component

	return client.GetPodLogs(pod.Name, containerName, follow)
}

// ListAllClusterComponents returns a list of all "components" on a cluster
// that are both odo and non-odo components.
//
// We then return a list of "components" intended for listing / output purposes specifically for commands such as:
// `odo list`
// that are both odo and non-odo components.
//
// We then return a list of "components" intended for listing / output purposes specifically for commands such as:
// `odo list`
func ListAllClusterComponents(client kclient.ClientInterface, namespace string) ([]OdoComponent, error) {

	// Get all the dynamic resources available
	resourceList, err := client.GetAllResourcesFromSelector("", namespace)
	if err != nil {
		return nil, fmt.Errorf("unable to list all dynamic resources required to find components: %w", err)
	}

	var components []OdoComponent

	for _, resource := range resourceList {

		// ignore "PackageManifest" as they are not components, it is just a record in OpenShift catalog.
		if resource.GetKind() == "PackageManifest" {
			continue
		}

		var labels, annotations map[string]string

		// Retrieve the labels and annotations from the unstructured resource output
		if resource.GetLabels() != nil {
			labels = resource.GetLabels()
		}
		if resource.GetAnnotations() != nil {
			annotations = resource.GetAnnotations()
		}

		// Figure out the correct name to use
		// if there is no instance label, we SKIP the resource as
		// it is not a component essential for Kubernetes.
		name := odolabels.GetComponentName(labels)
		if name == "" {
			continue
		}

		// Get the component type (if there is any..)
		componentType, err := odolabels.GetProjectType(nil, annotations)
		if err != nil || componentType == "" {
			componentType = StateTypeUnknown
		}

		// Get the managedBy label
		// IMPORTANT. If "managed-by" label is BLANK, it is most likely an operator
		// or a non-component. We do not want to show these in the list of components
		// so we skip them if there is no "managed-by" label.

		managedBy := odolabels.GetManagedBy(labels)
		if managedBy == "" {
			continue
		}

		// Generate the appropriate "component" with all necessary information
		component := OdoComponent{
			Name:      name,
			ManagedBy: managedBy,
			Type:      componentType,
		}
		mode := odolabels.GetMode(labels)
		found := false
		for v, otherCompo := range components {
			if component.Name == otherCompo.Name {
				found = true
				if mode != "" {
					components[v].Modes[mode] = true
				}
				if otherCompo.Type == StateTypeUnknown && component.Type != StateTypeUnknown {
					components[v].Type = component.Type
				}
				if otherCompo.ManagedBy == StateTypeUnknown && component.ManagedBy != StateTypeUnknown {
					components[v].ManagedBy = component.ManagedBy
				}
			}
		}
		if !found {
			if mode != "" {
				component.Modes = map[string]bool{
					mode: true,
				}
			} else {
				component.Modes = map[string]bool{}
			}
			components = append(components, component)
		}
	}

	return components, nil
}

// Contains checks to see if the component exists in an array or not
// by checking the name
func Contains(component OdoComponent, components []OdoComponent) bool {
	for _, comp := range components {
		if component.Name == comp.Name {
			return true
		}
	}
	return false
}
