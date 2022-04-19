package binding

import (
	"fmt"
	"strings"

	"github.com/devfile/library/pkg/devfile/parser"
	servicebinding "github.com/redhat-developer/service-binding-operator/apis/binding/v1alpha1"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/redhat-developer/odo/pkg/binding/asker"
	"github.com/redhat-developer/odo/pkg/binding/backend"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/libdevfile"
)

type BindingClient struct {
	// Backends
	flagsBackend       *backend.FlagsBackend
	interactiveBackend *backend.InteractiveBackend

	// Clients
	kubernetesClient kclient.ClientInterface
}

func NewBindingClient(kubernetesClient kclient.ClientInterface) *BindingClient {
	// We create the asker client and the backends here and not at the CLI level, as we want to hide these details to the CLI
	askerClient := asker.NewSurveyAsker()
	return &BindingClient{
		flagsBackend:       backend.NewFlagsBackend(),
		interactiveBackend: backend.NewInteractiveBackend(askerClient),
		kubernetesClient:   kubernetesClient,
	}
}

// GetFlags gets the flag specific to init operation so that it can correctly decide on the backend to be used
// It ignores all the flags except the ones specific to init operation, for e.g. verbosity flag
func (o *BindingClient) GetFlags(flags map[string]string) map[string]string {
	initFlags := map[string]string{}
	for flag, value := range flags {
		if flag == backend.FLAG_NAME || flag == backend.FLAG_SERVICE {
			initFlags[flag] = value
		}
	}
	return initFlags
}

// Validate calls Validate method of the adequate backend
func (o *BindingClient) Validate(flags map[string]string) error {
	var backend backend.CreateBindingBackend
	if len(flags) == 0 {
		backend = o.interactiveBackend
	} else {
		backend = o.flagsBackend
	}
	return backend.Validate(flags)
}

func (o *BindingClient) SelectServiceInstance(flags map[string]string, options []string) (string, error) {
	var backend backend.CreateBindingBackend
	if len(flags) == 0 {
		backend = o.interactiveBackend
	} else {
		backend = o.flagsBackend
	}
	return backend.SelectServiceInstance(flags, options)
}

func (o *BindingClient) AskBindingName(componentName string, flags map[string]string) (string, error) {
	var backend backend.CreateBindingBackend
	if len(flags) == 0 {
		backend = o.interactiveBackend
	} else {
		backend = o.flagsBackend
	}
	return backend.AskBindingName(componentName, flags)
}

func (o *BindingClient) AskBindAsFiles(flags map[string]string) (bool, error) {
	var backend backend.CreateBindingBackend
	if len(flags) == 0 {
		backend = o.interactiveBackend
	} else {
		backend = o.flagsBackend
	}
	return backend.AskBindAsFiles(flags)
}

func (o *BindingClient) CreateBinding(serviceName string, bindingName string, bindAsFiles bool, obj parser.DevfileObj, serviceMap map[string]servicebinding.Ref, componentContext string) error {
	deploymentName := fmt.Sprintf("%s-app", obj.GetMetadataName())
	deployment, err := o.kubernetesClient.GetDeploymentByName(deploymentName)
	if err != nil {
		return err
	}
	deploymentGVR, err := o.kubernetesClient.GetRestMappingFromGVK(deployment.GroupVersionKind())
	if err != nil {
		return err
	}
	serviceRef := serviceMap[serviceName]
	gvr, err := o.kubernetesClient.GetRestMappingFromGVK(schema.GroupVersionKind{
		Group:   serviceRef.Group,
		Version: serviceRef.Version,
		Kind:    serviceRef.Kind,
	})
	if err != nil {
		return err
	}

	service := servicebinding.Service{
		Id: &bindingName, // Id field is helpful if user wants to inject mappings (custom binding data)
		NamespacedRef: servicebinding.NamespacedRef{
			Ref: servicebinding.Ref{
				Group:    gvr.GroupVersionKind.Group,
				Version:  gvr.GroupVersionKind.Version,
				Kind:     gvr.GroupVersionKind.Kind,
				Name:     strings.Split(serviceName, " ")[0],
				Resource: gvr.Resource.Resource,
			},
		},
	}
	serviceBinding := &servicebinding.ServiceBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: strings.Join([]string{kclient.ServiceBindingGroup, kclient.ServiceBindingVersion}, "/"),
			Kind:       kclient.ServiceBindingKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: bindingName,
		},
		Spec: servicebinding.ServiceBindingSpec{
			DetectBindingResources: true,
			BindAsFiles:            bindAsFiles,
			Application: servicebinding.Application{
				Ref: servicebinding.Ref{
					Name:     deployment.Name,
					Group:    deploymentGVR.Resource.Group,
					Version:  deploymentGVR.Resource.Version,
					Resource: deploymentGVR.Resource.Resource,
				},
			},
			Mappings: []servicebinding.Mapping{},
			Services: []servicebinding.Service{service},
		},
	}

	serviceBindingUnstructured, err := kclient.ConvertK8sResourceToUnstructured(serviceBinding)
	if err != nil {
		return err
	}

	yamlDesc, err := yaml.Marshal(serviceBindingUnstructured.UnstructuredContent())
	if err != nil {
		return err
	}
	if bindAsFiles {
		err = libdevfile.AddKubernetesComponent(string(yamlDesc), serviceBinding.Name, componentContext, obj)
		if err != nil {
			return err
		}
	} else {
		err = libdevfile.AddKubernetesComponentToDevfile(string(yamlDesc), serviceBinding.Name, obj)
		if err != nil {
			return err
		}
	}
	return nil
}

func (o *BindingClient) GetServiceInstances() ([]string, map[string]servicebinding.Ref, error) {
	var bindableObjectMap = map[string]servicebinding.Ref{}
	bindableKind, err := o.kubernetesClient.GetBindableKinds()
	if err != nil {
		return nil, bindableObjectMap, err
	}
	var bindableObjects []servicebinding.Ref

	for _, bks := range bindableKind.Status {
		// check every GroupKind only once
		gkAlreadyAdded := false
		for _, bo := range bindableObjects {
			if bo.Group == bks.Group && bo.Kind == bks.Kind {
				gkAlreadyAdded = true
				continue
			}
		}
		if gkAlreadyAdded {
			continue
		}
		gvk := schema.GroupVersionKind{
			Group:   bks.Group,
			Version: bks.Version,
			Kind:    bks.Kind,
		}
		gvr, err := o.kubernetesClient.GetRestMappingFromGVK(gvk)
		if err != nil {
			return nil, bindableObjectMap, err
		}
		// TODO: Debug into why List returns all the versions instead of the GVR version
		resources, err := o.kubernetesClient.ListDynamicResources(gvr.Resource)
		if err != nil {
			return nil, bindableObjectMap, err
		}
		for _, result := range resources.Items {
			bindableObjects = append(bindableObjects, servicebinding.Ref{
				Name:    result.GetName(),
				Group:   result.GroupVersionKind().Group,
				Version: result.GroupVersionKind().Version,
				Kind:    result.GroupVersionKind().Kind,
			})
		}
	}
	var options []string

	for _, option := range bindableObjects {
		gvk := schema.GroupVersionKind{
			Group:   option.Group,
			Version: option.Version,
			Kind:    option.Kind,
		}
		serviceName := fmt.Sprintf("%s (%s)", option.Name, gvk.String())
		options = append(options, serviceName)
		bindableObjectMap[serviceName] = option
	}
	// TODO: if options is empty; then return a more user friendly error
	return options, bindableObjectMap, nil
}
