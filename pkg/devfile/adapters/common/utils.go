package common

import (
	"os"

	"github.com/golang/glog"

	"github.com/openshift/odo/pkg/devfile/parser/data"
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
)

// PredefinedDevfileCommands encapsulates constants for predefined devfile commands
type PredefinedDevfileCommands string

const (
	// DefaultDevfileInitCommand is a predefined devfile command for init
	DefaultDevfileInitCommand PredefinedDevfileCommands = "devinit"

	// DefaultDevfileBuildCommand is a predefined devfile command for build
	DefaultDevfileBuildCommand PredefinedDevfileCommands = "devbuild"

	// DefaultDevfileRunCommand is a predefined devfile command for run
	DefaultDevfileRunCommand PredefinedDevfileCommands = "devrun"

	// SupervisordInitContainerName The init container name for supervisord
	SupervisordInitContainerName = "copy-supervisord"

	// Default Image that will be used containing the supervisord binary and assembly scripts
	// use GetBootstrapperImage() function instead of this variable
	defaultBootstrapperImage = "registry.access.redhat.com/openshiftdo/odo-init-image-rhel7:1.1.2"

	// SupervisordVolumeName Create a custom name and (hope) that users don't use the *exact* same name in their deployment (occlient.go)
	SupervisordVolumeName = "odo-supervisord-shared-data"

	// SupervisordMountPath The supervisord Mount Path for the container mounting the supervisord volume
	SupervisordMountPath = "/opt/odo/"

	// SupervisordBinaryPath The supervisord binary path inside the container volume mount
	SupervisordBinaryPath = "/opt/odo/bin/supervisord"

	// SupervisordConfFile The supervisord configuration file inside the container volume mount
	SupervisordConfFile = "/opt/odo/conf/devfile-supervisor.conf"

	// OdoInitImageContents The path to the odo init image contents
	OdoInitImageContents = "/opt/odo-init/."

	// ENV variable to overwrite image used to bootstrap SupervisorD in S2I and Devfile builder Image
	bootstrapperImageEnvName = "ODO_BOOTSTRAPPER_IMAGE"
)

func isComponentSupported(component common.DevfileComponent) bool {
	// Currently odo only uses devfile components of type dockerimage, since most of the Che registry devfiles use it
	return component.Type == common.DevfileComponentTypeDockerimage
}

// GetBootstrapperImage returns the odo-init bootstrapper image
func GetBootstrapperImage() string {
	if env, ok := os.LookupEnv(bootstrapperImageEnvName); ok {
		return env
	}
	return defaultBootstrapperImage
}

// GetSupportedComponents iterates through the components in the devfile and returns a list of odo supported components
func GetSupportedComponents(data data.DevfileData) []common.DevfileComponent {
	var components []common.DevfileComponent
	// Only components with aliases are considered because without an alias commands cannot reference them
	for _, comp := range data.GetAliasedComponents() {
		if isComponentSupported(comp) {
			glog.V(3).Infof("Found component \"%v\" with alias \"%v\"\n", comp.Type, *comp.Alias)
			components = append(components, comp)
		}
	}
	return components
}
