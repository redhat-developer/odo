package binding

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	devfilefs "github.com/devfile/library/pkg/testingutil/filesystem"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	backendpkg "github.com/redhat-developer/odo/pkg/binding/backend"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/libdevfile"
)

// ValidateRemoveBinding validates if the command has adequate arguments/flags
func (o *BindingClient) ValidateRemoveBinding(flags map[string]string) error {
	if flags[backendpkg.FLAG_NAME] == "" {
		return fmt.Errorf("you must specify the service binding name with --%s flag", backendpkg.FLAG_NAME)
	}
	return nil
}

// RemoveBinding removes the binding from devfile
func (o *BindingClient) RemoveBinding(servicebindingName string, obj parser.DevfileObj) (parser.DevfileObj, error) {
	var componentName string
	var options []string
	// Get all the K8s type devfile components
	components, err := obj.Data.GetComponents(common.DevfileOptions{
		ComponentOptions: common.ComponentOptions{ComponentType: v1alpha2.KubernetesComponentType},
	})
	if err != nil {
		return obj, err
	}
	for _, component := range components {
		var unstructuredObjs []unstructured.Unstructured
		// Parse the K8s manifest
		unstructuredObjs, err = libdevfile.GetK8sComponentAsUnstructuredList(obj, component.Name, filepath.Dir(obj.Ctx.GetAbsPath()), devfilefs.DefaultFs{})
		if err != nil || len(unstructuredObjs) == 0 {
			continue
		}
		// We default to the first object in the list because as far as ServiceBinding is concerned,
		// we assume that only one resource will be defined for the Devfile K8s component; which is true
		unstructuredObj := unstructuredObjs[0]
		if unstructuredObj.GetKind() == kclient.ServiceBindingKind {
			options = append(options, unstructuredObj.GetName())
			if unstructuredObj.GetName() == servicebindingName {
				componentName = component.Name
				break
			}
		}
	}
	if componentName == "" {
		return obj, fmt.Errorf("Service Binding %q not found in the devfile. Available Service Bindings: %s", servicebindingName, strings.Join(options, ", "))
	}

	err = obj.Data.DeleteComponent(componentName)
	return obj, err
}
