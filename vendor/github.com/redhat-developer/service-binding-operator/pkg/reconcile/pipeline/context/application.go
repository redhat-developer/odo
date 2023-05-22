package context

import (
	"fmt"
	"reflect"

	"github.com/redhat-developer/service-binding-operator/apis/binding/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
)

var _ pipeline.Application = &application{}

type application struct {
	gvr                    *schema.GroupVersionResource
	persistedResource      *unstructured.Unstructured
	resource               *unstructured.Unstructured
	bindingPath            *v1alpha1.BindingPath
	bindableContainerNames sets.String
	resourceMapping        pipeline.WorkloadMapping
}

func (a *application) SecretPath() string {
	if a.bindingPath != nil {
		return a.bindingPath.SecretPath
	}
	return ""
}

func (a *application) Resource() *unstructured.Unstructured {
	if a.resource == nil {
		a.resource = a.persistedResource.DeepCopy()
	}
	return a.resource
}

func (a *application) IsUpdated() bool {
	return !reflect.DeepEqual(a.persistedResource, a.resource)
}

func (a *application) BindablePods() (*pipeline.MetaPodSpec, error) {
	var filteredContainers []pipeline.MetaContainer
	for _, container := range a.resourceMapping.Containers {
		pathData, err := container.Path.FindResults(a.Resource().Object)
		if err != nil {
			continue
		}
		for _, data := range pathData[0] {
			sliceData, isSlice := data.Interface().([]interface{})
			mapData, isMap := data.Interface().(map[string]interface{})
			if !(isSlice || isMap) {
				return nil, fmt.Errorf("Unable to convert data to suitable format: %v, received type %v", data, data.Type().String())
			}

			if isSlice {
				for _, c := range sliceData {
					cData, ok := c.(map[string]interface{})
					if !ok {
						return nil, fmt.Errorf("Unable to convert data to suitable format: %v, received type %v", data, data.Type().String())
					}
					mc, err := intoMetaContainer(container, cData)
					if mc != nil {
						if len(a.bindableContainerNames) == 0 || a.bindableContainerNames.Has(mc.Name) {
							filteredContainers = append(filteredContainers, *mc)
						}
					} else {
						return nil, err
					}
				}
			} else if isMap {
				mc, err := intoMetaContainer(container, mapData)
				if mc != nil {
					if len(a.bindableContainerNames) == 0 || a.bindableContainerNames.Has(mc.Name) {
						filteredContainers = append(filteredContainers, *mc)
					}
				} else {
					return nil, err
				}
			}
		}
	}

	if len(filteredContainers) == 0 {
		return nil, fmt.Errorf("no containers found in application resource: data: %v, names: %v", a.Resource(), a.bindableContainerNames)
	}

	containerTemplate := &pipeline.MetaPodSpec{
		Volume:     a.resourceMapping.Volume,
		Containers: filteredContainers,
		Data:       a.resource.Object,
	}

	return containerTemplate, nil
}

func intoMetaContainer(container pipeline.WorkloadContainer, data map[string]interface{}) (*pipeline.MetaContainer, error) {

	name, found, err := unstructured.NestedString(data, container.Name...)
	if !found {
		name = ""
	} else if err != nil {
		return nil, err
	}

	metaContainer := &pipeline.MetaContainer{
		Name:        name,
		Env:         container.Env,
		VolumeMount: container.VolumeMounts,
		EnvFrom:     container.EnvFrom,
		Data:        data,
	}

	return metaContainer, nil
}

func (a *application) SetMapping(mapping pipeline.WorkloadMapping) {
	a.resourceMapping = mapping
}

func (a *application) GroupVersionResource() schema.GroupVersionResource {
	return *a.gvr
}
