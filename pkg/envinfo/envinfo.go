// envinfo package is DEPRECATED and will be removed during v3 implementation
package envinfo

import (
	"github.com/devfile/library/pkg/devfile/parser"

	"github.com/redhat-developer/odo/pkg/devfile/location"
	"github.com/redhat-developer/odo/pkg/localConfigProvider"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
)

// EnvInfo holds all the env specific information relevant to a specific Component.
type EnvInfo struct {
	devfileObj parser.DevfileObj
}

// EnvSpecificInfo wraps the envinfo and provides helpers to
// serialize it.
type EnvSpecificInfo struct {
	devfilePath string
	EnvInfo     `yaml:",omitempty"`
}

var _ localConfigProvider.LocalConfigProvider = (*EnvSpecificInfo)(nil)

func (esi EnvSpecificInfo) GetDevfilePath() string {
	return esi.devfilePath
}

// NewEnvSpecificInfo retrieves the information about devfile path
func NewEnvSpecificInfo(envDir string) (*EnvSpecificInfo, error) {
	return newEnvSpecificInfo(envDir, filesystem.Get())
}

// newEnvSpecificInfo retrieves the information about devfile path
func newEnvSpecificInfo(envDir string, fs filesystem.Filesystem) (*EnvSpecificInfo, error) {
	// Get the path of the environment file
	devfilePath := location.DevfileLocation(envDir)

	// Organize that information into a struct
	e := EnvSpecificInfo{
		EnvInfo:     EnvInfo{},
		devfilePath: devfilePath,
	}

	return &e, nil
}

// SetDevfileObj sets the devfileObj for the envinfo
func (ei *EnvInfo) SetDevfileObj(devfileObj parser.DevfileObj) {
	ei.devfileObj = devfileObj
}

// GetDevfileObj returns devfileObj of the envinfo
func (ei *EnvInfo) GetDevfileObj() parser.DevfileObj {
	return ei.devfileObj
}
