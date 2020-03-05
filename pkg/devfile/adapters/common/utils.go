package common

import (
	"github.com/golang/glog"

	"github.com/openshift/odo/pkg/devfile/versions"
	"github.com/openshift/odo/pkg/devfile/versions/common"
)

// GetSupportedComponents iterates through the components in the devfile and returns a list of odo supported components
func GetSupportedComponents(data versions.DevfileData) []common.DevfileComponent {
	var components []common.DevfileComponent
	// Only components with aliases are considered because without an alias commands cannot reference them
	for _, comp := range data.GetAliasedComponents() {
		if comp.Type == common.DevfileComponentTypeDockerimage {
			glog.V(3).Infof("Found component %v with alias %v\n", comp.Type, *comp.Alias)
			components = append(components, comp)
		}
	}
	return components
}
