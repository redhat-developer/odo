package libdevfile

import (
	"bytes"
	"io"
	
	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/v2/pkg/devfile/parser"
	"github.com/devfile/library/v2/pkg/devfile/parser/data/v2/common"
	devfilefs "github.com/devfile/library/v2/pkg/testingutil/filesystem"
	"github.com/ghodss/yaml"
	yaml3 "gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// GetK8sComponentAsUnstructuredList parses the Inlined/URI K8s of the Devfile K8s component and returns a list of unstructured.Unstructured objects;
// List is returned here because it is possible to define multiple K8s resources against a single Devfile K8s component
func GetK8sComponentAsUnstructuredList(devfileObj parser.DevfileObj, componentName string,
	context string, fs devfilefs.Filesystem) ([]unstructured.Unstructured, error) {

	strCRD, err := GetK8sManifestsWithVariablesSubstituted(devfileObj, componentName, context, fs)
	if err != nil {
		return nil, err
	}

	var uList []unstructured.Unstructured
	// Use the decoder to correctly read file with multiple manifests
	decoder := yaml3.NewDecoder(bytes.NewBufferString(strCRD))
	for {
		var decodeU unstructured.Unstructured
		if err = decoder.Decode(&decodeU.Object); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		// Marshal the object's data so that it can be unmarshalled again into unstructured.Unstructured object
		// We do this again because yaml3 "gopkg.in/yaml.v3" pkg is unable to properly unmarshal the data into an unstructured object
		rawData, err := yaml3.Marshal(decodeU.Object)
		if err != nil {
			return nil, err
		}

		// Use "github.com/ghodss/yaml" pkg to correctly unmarshal the data into an unstructured object;
		var u unstructured.Unstructured
		if err = yaml.Unmarshal(rawData, &u.Object); err != nil {
			return nil, err
		}

		uList = append(uList, u)

	}
	return uList, nil
}

// ListKubernetesComponents lists all the kubernetes components from the devfile
func ListKubernetesComponents(devfileObj parser.DevfileObj, path string) (list []unstructured.Unstructured, err error) {
	components, err := devfileObj.Data.GetComponents(common.DevfileOptions{
		ComponentOptions: common.ComponentOptions{ComponentType: v1alpha2.KubernetesComponentType},
	})
	if err != nil {
		return
	}
	var u []unstructured.Unstructured
	for _, kComponent := range components {
		if kComponent.Kubernetes != nil {
			u, err = GetK8sComponentAsUnstructuredList(devfileObj, kComponent.Name, path, devfilefs.DefaultFs{})
			if err != nil {
				return
			}
			list = append(list, u...)
		}
	}
	return
}

// AddKubernetesComponentToDevfile adds a resource definition to devfile object as an inlined Kubernetes component
func AddKubernetesComponentToDevfile(crd, name string, devfileObj parser.DevfileObj) (parser.DevfileObj, error) {
	err := devfileObj.Data.AddComponents([]v1alpha2.Component{{
		Name: name,
		ComponentUnion: v1alpha2.ComponentUnion{
			Kubernetes: &v1alpha2.KubernetesComponent{
				K8sLikeComponent: v1alpha2.K8sLikeComponent{
					BaseComponent: v1alpha2.BaseComponent{},
					K8sLikeComponentLocation: v1alpha2.K8sLikeComponentLocation{
						Inlined: crd,
					},
				},
			},
		},
	}})
	if err != nil {
		return devfileObj, err
	}

	return devfileObj, nil
}
