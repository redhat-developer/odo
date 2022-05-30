package binding

import (
	"fmt"
	"path/filepath"

	sboApi "github.com/redhat-developer/service-binding-operator/apis/binding/v1alpha1"
	sbcApi "github.com/redhat-developer/service-binding-operator/apis/spec/v1alpha3"

	"gopkg.in/yaml.v2"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	devfilev1alpha2 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	parsercommon "github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	devfilefs "github.com/devfile/library/pkg/testingutil/filesystem"

	"github.com/redhat-developer/odo/pkg/api"
	"github.com/redhat-developer/odo/pkg/binding/asker"
	backendpkg "github.com/redhat-developer/odo/pkg/binding/backend"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/libdevfile"
)

type BindingClient struct {
	// Backends
	flagsBackend       *backendpkg.FlagsBackend
	interactiveBackend *backendpkg.InteractiveBackend

	// Clients
	kubernetesClient kclient.ClientInterface
}

func NewBindingClient(kubernetesClient kclient.ClientInterface) *BindingClient {
	// We create the asker client and the backends here and not at the CLI level, as we want to hide these details to the CLI
	askerClient := asker.NewSurveyAsker()
	return &BindingClient{
		flagsBackend:       backendpkg.NewFlagsBackend(),
		interactiveBackend: backendpkg.NewInteractiveBackend(askerClient),
		kubernetesClient:   kubernetesClient,
	}
}

// GetFlags gets the flag specific to add binding operation so that it can correctly decide on the backend to be used
// It ignores all the flags except the ones specific to add binding operation, for e.g. verbosity flag
func (o *BindingClient) GetFlags(flags map[string]string) map[string]string {
	bindingFlags := map[string]string{}
	for flag, value := range flags {
		if flag == backendpkg.FLAG_NAME || flag == backendpkg.FLAG_SERVICE || flag == backendpkg.FLAG_BIND_AS_FILES {
			bindingFlags[flag] = value
		}
	}
	return bindingFlags
}

