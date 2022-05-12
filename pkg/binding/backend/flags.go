package backend

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	dfutil "github.com/devfile/library/pkg/util"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	FLAG_SERVICE       = "service"
	FLAG_NAME          = "name"
	FLAG_BIND_AS_FILES = "bind-as-files"
)

// FlagsBackend is a backend that will extract all needed information from flags passed to the command
type FlagsBackend struct{}

func NewFlagsBackend() *FlagsBackend {
	return &FlagsBackend{}
}

func (o *FlagsBackend) Validate(flags map[string]string) error {
	if flags[FLAG_SERVICE] == "" {
		return errors.New("missing --service parameter: please add --service <name>[/<kind>.<apigroup>] to specify the service instance for binding")
	}
	if flags[FLAG_NAME] == "" {
		return errors.New("missing --name parameter: please add --name <name> to specify a name for the service binding instance")
	}

	return dfutil.ValidateK8sResourceName(FLAG_NAME, flags[FLAG_NAME])
}

// SelectServiceInstance parses the service's name, kind, and group from arg:serviceName,
// after which it checks if the service is available in arg:serviceMap, it further checks for kind, and group
// If a single service is found, it returns the service name in the form of '<name> (<kind>.<apigroup>)', else errors out.
// serviceMap: a map of bindable service name with it's unstructured.Unstructured; this map is used to stay independent of the service name format.
func (o *FlagsBackend) SelectServiceInstance(serviceName string, serviceMap map[string]unstructured.Unstructured) (string, error) {
	selectedServiceName, selectedServiceKind, selectedServiceGroup := parseServiceName(serviceName)
	// services tracks all the services that matches flags[FLAG_SERVICE]
	var services []string
	for option, unstructuredService := range serviceMap {
		// option has format `<name> (<kind>.<apigroup>)`
		if unstructuredService.GetName() == selectedServiceName {
			if selectedServiceKind != "" && unstructuredService.GetKind() == selectedServiceKind {
				if selectedServiceGroup != "" && unstructuredService.GroupVersionKind().Group == selectedServiceGroup {
					services = append(services, option)
					continue
				} else if selectedServiceGroup == "" {
					services = append(services, option)
					continue
				}
			} else if selectedServiceKind == "" {
				services = append(services, option)
			}
		}
	}
	if len(services) == 0 {
		return "", fmt.Errorf("%q service not found", serviceName)
	}
	if len(services) > 1 {
		return "", fmt.Errorf("Found more than one services with name %q [%+v]. Please mention <name>/<kind>.<apigroup>", serviceName, strings.Join(services, ","))
	}

	return services[0], nil
}

func (o *FlagsBackend) AskBindingName(_ string, flags map[string]string) (string, error) {
	return flags[FLAG_NAME], nil
}

func (o *FlagsBackend) AskBindAsFiles(flags map[string]string) (bool, error) {
	if flags[FLAG_BIND_AS_FILES] == "" {
		// default value for bindAsFiles must be true
		return true, nil
	}
	bindAsFiles, err := strconv.ParseBool(flags[FLAG_BIND_AS_FILES])
	if err != nil {
		return false, fmt.Errorf("unable to set %q to --%v, value must be a boolean", flags[FLAG_BIND_AS_FILES], FLAG_BIND_AS_FILES)
	}
	return bindAsFiles, nil
}

// parseServiceName parses various service name formats. It supports the following formats:
// - <name>
// - <name>.<kind>
// - <name>.<kind>.<apigroup>
// - <name>/<kind>
// - <name>/<kind>.<apigroup>
func parseServiceName(service string) (name, kind, group string) {
	if serviceNKG := strings.Split(service, "/"); len(serviceNKG) > 1 {
		// Parse <name>/<kind>
		name = serviceNKG[0]
		kindGroup := strings.SplitN(serviceNKG[1], ".", 2)
		kind = kindGroup[0]
		if len(kindGroup) > 1 {
			// Parse <name>/<kind>.<apigroup>
			group = kindGroup[1]
		}
	} else if serviceNKG = strings.SplitN(service, ".", 3); len(serviceNKG) > 1 {
		// Parse <name>.<kind>
		name = serviceNKG[0]
		kind = serviceNKG[1]
		if len(serviceNKG) > 2 {
			// Parse <name>.<kind>.<apigroup>
			group = serviceNKG[2]
		}
	} else {
		// Parse <name>
		name = service
	}
	return
}
