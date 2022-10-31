package component

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/api/v2/pkg/devfile"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/devfile/parser/data"
	"k8s.io/klog"

	"github.com/redhat-developer/odo/pkg/alizer"
	"github.com/redhat-developer/odo/pkg/api"
	"github.com/redhat-developer/odo/pkg/kclient"
	odolabels "github.com/redhat-developer/odo/pkg/labels"
	odocontext "github.com/redhat-developer/odo/pkg/odo/context"
	"github.com/redhat-developer/odo/pkg/platform"
	"github.com/redhat-developer/odo/pkg/util"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	NotAvailable = "Not available"
	UnknownValue = "Unknown"
)

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

// GetComponentRuntimeFromDevfileMetadata returns the Project Type defined in the Devfile metadata
// or if not set, the Language
func GetComponentRuntimeFromDevfileMetadata(metadata devfile.DevfileMetadata) string {
	if metadata.ProjectType != "" {
		return metadata.ProjectType
	}
	return metadata.Language
}

// GatherName computes and returns what should be used as name for the Devfile object specified.
//
// If a non-blank name is available in the Devfile metadata (which is optional), it is sanitized and returned.
//
// Otherwise, it uses Alizer to detect the name, from the project build tools (pom.xml, package.json, ...),
// or from the component directory name.
func GatherName(contextDir string, devfileObj *parser.DevfileObj) (string, error) {
	var name string
	if devfileObj != nil {
		name = devfileObj.GetMetadataName()
		if name == "" || strings.TrimSpace(name) == "" {
			// Use Alizer if Devfile has no (optional) metadata.name field.
			// We need to pass in the Devfile base directory (not the path to the devfile.yaml).
			// Name returned by alizer.DetectName is expected to be already sanitized.
			alizerClient := alizer.Alizer{} // TODO(feloy) fix with DI
			return alizerClient.DetectName(filepath.Dir(devfileObj.Ctx.GetAbsPath()))
		}
	} else {
		// Fallback to the context dir name
		baseDir, err := filepath.Abs(contextDir)
		if err != nil {
			return "", err
		}
		name = filepath.Base(baseDir)
	}

	//sanitize the name
	s := util.GetDNS1123Name(name)
	klog.V(3).Infof("name of component is %q, and sanitized name is %q", name, s)

	return s, nil
}

// Log returns log from component
func Log(platformClient platform.Client, componentName string, appName string, follow bool, command v1alpha2.Command) (io.ReadCloser, error) {

	pod, err := platformClient.GetRunningPodFromSelector(odolabels.GetSelector(componentName, appName, odolabels.ComponentDevMode, false))
	if err != nil {
		return nil, fmt.Errorf("the component %s doesn't exist on the cluster", componentName)
	}

	if pod.Status.Phase != corev1.PodRunning {
		return nil, fmt.Errorf("unable to show logs, component is not in running state. current status=%v", pod.Status.Phase)
	}

	containerName := command.Exec.Component

	return platformClient.GetPodLogs(pod.Name, containerName, follow)
}

// ListAllClusterComponents returns a list of all "components" on a cluster
// that are both odo and non-odo components.
//
// We then return a list of "components" intended for listing / output purposes specifically for commands such as:
// `odo list`
// that are both odo and non-odo components.
func ListAllClusterComponents(client kclient.ClientInterface, namespace string) ([]api.ComponentAbstract, error) {

	// Get all the dynamic resources available
	resourceList, err := client.GetAllResourcesFromSelector("", namespace)
	if err != nil {
		return nil, fmt.Errorf("unable to list all dynamic resources required to find components: %w", err)
	}

	var components []api.ComponentAbstract

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
		// if there is no instance label (app.kubernetes.io/instance),
		// we SKIP the resource as it is not a component essential for Kubernetes.
		name := odolabels.GetComponentName(labels)
		if name == "" {
			continue
		}

		// Get the component type (if there is any..)
		componentType, err := odolabels.GetProjectType(nil, annotations)
		if err != nil || componentType == "" {
			componentType = api.TypeUnknown
		}

		// Get the managedBy label
		// IMPORTANT. If both "managed-by" and "instance" labels are BLANK, it is most likely an operator
		// or a non-component. We do not want to show these in the list of components
		// so we skip them if there is no "managed-by" label.

		managedBy := odolabels.GetManagedBy(labels)
		if managedBy == "" && name == "" {
			continue
		}

		managedByVersion := odolabels.GetManagedByVersion(labels)

		// Generate the appropriate "component" with all necessary information
		component := api.ComponentAbstract{
			Name:             name,
			ManagedBy:        managedBy,
			Type:             componentType,
			ManagedByVersion: managedByVersion,
		}
		mode := odolabels.GetMode(labels)
		componentFound := false
		for v, otherCompo := range components {
			if component.Name == otherCompo.Name {
				componentFound = true
				if mode != "" {
					if components[v].RunningIn == nil {
						components[v].RunningIn = api.NewRunningModes()
					}
					components[v].RunningIn.AddRunningMode(api.RunningMode(strings.ToLower(mode)))
				}
				if otherCompo.Type == api.TypeUnknown && component.Type != api.TypeUnknown {
					components[v].Type = component.Type
				}
				if otherCompo.ManagedBy == api.TypeUnknown && component.ManagedBy != api.TypeUnknown {
					components[v].ManagedBy = component.ManagedBy
				}
			}
		}
		if !componentFound {
			if mode != "" {
				if component.RunningIn == nil {
					component.RunningIn = api.NewRunningModes()
				}
				component.RunningIn.AddRunningMode(api.RunningMode(strings.ToLower(mode)))
			}
			components = append(components, component)
		}
	}

	return components, nil
}

