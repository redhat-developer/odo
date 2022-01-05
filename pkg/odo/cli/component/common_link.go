package component

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/redhat-developer/odo/pkg/devfile"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/util"
	svc "github.com/redhat-developer/odo/pkg/service"

	v1 "k8s.io/api/core/v1"

	kerrors "k8s.io/apimachinery/pkg/api/errors"

	servicebinding "github.com/redhat-developer/service-binding-operator/apis/binding/v1alpha1"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const unlink = "unlink"

type commonLinkOptions struct {
	secretName       string
	isTargetAService bool
	name             string
	bindAsFiles      bool

	devfilePath string

	suppliedName  string
	operation     func(secretName, componentName, applicationName string) error
	operationName string

	// Service Binding Operator options
	serviceBinding *servicebinding.ServiceBinding
	serviceType    string
	serviceName    string
	*genericclioptions.Context
	// choose between Operator Hub and Service Catalog. If true, Operator Hub
	csvSupport bool

	inlined bool
	// mappings is an array of strings representing the custom binding data that user wants to inject into the component
	mappings []string
}

func newCommonLinkOptions() *commonLinkOptions {
	return &commonLinkOptions{}
}

func (o *commonLinkOptions) getLinkType() string {
	linkType := "component"
	if o.isTargetAService {
		linkType = "service"
	}
	return linkType
}

// Complete completes LinkOptions after they've been created
func (o *commonLinkOptions) complete(cmdline cmdline.Cmdline, args []string, context string) (err error) {
	o.operationName = cmdline.GetName()
	o.suppliedName = args[0]

	o.Context, err = genericclioptions.New(genericclioptions.NewCreateParameters(cmdline).NeedDevfile(context))
	if err != nil {
		return err
	}

	o.csvSupport, _ = o.KClient.IsCSVSupported()

	o.serviceType, o.serviceName, err = svc.IsOperatorServiceNameValid(o.suppliedName)
	if err != nil {
		// error indicates the service name provided by user doesn't adhere to <crd-name>/<instance-name>
		// so it's another odo component that they want to link/unlink to/from
		o.serviceName = o.suppliedName
		o.isTargetAService = false
		o.serviceType = "Service" // Kubernetes Service

		// TODO find the service using an app name to link components in other apps
		// requires modification of the app flag or finding some other way
		var s *v1.Service
		s, err = o.Context.KClient.GetOneService(o.suppliedName, o.EnvSpecificInfo.GetApplication())
		if kerrors.IsNotFound(err) {
			return fmt.Errorf("couldn't find component named %q. Refer %q to see list of running components", o.suppliedName, "odo list")
		}
		if err != nil {
			return err
		}
		o.serviceName = s.Name
	} else {
		o.isTargetAService = true
	}

	if o.operationName == unlink {
		// rest of the code is specific to link operation
		return nil
	}

	componentName := o.EnvSpecificInfo.GetName()

	deployment, err := o.KClient.GetOneDeployment(componentName, o.EnvSpecificInfo.GetApplication())
	if err != nil {
		return err
	}

	deploymentGVR, err := o.KClient.GetDeploymentAPIVersion()
	if err != nil {
		return err
	}

	paramsMap, err := util.MapFromParameters(o.mappings)
	if err != nil {
		return err
	}

	// MappingsMap is a map of mappings to be used in the ServiceBinding we create for an "odo link"
	var mappingsMap []servicebinding.Mapping
	for kv := range paramsMap {
		mapping := servicebinding.Mapping{
			Name:  kv,
			Value: paramsMap[kv],
		}
		mappingsMap = append(mappingsMap, mapping)
	}

	var service servicebinding.Service
	if o.isTargetAService {
		// since the service exists, let's get more info to populate service binding request
		// first get the CR itself
		cr, err := o.KClient.GetCustomResource(o.serviceType)
		if err != nil {
			return err
		}

		// now get the group, version, kind information from CR
		group, version, kind, err := svc.GetGVKFromCR(cr)
		if err != nil {
			return err
		}

		service = servicebinding.Service{
			Id: &o.serviceName, // Id field is helpful if user wants to inject mappings (custom binding data)
			NamespacedRef: servicebinding.NamespacedRef{
				Ref: servicebinding.Ref{
					Group:   group,
					Version: version,
					Kind:    kind,
					Name:    o.serviceName,
				},
			},
		}
	} else {
		service = servicebinding.Service{
			Id: &o.serviceName, // Id field is helpful if user wants to inject mappings (custom binding data)
			NamespacedRef: servicebinding.NamespacedRef{
				Ref: servicebinding.Ref{
					Version: "v1",
					Kind:    "Service",
					Name:    o.serviceName,
				},
			},
		}
	}

	o.serviceBinding = &servicebinding.ServiceBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: strings.Join([]string{kclient.ServiceBindingGroup, kclient.ServiceBindingVersion}, "/"),
			Kind:       kclient.ServiceBindingKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: o.getServiceBindingName(componentName),
		},
		Spec: servicebinding.ServiceBindingSpec{
			DetectBindingResources: true,
			BindAsFiles:            o.bindAsFiles,
			Application: servicebinding.Application{
				Ref: servicebinding.Ref{
					Name:     deployment.Name,
					Group:    deploymentGVR.Group,
					Version:  deploymentGVR.Version,
					Resource: deploymentGVR.Resource,
				},
			},
			Mappings: mappingsMap,
			Services: []servicebinding.Service{service},
		},
	}

	return nil
}