// Validate calls Validate method of the adequate backend
func (o *BindingClient) Validate(flags map[string]string) error {
	var backend backendpkg.AddBindingBackend
	if len(flags) == 0 {
		backend = o.interactiveBackend
	} else {
		backend = o.flagsBackend
	}
	return backend.Validate(flags)
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

func (o *BindingClient) AddBinding(bindingName string, bindAsFiles bool, unstructuredService unstructured.Unstructured, obj parser.DevfileObj, componentContext string) (parser.DevfileObj, error) {
	service, err := o.kubernetesClient.NewServiceBindingServiceObject(unstructuredService, bindingName)
	if err != nil {
		return obj, err
	}

	deploymentName := fmt.Sprintf("%s-app", obj.GetMetadataName())
	deploymentGVR, err := o.kubernetesClient.GetDeploymentAPIVersion()
	if err != nil {
		return obj, err
	}

	serviceBinding := kclient.NewServiceBindingObject(bindingName, bindAsFiles, deploymentName, deploymentGVR, []sboApi.Mapping{}, []sboApi.Service{service})

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

func (o *BindingClient) GetServiceInstances() (map[string]unstructured.Unstructured, error) {
	// Get the BindableKinds/bindable-kinds object
	bindableKind, err := o.kubernetesClient.GetBindableKinds()
	if err != nil {
		return nil, err
	}

	// get a list of restMappings of all the GVKs present in bindableKind's Status
	bindableKindRestMappings, err := o.kubernetesClient.GetBindableKindStatusRestMapping(bindableKind.Status)
	if err != nil {
		return nil, err
	}

	var bindableObjectMap = map[string]unstructured.Unstructured{}
	for _, restMapping := range bindableKindRestMappings {
		// TODO: Debug into why List returns all the versions instead of the GVR version
		// List all the instances of the restMapping object
		resources, err := o.kubernetesClient.ListDynamicResources(restMapping.Resource)
		if err != nil {
			return nil, err
		}

		for _, item := range resources.Items {
			// format: `<name> (<kind>.<group>)`
			serviceName := fmt.Sprintf("%s (%s.%s)", item.GetName(), item.GetKind(), item.GroupVersionKind().Group)
			bindableObjectMap[serviceName] = item
		}

	}

	return bindableObjectMap, nil
}

func (o *BindingClient) GetBindingsFromDevfile(devfileObj parser.DevfileObj, context string) ([]api.ServiceBinding, error) {
	result := []api.ServiceBinding{}
	kubeComponents, err := devfileObj.Data.GetComponents(parsercommon.DevfileOptions{
		ComponentOptions: parsercommon.ComponentOptions{
			ComponentType: devfilev1alpha2.KubernetesComponentType,
		},
	})
	if err != nil {
		return nil, err
	}

	for _, component := range kubeComponents {
		strCRD, err := libdevfile.GetK8sManifestWithVariablesSubstituted(devfileObj, component.Name, context, devfilefs.DefaultFs{})
		if err != nil {
			return nil, err
		}

		u := unstructured.Unstructured{}
		if err := yaml.Unmarshal([]byte(strCRD), &u.Object); err != nil {
			return nil, err
		}

		switch u.GetObjectKind().GroupVersionKind() {
		case sboApi.GroupVersionKind:

			sb, err := o.getApiServiceBindingFromBinding(u)
			if err != nil {
				return nil, err
			}

			sb.Status, err = o.getStatusFromBinding(sb.Name)
			if err != nil {
				return nil, err
			}

			result = append(result, sb)

		case sbcApi.GroupVersion.WithKind("ServiceBinding"):

			sb, err := o.getApiServiceBindingFromSpecBinding(u)
			if err != nil {
				return nil, err
			}

			sb.Status, err = o.getStatusFromSpecBinding(sb.Name)
			if err != nil {
				return nil, err
			}

			result = append(result, sb)

		}
	}
	return result, nil
}

func (o *BindingClient) getApiServiceBindingFromBinding(u unstructured.Unstructured) (api.ServiceBinding, error) {
	var sb sboApi.ServiceBinding
	err := o.kubernetesClient.ConvertUnstructuredToResource(u, &sb)
	if err != nil {
		return api.ServiceBinding{}, err
	}

	var dstSvcs []sbcApi.ServiceBindingServiceReference
	for _, srcSvc := range sb.Spec.Services {
		dstSvc := sbcApi.ServiceBindingServiceReference{
			Name: srcSvc.Name,
		}
		dstSvc.APIVersion, dstSvc.Kind = schema.GroupVersion{
			Group:   srcSvc.Group,
			Version: srcSvc.Version,
		}.WithKind(srcSvc.Kind).ToAPIVersionAndKind()
		dstSvcs = append(dstSvcs, dstSvc)
	}
	return api.ServiceBinding{
		Name: sb.Name,
		Spec: api.ServiceBindingSpec{
			Services:               dstSvcs,
			DetectBindingResources: sb.Spec.DetectBindingResources,
			BindAsFiles:            sb.Spec.BindAsFiles,
		},
	}, nil
}

func (o *BindingClient) getStatusFromBinding(name string) (*api.ServiceBindingStatus, error) {
	sb, err := o.kubernetesClient.GetServiceBinding(name)
	if err != nil {
		if kerrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	secretName := sb.Status.Secret
	secret, err := o.kubernetesClient.GetSecret(secretName, o.kubernetesClient.GetCurrentNamespace())
	if err != nil {
		return nil, err
	}

	if sb.Spec.BindAsFiles {
		bindings := make([]string, 0, len(secret.Data))
		for k := range secret.Data {
			bindingName := filepath.ToSlash(filepath.Join("${SERVICE_BINDING_ROOT}", name, k))
			bindings = append(bindings, bindingName)
		}
		return &api.ServiceBindingStatus{
			BindingFiles: bindings,
		}, nil
	}

	bindings := make([]string, 0, len(secret.Data))
	for k := range secret.Data {
		bindings = append(bindings, k)
	}
	return &api.ServiceBindingStatus{
		BindingEnvVars: bindings,
	}, nil
}

func (o *BindingClient) getStatusFromSpecBinding(name string) (*api.ServiceBindingStatus, error) {
	sb, err := o.kubernetesClient.GetSpecServiceBinding(name)
	if err != nil {
		if kerrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	if sb.Status.Binding == nil {
		return nil, nil
	}
	secretName := sb.Status.Binding.Name
	secret, err := o.kubernetesClient.GetSecret(secretName, o.kubernetesClient.GetCurrentNamespace())
	if err != nil {
		return nil, err
	}
	bindingFiles := make([]string, 0, len(secret.Data))
	bindingEnvVars := make([]string, 0, len(sb.Spec.Env))
	for k := range secret.Data {
		bindingName := filepath.ToSlash(filepath.Join("${SERVICE_BINDING_ROOT}", name, k))
		bindingFiles = append(bindingFiles, bindingName)
	}
	for _, env := range sb.Spec.Env {
		bindingEnvVars = append(bindingEnvVars, env.Name)
	}
	return &api.ServiceBindingStatus{
		BindingFiles:   bindingFiles,
		BindingEnvVars: bindingEnvVars,
	}, nil
}

func (o *BindingClient) getApiServiceBindingFromSpecBinding(u unstructured.Unstructured) (api.ServiceBinding, error) {
	var sb sbcApi.ServiceBinding
	err := o.kubernetesClient.ConvertUnstructuredToResource(u, &sb)
	if err != nil {
		return api.ServiceBinding{}, err
	}
	return api.ServiceBinding{
		Name: sb.Name,
		Spec: api.ServiceBindingSpec{
			Services:               []sbcApi.ServiceBindingServiceReference{sb.Spec.Service},
			DetectBindingResources: false,
			BindAsFiles:            true,
		},
	}, nil
}
