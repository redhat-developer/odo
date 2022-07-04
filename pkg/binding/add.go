package binding

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/devfile/library/pkg/devfile/parser"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	sboApi "github.com/redhat-developer/service-binding-operator/apis/binding/v1alpha1"

	"github.com/redhat-developer/odo/pkg/binding/asker"
	backendpkg "github.com/redhat-developer/odo/pkg/binding/backend"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/libdevfile"
)

// ValidateAddBinding calls Validate method of the adequate backend
func (o *BindingClient) ValidateAddBinding(flags map[string]string, withDevfile bool) error {
	var backend backendpkg.AddBindingBackend
	if len(flags) == 0 {
		backend = o.interactiveBackend
	} else {
		backend = o.flagsBackend
	}
	return backend.Validate(flags, withDevfile)
}

func (o *BindingClient) SelectServiceInstance(flags map[string]string, serviceMap map[string]unstructured.Unstructured) (string, error) {
	var backend backendpkg.AddBindingBackend
	if len(flags) == 0 {
		backend = o.interactiveBackend
	} else {
		backend = o.flagsBackend
	}
	return backend.SelectServiceInstance(flags[backendpkg.FLAG_SERVICE], serviceMap)
}

func (o *BindingClient) SelectWorkloadInstance(flags map[string]string) (string, schema.GroupVersionKind, error) {
	var backend backendpkg.AddBindingBackend
	if len(flags) == 0 {
		backend = o.interactiveBackend
	} else {
		backend = o.flagsBackend
	}
	resource, gvk, err := backend.SelectWorkloadInstance(flags[backendpkg.FLAG_WORKLOAD])
	if err != nil {
		return "", schema.GroupVersionKind{}, err
	}
	return resource, gvk, nil
}

func (o *BindingClient) AskBindingName(serviceName, componentName string, flags map[string]string) (string, error) {
	var backend backendpkg.AddBindingBackend
	if len(flags) == 0 {
		backend = o.interactiveBackend
	} else {
		backend = o.flagsBackend
	}
	defaultBindingName := fmt.Sprintf("%v-%v", componentName, serviceName)
	return backend.AskBindingName(defaultBindingName, flags)
}

func (o *BindingClient) AskBindAsFiles(flags map[string]string) (bool, error) {
	var backend backendpkg.AddBindingBackend
	if len(flags) == 0 {
		backend = o.interactiveBackend
	} else {
		backend = o.flagsBackend
	}
	return backend.AskBindAsFiles(flags)
}

func (o *BindingClient) AskNamingStrategy(flags map[string]string) (string, error) {
	var backend backendpkg.AddBindingBackend
	if len(flags) == 0 {
		backend = o.interactiveBackend
	} else {
		backend = o.flagsBackend
	}
	return backend.AskNamingStrategy(flags)
}

func (o *BindingClient) AddBindingToDevfile(
	bindingName string,
	bindAsFiles bool,
	namingStrategy string,
	unstructuredService unstructured.Unstructured,
	obj parser.DevfileObj,
) (parser.DevfileObj, error) {
	service, err := o.kubernetesClient.NewServiceBindingServiceObject(unstructuredService, bindingName)
	if err != nil {
		return obj, err
	}

	deploymentName := fmt.Sprintf("%s-app", obj.GetMetadataName())
	deploymentGVK, err := o.kubernetesClient.GetDeploymentAPIVersion()
	if err != nil {
		return obj, err
	}

	serviceBinding := kclient.NewServiceBindingObject(
		bindingName, bindAsFiles, deploymentName, namingStrategy, deploymentGVK, []sboApi.Mapping{}, []sboApi.Service{service}, sboApi.ServiceBindingStatus{})

	// Note: we cannot directly marshal the serviceBinding object to yaml because it doesn't do that in the correct k8s manifest format
	serviceBindingUnstructured, err := kclient.ConvertK8sResourceToUnstructured(serviceBinding)
	if err != nil {
		return obj, err
	}
	yamlDesc, err := yaml.Marshal(serviceBindingUnstructured.UnstructuredContent())
	if err != nil {
		return obj, err
	}

	return libdevfile.AddKubernetesComponentToDevfile(string(yamlDesc), serviceBinding.Name, obj)
}

func (o *BindingClient) AddBinding(
	flags map[string]string,
	bindingName string,
	bindAsFiles bool,
	namingStrategy string,
	unstructuredService unstructured.Unstructured,
	workloadName string,
	workloadGVK schema.GroupVersionKind,
) ([]asker.CreationOption, string, string, error) {
	service, err := o.kubernetesClient.NewServiceBindingServiceObject(unstructuredService, bindingName)
	if err != nil {
		return nil, "", "", err
	}

	serviceBinding := kclient.NewServiceBindingObject(
		bindingName, bindAsFiles, workloadName, namingStrategy, workloadGVK, []sboApi.Mapping{}, []sboApi.Service{service}, sboApi.ServiceBindingStatus{})

	var backend backendpkg.AddBindingBackend
	if len(flags) == 0 {
		backend = o.interactiveBackend
	} else {
		backend = o.flagsBackend
	}

	var options []asker.CreationOption
	for len(options) == 0 {
		options, err = backend.SelectCreationOptions(flags)
		if err != nil {
			return nil, "", "", err
		}
	}

	// Note: we cannot directly marshal the serviceBinding object to yaml because it doesn't do that in the correct k8s manifest format
	serviceBindingUnstructured, err := kclient.ConvertK8sResourceToUnstructured(serviceBinding)
	if err != nil {
		return nil, "", "", err
	}
	yamlDesc, err := yaml.Marshal(serviceBindingUnstructured.UnstructuredContent())
	if err != nil {
		return nil, "", "", err
	}

	var filename string
	for _, option := range options {
		if option == asker.OutputToFile {
			filename, err = backend.AskOutputFilePath(flags, filepath.Join("kubernetes", serviceBinding.GetName()+".yaml"))
			if err != nil {
				return nil, "", "", err
			}
			break
		}
	}

	var output string
	for _, option := range options {
		switch option {
		case asker.OutputToFile:
			err = os.MkdirAll(filepath.Dir(filename), 0750)
			if err != nil {
				return nil, "", "", err
			}
			err = os.WriteFile(filename, yamlDesc, 0600)
			if err != nil {
				return nil, "", "", err
			}

		case asker.OutputToStdout:
			output = string(yamlDesc)

		case asker.CreateOnCluster:
			_, err = o.kubernetesClient.PatchDynamicResource(serviceBindingUnstructured)
			if err != nil {
				return nil, "", "", err
			}
		}
	}

	return options, output, filename, nil
}