func (o *commonLinkOptions) validate() (err error) {
	if o.EnvSpecificInfo == nil {
		return fmt.Errorf("failed to find environment info to validate")
	}

	var svcFullName string

	if o.isTargetAService {
		// let's validate if the service exists
		svcFullName = strings.Join([]string{o.serviceType, o.serviceName}, "/")
		svcExists, err := svc.OperatorSvcExists(o.KClient, svcFullName)
		if err != nil {
			return err
		}
		if !svcExists {
			return fmt.Errorf("couldn't find service named %q. Refer %q to see list of running services", svcFullName, "odo service list")
		}
	} else {
		svcFullName = o.serviceName
		if o.suppliedName == o.EnvSpecificInfo.GetName() {
			if o.operationName == unlink {
				return fmt.Errorf("the component %q cannot be unlinked from itself", o.suppliedName)
			} else {
				return fmt.Errorf("the component %q cannot be linked with itself", o.suppliedName)
			}
		}
	}

	if o.operationName == unlink {
		_, found, err := svc.FindDevfileServiceBinding(o.EnvSpecificInfo.GetDevfileObj(), o.serviceType, o.serviceName, o.GetComponentContext())
		if err != nil {
			return err
		}
		if !found {
			if o.getLinkType() == "service" {
				return fmt.Errorf("failed to unlink the %s %q since no link was found in the configuration referring this %s", o.getLinkType(), svcFullName, o.getLinkType())
			}
			return fmt.Errorf("failed to unlink the %s %q since no link was found in the configuration referring this %s", o.getLinkType(), o.suppliedName, o.getLinkType())
		}
		return nil
	}

	return nil
}

func (o *commonLinkOptions) run() (err error) {
	if o.Context.EnvSpecificInfo != nil {
		if o.operationName == unlink {
			return o.unlinkOperator()
		}
		return o.linkOperator()
	}

	var component string
	if o.Context.EnvSpecificInfo != nil {
		component = o.EnvSpecificInfo.GetName()
		err = o.operation(o.secretName, component, o.GetApplication())
	} else {
		component, err = o.Component()
		if err != nil {
			return err
		}
		err = o.operation(o.secretName, component, o.GetApplication())
	}

	if err != nil {
		return err
	}

	switch o.operationName {
	case "link":
		log.Successf("The %s %s has been successfully linked to the component %s\n", o.getLinkType(), o.suppliedName, component)
	case "unlink":
		log.Successf("The %s %s has been successfully unlinked from the component %s\n", o.getLinkType(), o.suppliedName, component)
	default:
		return fmt.Errorf("unknown operation %s", o.operationName)
	}

	secret, err := o.KClient.GetSecret(o.secretName, o.GetProject())
	if err != nil {
		return err
	}

	if len(secret.Data) == 0 {
		log.Infof("There are no secret environment variables to expose within the %s service", o.suppliedName)
	} else {
		if o.operationName == "link" {
			log.Infof("The below secret environment variables were added to the '%s' component:\n", component)
		} else {
			log.Infof("The below secret environment variables were removed from the '%s' component:\n", component)
		}

		// Output the environment variables
		for i := range secret.Data {
			fmt.Printf("· %v\n", i)
		}

		// Retrieve the first variable to use as an example.
		// Have to use a range to access the map
		var exampleEnv string
		for i := range secret.Data {
			exampleEnv = i
			break
		}

		// Output what to do next if first linking...
		if o.operationName == "link" {
			log.Italicf(`
You can now access the environment variables from within the component pod, for example:
$%s is now available as a variable within component %s`, exampleEnv, component)
		}
	}
	return
}

