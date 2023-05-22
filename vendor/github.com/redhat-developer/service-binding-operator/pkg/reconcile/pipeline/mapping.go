/*
Copyright 2022.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package pipeline

import (
	"fmt"
	"path"

	"github.com/redhat-developer/service-binding-operator/apis/spec/v1beta1"
	"github.com/redhat-developer/service-binding-operator/pkg/converter"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/jsonpath"
)

type WorkloadContainer struct {
	Path         *jsonpath.JSONPath
	Name         []string
	Env          []string
	EnvFrom      []string
	VolumeMounts []string
}

type WorkloadMapping struct {
	Containers []WorkloadContainer
	Volume     []string
}

func FromWorkloadResourceMappingTemplate(mapping v1beta1.ClusterWorkloadResourceMappingTemplate) (*WorkloadMapping, error) {
	var containers []WorkloadContainer
	for _, container := range mapping.Containers {
		wc := WorkloadContainer{}

		if err := constructRestrictedPath(container.Name, []string{"name"}, &wc.Name); err != nil {
			return nil, err
		}
		if err := constructRestrictedPath(container.Env, []string{"env"}, &wc.Env); err != nil {
			return nil, err
		}
		if err := constructRestrictedPath(container.VolumeMounts, []string{"volumeMounts"}, &wc.VolumeMounts); err != nil {
			return nil, err
		}
		// not in the spec, but we need this for compatability with the coreos api
		wc.EnvFrom = []string{"envFrom"}

		path := jsonpath.New("")
		formatted := fmt.Sprintf("{%s}", container.Path)
		if err := path.Parse(formatted); err != nil {
			return nil, fmt.Errorf("Error in parsing JSONPath expression %v: %v", formatted, err)
		}

		wc.Path = path

		containers = append(containers, wc)
	}

	var volume []string
	if err := constructRestrictedPath(mapping.Volumes, []string{"spec", "template", "spec", "volumes"}, &volume); err != nil {
		return nil, err
	}

	return &WorkloadMapping{Containers: containers, Volume: volume}, nil
}

func constructRestrictedPath(value string, defaultValue []string, target *[]string) error {
	if value != "" {
		val, err := isValidRestrictedJsonPath(value)
		if err != nil {
			return err
		}
		*target = val
	} else {
		*target = defaultValue
	}
	return nil
}

func isValidRestrictedJsonPath(path string) ([]string, error) {
	parser, err := jsonpath.Parse("", fmt.Sprintf("{%s}", path))
	if err != nil {
		return nil, err
	}
	return verifyJsonPath(parser.Root)
}

func verifyJsonPath(node jsonpath.Node) ([]string, error) {
	switch node.Type() {
	case jsonpath.NodeField:
		field := node.(*jsonpath.FieldNode)
		return []string{field.Value}, nil
	case jsonpath.NodeList:
		list := node.(*jsonpath.ListNode)
		var paths []string
		for _, node := range list.Nodes {
			nested, err := verifyJsonPath(node)
			if err != nil {
				return nil, err
			}
			paths = append(paths, nested...)
		}
		return paths, nil
	default:
		return nil, fmt.Errorf("Node type %q not allowed in restricted JSONPath contexts", node)
	}
}

const bindingRootEnvVar = "SERVICE_BINDING_ROOT"
const bindingRoot = "/bindings"

func (container *MetaContainer) MountPath(bindingName string) (string, error) {
	envs, found, err := converter.NestedResources(&corev1.EnvVar{}, container.Data, container.Env...)
	if err != nil {
		return "", err
	} else if found {
		for _, e := range envs {
			var envVar corev1.EnvVar
			if err := runtime.DefaultUnstructuredConverter.FromUnstructured(e, &envVar); err != nil {
				continue // should be unreachable
			}
			if envVar.Name == bindingRootEnvVar {
				return path.Join(envVar.Value, bindingName), nil
			}
		}
	}

	mp := path.Join(bindingRoot, bindingName)

	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&corev1.EnvVar{
		Name:  bindingRootEnvVar,
		Value: bindingRoot,
	})
	if err != nil {
		return "", err
	}

	if found {
		envs = append(envs, u)
	} else {
		envs = []map[string]interface{}{u}
	}
	if err := setSlice(container.Data, envs, container.Env); err != nil {
		return "", err
	}
	return mp, nil
}

// We can't use unstructured.SetNestedField when we try to set a field to a value of type
// []map[string]interface{}, since (apparently) map[string]interface{} doesn't implement interface{}.
// Runtime panics occur when using SetNestedField, and SetNesetedSlice gives a compile-time error.
func setSlice(obj map[string]interface{}, resources []map[string]interface{}, path []string) error {
	dest := obj
	if len(path) > 1 {
		slice, found, err := unstructured.NestedFieldNoCopy(obj, path[:len(path)-1]...)
		if err != nil {
			return err
		} else if !found {
			current := obj
			for i, p := range path[:len(path)-1] {
				x, exists := current[p]
				y, ok := x.(map[string]interface{})
				if !exists {
					y = make(map[string]interface{})
					current[p] = y
				} else if !ok {
					return fmt.Errorf("Value already exists at path %v", path[:i])
				}
				current = y
			}

			dest = current
		} else {
			x, ok := slice.(map[string]interface{})
			if !ok {
				return fmt.Errorf("Value already exists at path %v", path[:len(path)-1])
			}
			dest = x
		}
	}

	key := path[len(path)-1]
	dest[key] = resources
	return nil
}

func addRaw(obj map[string]interface{}, resource interface{}, data []map[string]interface{}, path []string) error {
	nestedData, found, err := converter.NestedResources(resource, obj, path...)
	if err != nil {
		return err
	}

	if !found {
		nestedData = data
	} else {
		nestedData = append(nestedData, data...)
	}
	return setSlice(obj, nestedData, path)
}

func (container *MetaContainer) AddEnvVars(vars []corev1.EnvVar) error {
	for _, v := range vars {
		// ignore errors, we could be projecting env vars that don't exist yet
		_ = container.RemoveEnvVars(v.Name)
	}
	var data []map[string]interface{}
	for _, variable := range vars {
		x, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&variable)
		if err != nil {
			return err
		}
		data = append(data, x)
	}
	return addRaw(container.Data, &corev1.EnvVar{}, data, container.Env)
}

func (container *MetaContainer) RemoveEnvVars(name string) error {
	envFrom, found, err := converter.NestedResources(&corev1.EnvVar{}, container.Data, container.Env...)
	if err != nil {
		return err
	} else if !found {
		return nil
	}
	for i, envSource := range envFrom {
		if val, found, err := unstructured.NestedString(envSource, "name"); found && err == nil && val == name {
			s := append(envFrom[:i], envFrom[i+1:]...)
			if err := setSlice(container.Data, s, container.Env); err != nil {
				return err
			}
			break
		}
	}
	return nil
}

func (container *MetaContainer) AddEnvFromVar(envVar corev1.EnvFromSource) error {
	if len(container.EnvFrom) != 0 {
		// ignore errors, we could be projecting env vars that don't exist yet
		_ = container.RemoveEnvFromVars(envVar.SecretRef.Name)
		data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&envVar)
		if err != nil {
			return err
		}
		return addRaw(container.Data, &corev1.EnvFromSource{}, []map[string]interface{}{data}, container.EnvFrom)
	}
	return fmt.Errorf("No envFrom field")
}

func (container *MetaContainer) RemoveEnvFromVars(secretName string) error {
	envFrom, found, err := converter.NestedResources(&corev1.EnvFromSource{}, container.Data, container.EnvFrom...)
	if err != nil {
		return err
	} else if !found {
		return nil
	}
	for i, envSource := range envFrom {
		if val, found, err := unstructured.NestedString(envSource, "secretRef", "name"); found && err == nil && val == secretName {
			s := append(envFrom[:i], envFrom[i+1:]...)
			if err := setSlice(container.Data, s, container.EnvFrom); err != nil {
				return err
			}
			break
		}
	}
	return nil
}

func (container *MetaContainer) AddVolumeMount(mount corev1.VolumeMount) error {
	var newVolumes []map[string]interface{}
	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&mount)
	if err != nil {
		return err
	}

	volumes, found, err := converter.NestedResources(&corev1.VolumeMount{}, container.Data, container.VolumeMount...)
	if err != nil {
		return err
	} else if !found {
		newVolumes = append(newVolumes, u)
	} else {
		exist := false
		for _, vol := range volumes {
			var tmpVol corev1.VolumeMount
			if err := runtime.DefaultUnstructuredConverter.FromUnstructured(vol, &tmpVol); err != nil {
				return err
			}
			if tmpVol.Name == mount.Name {
				exist = true
				newVolumes = append(newVolumes, u)
			} else {
				newVolumes = append(newVolumes, vol)
			}
		}
		if !exist {
			newVolumes = append(newVolumes, u)
		}
	}

	return setSlice(container.Data, newVolumes, container.VolumeMount)
}

func (container *MetaContainer) RemoveVolumeMount(name string) error {
	volumeMounts, found, err := converter.NestedResources(&corev1.VolumeMount{}, container.Data, container.VolumeMount...)
	if err != nil {
		return err
	} else if !found {
		return nil
	}

	for i, vm := range volumeMounts {
		mount := &corev1.VolumeMount{}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(vm, mount); err == nil && mount.Name == name {
			s := append(volumeMounts[:i], volumeMounts[i+1:]...)
			if err := setSlice(container.Data, s, container.VolumeMount); err != nil {
				return err
			}
			break
		}
	}
	return nil
}

func (template *MetaPodSpec) AddVolume(volume corev1.Volume) error {
	var newVolumes []map[string]interface{}
	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&volume)
	if err != nil {
		return err
	}

	volumes, found, err := converter.NestedResources(&corev1.Volume{}, template.Data, template.Volume...)
	if err != nil {
		return err
	} else if !found {
		newVolumes = append(newVolumes, u)
	} else {
		exist := false
		for _, vol := range volumes {
			if vol["name"] == volume.Name {
				exist = true
				newVolumes = append(newVolumes, u)
			} else {
				newVolumes = append(newVolumes, vol)
			}
		}
		if !exist {
			newVolumes = append(newVolumes, u)
		}
	}

	return setSlice(template.Data, newVolumes, template.Volume)
}

func (template *MetaPodSpec) RemoveVolume(name string) error {
	volumeResources, found, err := converter.NestedResources(&corev1.Volume{}, template.Data, template.Volume...)
	if err != nil {
		return err
	} else if !found {
		return nil
	}
	for i, vol := range volumeResources {
		if val, found, err := unstructured.NestedString(vol, "name"); found && err == nil && val == name {
			s := append(volumeResources[:i], volumeResources[i+1:]...)
			if err := setSlice(template.Data, s, template.Volume); err != nil {
				return err
			}
			break
		}
	}
	return nil
}
