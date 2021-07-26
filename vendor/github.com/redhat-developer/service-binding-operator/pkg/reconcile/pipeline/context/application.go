package context

import (
	"errors"
	"fmt"
	"github.com/redhat-developer/service-binding-operator/apis/binding/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/converter"
	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"reflect"
	"strings"
)

const defaultContainerPath = "spec.template.spec.containers"

var _ pipeline.Application = &application{}

type application struct {
	gvr                    *schema.GroupVersionResource
	persistedResource      *unstructured.Unstructured
	resource               *unstructured.Unstructured
	bindingPath            *v1alpha1.BindingPath
	bindableContainerNames sets.String
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

func (a *application) ContainersPath() string {
	if a.bindingPath == nil || a.bindingPath.ContainersPath == "" {
		return defaultContainerPath
	}
	return a.bindingPath.ContainersPath
}

func (a *application) IsUpdated() bool {
	return !reflect.DeepEqual(a.persistedResource, a.resource)
}

func (a *application) BindableContainers() ([]map[string]interface{}, error) {
	path := strings.Split(a.ContainersPath(), ".")
	containers, found, err := converter.NestedResources(&corev1.Container{}, a.Resource().Object, path...)
	if !found {
		err = errors.New("no containers found in application resource")
	}
	if err != nil {
		return nil, err
	}
	initPath := append(path[:len(path)-1], "initContainers")
	initContainers, found, err := converter.NestedResources(&corev1.Container{}, a.Resource().Object, initPath...)
	if found && err == nil {
		containers = append(containers, initContainers...)
	}

	if len(a.bindableContainerNames) == 0 {
		return containers, err
	}
	filteredContainers := make([]map[string]interface{}, 0, len(containers))
	for _, c := range containers {
		cname, ok := c["name"]
		if ok && a.bindableContainerNames.Has(fmt.Sprintf("%v", cname)) {
			filteredContainers = append(filteredContainers, c)
		}
	}
	return filteredContainers, nil
}
