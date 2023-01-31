package devfile

import (
	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/v2/pkg/devfile/parser"
	parsercommon "github.com/devfile/library/v2/pkg/devfile/parser/data/v2/common"
)

// GetK8sAndOcComponentsToPush returns the list of Kubernetes and OpenShift components to push,
// by getting the list of Kubernetes and OpenShift components and removing the ones
// referenced from a command in the devfile
// It takes an additional allowApply boolean, which set to true, will append the components from apply command to the list
func GetK8sAndOcComponentsToPush(devfileObj parser.DevfileObj, allowApply bool) ([]devfilev1.Component, error) {
	k8sComponents, err := devfileObj.Data.GetComponents(parsercommon.DevfileOptions{
		ComponentOptions: parsercommon.ComponentOptions{ComponentType: devfilev1.KubernetesComponentType},
	})
	if err != nil {
		return nil, err
	}

	ocComponents, err := devfileObj.Data.GetComponents(parsercommon.DevfileOptions{
		ComponentOptions: parsercommon.ComponentOptions{ComponentType: devfilev1.OpenshiftComponentType},
	})
	if err != nil {
		return nil, err
	}

	allComponents := []devfilev1.Component{}
	allComponents = append(allComponents, k8sComponents...)
	allComponents = append(allComponents, ocComponents...)

	componentsMap := map[string]devfilev1.Component{}
	for _, component := range allComponents {
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
		} else if !allowApply && command.Apply != nil {
			componentName = command.Apply.Component
		}
		if componentName == "" {
			continue
		}
		delete(componentsMap, componentName)
	}

	allComponentsToPush := make([]devfilev1.Component, len(componentsMap))
	i := 0
	for _, v := range componentsMap {
		allComponentsToPush[i] = v
		i++
	}

	return allComponentsToPush, err
}