// getServiceBindingName creates a name to be used for creation/deletion of SBR during link/unlink operations
func (o *commonLinkOptions) getServiceBindingName(componentName string) string {
	if len(o.name) > 0 {
		return o.name
	}
	if !o.isTargetAService {
		return strings.Join([]string{componentName, o.serviceName}, "-")
	}
	return strings.Join([]string{componentName, strings.ToLower(o.serviceType), o.serviceName}, "-")
}

// linkOperator creates a service binding resource and links
// the current component with the given odo service or
// the current component with the given component's service
// and stores the link info in the env
func (o *commonLinkOptions) linkOperator() (err error) {
	// Convert ServiceBinding -> JSON -> Map -> YAML
	// JSON conversion step is necessary to inline TypeMeta

	intermediate, err := json.Marshal(o.serviceBinding)
	if err != nil {
		return err
	}

	serviceBindingMap := make(map[string]interface{})
	err = json.Unmarshal(intermediate, &serviceBindingMap)
	if err != nil {
		return err
	}

	yamlDesc, err := yaml.Marshal(serviceBindingMap)
	if err != nil {
		return err
	}

	// check if the component is already linked to the requested component/service
	_, found, err := svc.FindDevfileServiceBinding(o.EnvSpecificInfo.GetDevfileObj(), o.serviceType, o.serviceName, o.GetComponentContext())
	if err != nil {
		return err
	}
	if found {
		return fmt.Errorf("component %q is already linked with the %s %q", o.Context.EnvSpecificInfo.GetName(), o.getLinkType(), o.suppliedName)
	}

	if o.inlined {
		err = devfile.AddKubernetesComponentToDevfile(string(yamlDesc), o.serviceBinding.Name, o.EnvSpecificInfo.GetDevfileObj())
		if err != nil {
			return err
		}
	} else {
		err = devfile.AddKubernetesComponent(string(yamlDesc), o.serviceBinding.Name, o.GetComponentContext(), o.EnvSpecificInfo.GetDevfileObj())
		if err != nil {
			return err
		}
	}

	log.Successf("Successfully created link between component %q and %s %q\n", o.Context.EnvSpecificInfo.GetName(), o.getLinkType(), o.suppliedName)
	log.Italic("To apply the link, please use `odo push`")
	return err
}

// unlinkOperator deletes the service binding resource from the devfile
func (o *commonLinkOptions) unlinkOperator() (err error) {

	// We already tested `found` in `validateForOperator`
	name, _, err := svc.FindDevfileServiceBinding(o.EnvSpecificInfo.GetDevfileObj(), o.serviceType, o.serviceName, o.GetComponentContext())
	if err != nil {
		return err
	}

	err = devfile.DeleteKubernetesComponentFromDevfile(name, o.EnvSpecificInfo.GetDevfileObj(), o.GetComponentContext())
	if err != nil {
		return err
	}

	log.Successf("Successfully unlinked component %q from %s %q\n", o.Context.EnvSpecificInfo.GetName(), o.getLinkType(), o.suppliedName)
	log.Italic("To apply the changes, please use `odo push`")
	return nil
}
