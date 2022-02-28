package libdevfile

import (
	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	devfilefs "github.com/devfile/library/pkg/testingutil/filesystem"
	"github.com/ghodss/yaml"
	"github.com/redhat-developer/odo/pkg/util"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// GetK8sComponentAsUnstructured parses the Inlined/URI K8s of the devfile K8s component
func GetK8sComponentAsUnstructured(component *v1alpha2.KubernetesComponent, context string, fs devfilefs.Filesystem) (unstructured.Unstructured, error) {
	strCRD := component.Inlined
	var err error
	if component.Uri != "" {
		strCRD, err = util.GetDataFromURI(component.Uri, context, fs)
		if err != nil {
			return unstructured.Unstructured{}, err
		}
	}

	// convert the YAML definition into map[string]interface{} since it's needed to create dynamic resource
	u := unstructured.Unstructured{}
	if err = yaml.Unmarshal([]byte(strCRD), &u.Object); err != nil {
		return unstructured.Unstructured{}, err
	}
	return u, nil
}

// ListKubernetesComponents lists all the kubernetes components from the devfile
func ListKubernetesComponents(devfileObj parser.DevfileObj, path string) (list []unstructured.Unstructured, err error) {
	components, err := devfileObj.Data.GetComponents(common.DevfileOptions{
		ComponentOptions: common.ComponentOptions{ComponentType: v1alpha2.KubernetesComponentType},
	})
	if err != nil {
		return
	}
	var u unstructured.Unstructured
	for _, kComponent := range components {
		if kComponent.Kubernetes != nil {
			u, err = GetK8sComponentAsUnstructured(kComponent.Kubernetes, path, devfilefs.DefaultFs{})
			if err != nil {
				return
			}
			list = append(list, u)
		}
	}
	return
}
