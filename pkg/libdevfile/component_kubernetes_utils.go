package libdevfile

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	devfilefs "github.com/devfile/library/pkg/testingutil/filesystem"
	"github.com/ghodss/yaml"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/redhat-developer/odo/pkg/devfile/consts"
	"github.com/redhat-developer/odo/pkg/util"
)

// GetK8sComponentAsUnstructured parses the Inlined/URI K8s of the devfile K8s component
func GetK8sComponentAsUnstructured(devfileObj parser.DevfileObj, componentName string,
	context string, fs devfilefs.Filesystem) (unstructured.Unstructured, error) {

	strCRD, err := GetK8sManifestWithVariablesSubstituted(devfileObj, componentName, context, fs)
	if err != nil {
		return unstructured.Unstructured{}, err
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
			u, err = GetK8sComponentAsUnstructured(devfileObj, kComponent.Name, path, devfilefs.DefaultFs{})
			if err != nil {
				return
			}
			list = append(list, u)
		}
	}
	return
}

// AddKubernetesComponentToDevfile adds a resource definition to devfile as an inlined Kubernetes component
func AddKubernetesComponentToDevfile(crd, name string, devfileObj parser.DevfileObj) error {
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
		return err
	}

	return devfileObj.WriteYamlDevfile()
}

// AddKubernetesComponent adds the crd information to a separate file and adds the uri information to a devfile component
func AddKubernetesComponent(crd, name, componentContext string, devfile parser.DevfileObj) error {
	return addKubernetesComponent(crd, name, componentContext, devfile, devfilefs.DefaultFs{}, "odo-service")
}

// addKubernetesComponent adds the crd information to a separate file and adds the uri information to a devfile component
func addKubernetesComponent(crd, name, componentContext string, devfileObj parser.DevfileObj, fs devfilefs.Filesystem, filePrefix string) error {
	filePath := filepath.Join(componentContext, consts.UriFolder, filePrefix+name+".yaml")
	if _, err := fs.Stat(filepath.Join(componentContext, consts.UriFolder)); os.IsNotExist(err) {
		err = fs.MkdirAll(filepath.Join(componentContext, consts.UriFolder), os.ModePerm)
		if err != nil {
			return err
		}
	}

	if _, err := fs.Stat(filePath); !os.IsNotExist(err) {
		return fmt.Errorf("the file %q already exists", filePath)
	}

	err := fs.WriteFile(filePath, []byte(crd), 0755)
	if err != nil {
		return err
	}

	err = devfileObj.Data.AddComponents([]v1alpha2.Component{{
		Name: name,
		ComponentUnion: v1alpha2.ComponentUnion{
			Kubernetes: &v1alpha2.KubernetesComponent{
				K8sLikeComponent: v1alpha2.K8sLikeComponent{
					BaseComponent: v1alpha2.BaseComponent{},
					K8sLikeComponentLocation: v1alpha2.K8sLikeComponentLocation{
						Uri: filepath.Join(consts.UriFolder, filePrefix+name+".yaml"),
					},
				},
			},
		},
	}})
	if err != nil {
		return err
	}

	return devfileObj.WriteYamlDevfile()
}
