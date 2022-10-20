package backend

import (
	"fmt"
	"sort"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog"

	"github.com/redhat-developer/odo/pkg/binding/asker"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/project"
)

type selectWorkloadStep int

const (
	step_select_kind selectWorkloadStep = iota
	step_select_name
	step_selected
)

// InteractiveBackend is a backend that will ask information interactively using the `asker` package
type InteractiveBackend struct {
	askerClient      asker.Asker
	projectClient    project.Client
	kubernetesClient kclient.ClientInterface
}

var _ AddBindingBackend = (*InteractiveBackend)(nil)

func NewInteractiveBackend(askerClient asker.Asker, projectClient project.Client, kubernetesClient kclient.ClientInterface) *InteractiveBackend {
	return &InteractiveBackend{
		askerClient:      askerClient,
		projectClient:    projectClient,
		kubernetesClient: kubernetesClient,
	}
}

func (o *InteractiveBackend) Validate(_ map[string]string, _ bool) error {
	return nil
}

func (o *InteractiveBackend) SelectWorkloadInstance(_ string) (string, schema.GroupVersionKind, error) {

	step := step_select_kind
	var selectedGVK schema.GroupVersionKind
	var selectedName string
loop:
	for {
		switch step {
		case step_select_kind:
			options, allWorkloadsKinds, err := o.kubernetesClient.GetWorkloadKinds()
			if err != nil {
				return "", schema.GroupVersionKind{}, err
			}
			i, err := o.askerClient.SelectWorkloadResource(options)
			if err != nil {
				return "", schema.GroupVersionKind{}, err
			}
			selectedGVK = allWorkloadsKinds[i]
			step++

		case step_select_name:
			// Get the resources of this kind
			gvr, err := o.kubernetesClient.GetGVRFromGVK(selectedGVK)
			if err != nil {
				return "", schema.GroupVersionKind{}, err
			}
			resourceList, err := o.kubernetesClient.ListDynamicResources("", gvr, "")
			if err != nil {
				return "", schema.GroupVersionKind{}, err
			}

			// Ask to select the name of the resource
			names := make([]string, 0, len(resourceList.Items))
			for _, resource := range resourceList.Items {
				names = append(names, resource.GetName())
			}
			var back bool
			back, selectedName, err = o.askerClient.SelectWorkloadResourceName(names)
			if err != nil {
				return "", schema.GroupVersionKind{}, err
			}
			if back {
				step--
			} else {
				step++
			}

		case step_selected:
			break loop
		}
	}

	// Ask the name if DOES NOT EXIST is selected
	var err error
	if selectedName == "" {
		selectedName, err = o.askerClient.AskWorkloadResourceName()
		if err != nil {
			return "", schema.GroupVersionKind{}, err
		}
	}
	return selectedName, selectedGVK, nil
}

// SelectNamespace prompts users to select the namespace which services instances should be listed from.
// If they choose all the namespaces they have access to, it attempts to get the list of accessible namespaces in the cluster,
// from which the user can select one.
// If the list is empty (e.g. because of permission-related issues), the user is prompted to manually provide a namespace.
func (o *InteractiveBackend) SelectNamespace(_ map[string]string) (string, error) {
	option, err := o.askerClient.SelectNamespaceListOption()
	if err != nil {
		return "", err
	}

	switch option {
	case asker.CurrentNamespace:
		return "", nil
	case asker.AllAccessibleNamespaces:
		klog.V(2).Infof("Listing all projects/namespaces...")
		var nsList []string
		nsList, err = o.getAllNamespaces()
		if err != nil {
			return "", err
		}
		sort.Strings(nsList)
		klog.V(4).Infof("all accessible namespaces: %v", nsList)
		if len(nsList) == 0 {
			// User needs to provide a namespace
			return o.askerClient.AskNamespace()
		}
		// Let users select a namespace from the list
		return o.askerClient.SelectNamespace(nsList)
	default:
		return "", fmt.Errorf("unknown namespace list option: %d", option)
	}
}

func (o *InteractiveBackend) SelectServiceInstance(_ string, serviceMap map[string]unstructured.Unstructured) (string, error) {
	var options []string
	for name := range serviceMap {
		options = append(options, name)
	}
	return o.askerClient.AskServiceInstance(options)
}

func (o *InteractiveBackend) AskBindingName(defaultName string, _ map[string]string) (string, error) {
	return o.askerClient.AskServiceBindingName(defaultName)
}

func (o *InteractiveBackend) AskBindAsFiles(_ map[string]string) (bool, error) {
	return o.askerClient.AskBindAsFiles()
}

func (o *InteractiveBackend) AskNamingStrategy(_ map[string]string) (string, error) {
	namingStrategy, err := o.askerClient.SelectNamingStrategy()
	if err != nil {
		return "", err
	}
	if namingStrategy == asker.NamingStrategyCustom {
		return o.askerClient.AskNamingStrategy()
	}
	return namingStrategy, nil
}

func (o *InteractiveBackend) SelectCreationOptions(flags map[string]string) ([]asker.CreationOption, error) {
	return o.askerClient.SelectCreationOptions()
}

func (o *InteractiveBackend) AskOutputFilePath(flags map[string]string, defaultValue string) (string, error) {
	return o.askerClient.AskOutputFilePath(defaultValue)
}

func (o *InteractiveBackend) getAllNamespaces() ([]string, error) {
	accessibleNsList, err := o.projectClient.List()
	if err != nil {
		klog.V(2).Infof("Failed to list namespaces/projects: %v", err)
		if kerrors.IsForbidden(err) {
			// If status is forbidden, this might be an RBAC error due to user not having permission to list namespaces.
			// In this case, user will need to manually specify the namespace.
			log.Warningf("Failed to list namespaces/projects: %v", err)
			return nil, nil
		}
		return nil, err
	}
	nsList := make([]string, 0, len(accessibleNsList.Items))
	for _, ns := range accessibleNsList.Items {
		nsList = append(nsList, ns.Name)
	}
	return nsList, nil
}
