package backend

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	dfutil "github.com/devfile/library/pkg/util"
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

func (o *FlagsBackend) SelectServiceInstance(flags map[string]string, options []string) (string, error) {
	var serviceName = flags[FLAG_SERVICE]
	var counter int
	var service string
	for _, option := range options {
		if strings.Contains(option, serviceName) {
			counter++
			service = option
		}
	}
	if counter == 0 {
		return "", fmt.Errorf("%q service not found", serviceName)
	}
	if counter > 1 {
		return "", fmt.Errorf("Found more than one services with name %q. Please mention <name>/<kind>.<apigroup>", serviceName)
	}

	// 	TODO: if a service with the name exists, do nothing, else error out and tell the user they need to mention <name>/<kind>.<apigroup>
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
