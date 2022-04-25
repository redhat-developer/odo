package backend

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	dfutil "github.com/devfile/library/pkg/util"
	servicebinding "github.com/redhat-developer/service-binding-operator/apis/binding/v1alpha1"
)

const (
	FLAG_SERVICE       = "service"
	FLAG_NAME          = "name"
	FLAG_BIND_AS_FILES = "bind-as-files"
)

var BINDING_FLAGS = []string{FLAG_NAME, FLAG_SERVICE, FLAG_BIND_AS_FILES}

// FlagsBackend is a backend that will extract all needed information from flags passed to the command
type FlagsBackend struct {
}

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

	err := dfutil.ValidateK8sResourceName("name", flags[FLAG_NAME])
	if err != nil {
		return err
	}

	return nil
}

// SelectServiceInstance returns the service name in the form of '<name> (<kind>.<apigroup>)'
func (o *FlagsBackend) SelectServiceInstance(flags map[string]string, options []string, serviceMap map[string]servicebinding.Ref) (string, error) {
	var service string
	serviceName, kind, group := parseServiceName(flags[FLAG_SERVICE])
	// services tracks all the services that matches flags[FLAG_SERVICE]
	var services []string
	for _, option := range options {
		// option has format `<name> (<kind>.<apigroup>)`
		optionName := strings.Split(option, " ")[0]
		if optionName == serviceName {
			if kind != "" && serviceMap[option].Kind == kind {
				if group != "" && serviceMap[option].Group == group {
					service = option
					services = append(services, service)
					continue
				} else if group == "" {
					service = option
					services = append(services, service)
					continue
				}
			} else if kind == "" {
				service = option
				services = append(services, service)
			}
		}
	}
	if len(services) == 0 {
		return "", fmt.Errorf("%q service not found", flags[FLAG_SERVICE])
	}
	if len(services) > 1 {
		return "", fmt.Errorf("Found more than one services with name %q [%+v]. Please mention <name>/<kind>.<apigroup>", flags[FLAG_SERVICE], strings.Join(services, ","))
	}

	return service, nil
}

func (o *FlagsBackend) AskBindingName(_ string, flags map[string]string) (string, error) {
	return flags[FLAG_NAME], nil
}

func (o *FlagsBackend) AskBindAsFiles(flags map[string]string) (bool, error) {
	if flags[FLAG_BIND_AS_FILES] == "" {
		return false, nil
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
