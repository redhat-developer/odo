package devfile

import (
	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/v2/pkg/devfile/parser"
	parsercommon "github.com/devfile/library/v2/pkg/devfile/parser/data/v2/common"
)

// GetK8sAndOcComponentsToPush returns the list of Kubernetes and OpenShift components to push,
// The list returned is governed by the DeployByDefault field in each component.
// All components with DeployByDefault set to true are included, along with those with no DeployByDefault set and not-referenced.
// It takes an additional allowApply boolean, which set to true, will append the components referenced from apply commands to the list.
func GetK8sAndOcComponentsToPush(devfileObj parser.DevfileObj, allowApply bool) ([]devfilev1.Component, error) {
	var allComponents []devfilev1.Component

	k8sComponents, err := devfileObj.Data.GetComponents(parsercommon.DevfileOptions{
		ComponentOptions: parsercommon.ComponentOptions{ComponentType: devfilev1.KubernetesComponentType},
	})
	if err != nil {
		return nil, err
	}
	allComponents = append(allComponents, k8sComponents...)

	ocComponents, err := devfileObj.Data.GetComponents(parsercommon.DevfileOptions{
		ComponentOptions: parsercommon.ComponentOptions{ComponentType: devfilev1.OpenshiftComponentType},
	})
	if err != nil {
		return nil, err
	}
	allComponents = append(allComponents, ocComponents...)

	allApplyCommands, err := devfileObj.Data.GetCommands(parsercommon.DevfileOptions{
		CommandOptions: parsercommon.CommandOptions{CommandType: devfilev1.ApplyCommandType},
	})
	if err != nil {
		return nil, err
	}

	m := make(map[string]devfilev1.Component)
	for _, comp := range allComponents {
		if comp.Kubernetes == nil && comp.Openshift == nil {
			continue
		}
		var k devfilev1.K8sLikeComponent
		if comp.Kubernetes != nil {
			k = comp.Kubernetes.K8sLikeComponent
		} else {
			k = comp.Openshift.K8sLikeComponent
		}
		var add bool
		if allowApply && isComponentReferenced(allApplyCommands, comp.Name) {
			add = true
		} else if k.DeployByDefault == nil {
			// auto-created only if not referenced by any apply command
			if !isComponentReferenced(allApplyCommands, comp.Name) {
				add = true
			}
		} else if *k.DeployByDefault {
			add = true
		}
		if !add {
			continue
		}
		if _, present := m[comp.Name]; !present {
			m[comp.Name] = comp
		}
	}

	var result []devfilev1.Component
	for _, comp := range m {
		result = append(result, comp)
	}
	return result, nil
}

func isComponentReferenced(allApplyCommands []devfilev1.Command, cmpName string) bool {
	for _, cmd := range allApplyCommands {
		if cmd.Apply == nil {
			continue
		}
		if cmd.Apply.Component == cmpName {
			return true
		}
	}
	return false
}