func ListAllComponents(client kclient.ClientInterface, namespace string, devObj *parser.DevfileObj, componentName string) ([]api.ComponentAbstract, string, error) {
	devfileComponents, err := ListAllClusterComponents(client, namespace)
	if err != nil {
		return nil, "", err
	}

	localComponent := api.ComponentAbstract{
		Name:      componentName,
		ManagedBy: "",
		RunningIn: api.NewRunningModes(),
	}
	if devObj != nil {
		localComponent.Type = GetComponentTypeFromDevfileMetadata(devObj.Data.GetMetadata())
	}

	componentInDevfile := ""
	if localComponent.Name != "" {
		if !Contains(localComponent, devfileComponents) {
			devfileComponents = append(devfileComponents, localComponent)
		}
		componentInDevfile = localComponent.Name
	}
	return devfileComponents, componentInDevfile, nil
}

func getResourcesForComponent(
	ctx context.Context,
	client kclient.ClientInterface,
	name string,
	namespace string,
) ([]unstructured.Unstructured, error) {
	selector := odolabels.GetSelector(name, odocontext.GetApplication(ctx), odolabels.ComponentAnyMode, false)
	resourceList, err := client.GetAllResourcesFromSelector(selector, namespace)
	if err != nil {
		return nil, err
	}
	filteredList := []unstructured.Unstructured{}
	for _, resource := range resourceList {
		// ignore "PackageManifest" as they are not components, it is just a record in OpenShift catalog.
		if resource.GetKind() == "PackageManifest" {
			continue
		}
		filteredList = append(filteredList, resource)
	}
	return filteredList, nil
}

// GetRunningModes returns the list of modes on which a "name" component is deployed, by looking into namespace
// the resources deployed with matching labels, based on the "odo.dev/mode" label
func GetRunningModes(ctx context.Context, client kclient.ClientInterface, name string) (api.RunningModes, error) {
	list, err := getResourcesForComponent(ctx, client, name, client.GetCurrentNamespace())
	if err != nil {
		return api.RunningModes{}, nil
	}

	if len(list) == 0 {
		return nil, NewNoComponentFoundError(name, client.GetCurrentNamespace())
	}

	mapResult := api.NewRunningModes()
	for _, resource := range list {
		resourceLabels := resource.GetLabels()
		mode := odolabels.GetMode(resourceLabels)
		if mode != "" {
			mapResult.AddRunningMode(api.RunningMode(strings.ToLower(mode)))
		}
	}
	return mapResult, nil
}

// Contains checks to see if the component exists in an array or not
// by checking the name
func Contains(component api.ComponentAbstract, components []api.ComponentAbstract) bool {
	for _, comp := range components {
		if component.Name == comp.Name {
			return true
		}
	}
	return false
}

// GetDevfileInfoFromCluster extracts information from the labels and annotations of resources to rebuild a Devfile
func GetDevfileInfoFromCluster(ctx context.Context, client kclient.ClientInterface, name string) (parser.DevfileObj, error) {
	list, err := getResourcesForComponent(ctx, client, name, client.GetCurrentNamespace())
	if err != nil {
		return parser.DevfileObj{}, nil
	}

	if len(list) == 0 {
		return parser.DevfileObj{}, nil
	}

	devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
	if err != nil {
		return parser.DevfileObj{}, err
	}
	metadata := devfileData.GetMetadata()
	metadata.Name = UnknownValue
	metadata.DisplayName = UnknownValue
	metadata.ProjectType = UnknownValue
	metadata.Language = UnknownValue
	metadata.Version = UnknownValue
	metadata.Description = UnknownValue

	for _, resource := range list {
		labels := resource.GetLabels()
		annotations := resource.GetAnnotations()
		name := odolabels.GetComponentName(labels)
		if len(name) > 0 && metadata.Name == UnknownValue {
			metadata.Name = name
		}
		typ, err := odolabels.GetProjectType(labels, annotations)
		if err != nil {
			continue
		}
		if len(typ) > 0 && metadata.ProjectType == UnknownValue {
			metadata.ProjectType = typ
		}
	}
	devfileData.SetMetadata(metadata)
	return parser.DevfileObj{
		Data: devfileData,
	}, nil
}
