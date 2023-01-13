package devfile

import (
	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	parsercommon "github.com/devfile/library/pkg/devfile/parser/data/v2/common"
)

// GetKubernetesComponentsToPush returns the list of Kubernetes components to push,
// by getting the list of Kubernetes components and removing the ones
// referenced from a command in the devfile
// It takes an additional allowApply boolean, which set to true, will append the components from apply command to the list
func GetKubernetesComponentsToPush(devfileObj parser.DevfileObj, allowApply bool) ([]devfilev1.Component, error) {
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
		} else if !allowApply && command.Apply != nil {
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

// GetApplyComponentsToPush returns devfile K8s components associated with the given commandName and commandGroupKind;
// these resources will always be referenced with an apply command, hence the function name
func GetApplyComponentsToPush(devfileObj parser.DevfileObj, commandGroupKind devfilev1.CommandGroupKind, commandName string, wantK8sComponent, wantImageComponent bool) (k8sComponentNames, imageComponentNames []string, _ error) {
	commands, err := devfileObj.Data.GetCommands(parsercommon.DevfileOptions{
		CommandOptions: parsercommon.CommandOptions{
			CommandGroupKind: commandGroupKind,
			CommandType:      devfilev1.CompositeCommandType,
		},
		FilterByName: commandName,
	})
	if err != nil {
		return nil, nil, err
	}
	if len(commands) == 0 {
		return nil, nil, nil
	}
	// we assume there will be only one command that matches the command ID
	command := commands[0]
	for _, subCommand := range command.Composite.Commands {
		cmd, err := devfileObj.Data.GetCommands(parsercommon.DevfileOptions{
			FilterByName:   subCommand,
			CommandOptions: parsercommon.CommandOptions{CommandType: devfilev1.ApplyCommandType},
		})
		if err != nil {
			return nil, nil, err
		}
		if len(cmd) == 0 {
			continue
		}
		// we assume there will be only one command that matches the subCommand ID
		if applyCmd := cmd[0]; applyCmd.Apply != nil {
			applyComponents, err := devfileObj.Data.GetComponents(parsercommon.DevfileOptions{
				FilterByName: applyCmd.Apply.Component,
			})
			if err != nil {
				return nil, nil, err
			}
			if len(applyComponents) == 0 {
				continue
			}
			// we assume there will be only one component that matches the ID
			cmp := applyComponents[0]
			if wantK8sComponent && cmp.Kubernetes != nil {
				k8sComponentNames = append(k8sComponentNames, cmp.Name)
			} else if wantImageComponent && cmp.Image != nil {
				imageComponentNames = append(imageComponentNames, cmp.Name)
			}
		}
	}
	return k8sComponentNames, imageComponentNames, nil

}
