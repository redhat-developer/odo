package devfile

import (
	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	parsercommon "github.com/devfile/library/pkg/devfile/parser/data/v2/common"
)

const (
	filePrefix = "odo-service-"
)

// GetKubernetesComponentsToPush returns the list of Kubernetes components to push,
// by getting the list of Kubernetes components and removing the ones
// referenced from a command in the devfile
func GetKubernetesComponentsToPush(devfileObj parser.DevfileObj) ([]devfilev1.Component, error) {
	k8sComponents, err := devfileObj.Data.GetComponents(parsercommon.DevfileOptions{
		ComponentOptions: parsercommon.ComponentOptions{ComponentType: devfilev1.KubernetesComponentType},
	})
	if err != nil {
		return nil, err
	}

	componentsMap := map[string]devfilev1.Component{}
	for _, component := range k8sComponents {
		componentsMap[component.Name] = component
	}

	commands, err := devfileObj.Data.GetCommands(parsercommon.DevfileOptions{})
	if err != nil {
		return nil, err
	}

	for _, command := range commands {
		componentName := ""
		if command.Exec != nil {
			componentName = command.Exec.Component
		} else if command.Apply != nil {
			componentName = command.Apply.Component
		}
		if componentName == "" {
			continue
		}
		delete(componentsMap, componentName)
	}

	k8sComponents = make([]devfilev1.Component, len(componentsMap))
	i := 0
	for _, v := range componentsMap {
		k8sComponents[i] = v
		i++
	}

	return k8sComponents, err
}

// IsComponentDefined checks if a component with the given name is defined in a DevFile
func IsComponentDefined(name string, devfileObj parser.DevfileObj) (bool, error) {
	components, err := devfileObj.Data.GetComponents(common.DevfileOptions{})
	if err != nil {
		return false, err
	}
	for _, c := range components {
		if c.Name == name {
			return true, nil
		}
	}
	return false, nil
}
