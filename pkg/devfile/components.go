package devfile

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/devfile/parser/data"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	parsercommon "github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	devfileFileSystem "github.com/devfile/library/pkg/testingutil/filesystem"
	devfilefs "github.com/devfile/library/pkg/testingutil/filesystem"
)

const (
	UriFolder  = "kubernetes"
	filePrefix = "odo-service-"
)

type InlinedComponent struct {
	Name    string
	Inlined string
}

type URIComponent struct {
	Name string
	URI  string
}

// SetupTestFolder for testing
func SetupTestFolder(testFolderName string, fs devfileFileSystem.Filesystem) (devfileFileSystem.File, error) {
	err := fs.MkdirAll(testFolderName, os.ModePerm)
	if err != nil {
		return nil, err
	}
	err = fs.MkdirAll(filepath.Join(testFolderName, UriFolder), os.ModePerm)
	if err != nil {
		return nil, err
	}
	testFileName, err := fs.Create(filepath.Join(testFolderName, UriFolder, "example.yaml"))
	if err != nil {
		return nil, err
	}
	return testFileName, nil
}

// GetDevfileData can be used to build a devfile structure for tests
func GetDevfileData(t *testing.T, inlined []InlinedComponent, uriComp []URIComponent) data.DevfileData {
	devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
	if err != nil {
		t.Error(err)
	}
	for _, component := range inlined {
		err = devfileData.AddComponents([]v1alpha2.Component{{
			Name: component.Name,
			ComponentUnion: devfilev1.ComponentUnion{
				Kubernetes: &devfilev1.KubernetesComponent{
					K8sLikeComponent: devfilev1.K8sLikeComponent{
						BaseComponent: devfilev1.BaseComponent{},
						K8sLikeComponentLocation: devfilev1.K8sLikeComponentLocation{
							Inlined: component.Inlined,
						},
					},
				},
			},
		},
		})
		if err != nil {
			t.Error(err)
		}
	}
	for _, component := range uriComp {
		err = devfileData.AddComponents([]v1alpha2.Component{{
			Name: component.Name,
			ComponentUnion: devfilev1.ComponentUnion{
				Kubernetes: &devfilev1.KubernetesComponent{
					K8sLikeComponent: devfilev1.K8sLikeComponent{
						BaseComponent: devfilev1.BaseComponent{},
						K8sLikeComponentLocation: devfilev1.K8sLikeComponentLocation{
							Uri: component.URI,
						},
					},
				},
			},
		},
		})
		if err != nil {
			t.Error(err)
		}
	}
	return devfileData
}

// GetComponentsToPush returns the list of Kubernetes components to push,
// by getting the list of Kubernetes components and removing the ones
// referenced from a command in the devfile
func GetKubernetesComponentsToPush(devfileObj parser.DevfileObj) ([]devfilev1.Component, error) {
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
		} else if command.Apply != nil {
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

// IsComponentDefined checks if a component with the given name is defined in a DevFile
func IsComponentDefined(name string, devfileObj parser.DevfileObj) (bool, error) {
	components, err := devfileObj.Data.GetComponents(common.DevfileOptions{})
	if err != nil {
		return false, err
	}
	for _, c := range components {
		if c.Name == name {
			return true, nil
		}
	}
	return false, nil
}

// AddKubernetesComponentToDevfile adds service definition to devfile as an inlined Kubernetes component
func AddKubernetesComponentToDevfile(crd, name string, devfileObj parser.DevfileObj) error {
	err := devfileObj.Data.AddComponents([]devfilev1.Component{{
		Name: name,
		ComponentUnion: devfilev1.ComponentUnion{
			Kubernetes: &devfilev1.KubernetesComponent{
				K8sLikeComponent: devfilev1.K8sLikeComponent{
					BaseComponent: devfilev1.BaseComponent{},
					K8sLikeComponentLocation: devfilev1.K8sLikeComponentLocation{
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
	return addKubernetesComponent(crd, name, componentContext, devfile, devfilefs.DefaultFs{})
}

// AddKubernetesComponent adds the crd information to a separate file and adds the uri information to a devfile component
func addKubernetesComponent(crd, name, componentContext string, devfileObj parser.DevfileObj, fs devfilefs.Filesystem) error {
	filePath := filepath.Join(componentContext, UriFolder, filePrefix+name+".yaml")
	if _, err := fs.Stat(filepath.Join(componentContext, UriFolder)); os.IsNotExist(err) {
		err = fs.MkdirAll(filepath.Join(componentContext, UriFolder), os.ModePerm)
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

	err = devfileObj.Data.AddComponents([]devfilev1.Component{{
		Name: name,
		ComponentUnion: devfilev1.ComponentUnion{
			Kubernetes: &devfilev1.KubernetesComponent{
				K8sLikeComponent: devfilev1.K8sLikeComponent{
					BaseComponent: devfilev1.BaseComponent{},
					K8sLikeComponentLocation: devfilev1.K8sLikeComponentLocation{
						Uri: filepath.Join(UriFolder, filePrefix+name+".yaml"),
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

// DeleteKubernetesComponentFromDevfile deletes an inlined Kubernetes component from devfile, if one exists
func DeleteKubernetesComponentFromDevfile(name string, devfileObj parser.DevfileObj, componentContext string) error {
	return deleteKubernetesComponentFromDevfile(name, devfileObj, componentContext, devfilefs.DefaultFs{})
}

// deleteKubernetesComponentFromDevfile deletes an inlined Kubernetes component from devfile, if one exists
func deleteKubernetesComponentFromDevfile(name string, devfileObj parser.DevfileObj, componentContext string, fs devfilefs.Filesystem) error {
	components, err := devfileObj.Data.GetComponents(common.DevfileOptions{})
	if err != nil {
		return err
	}

	found := false
	for _, c := range components {
		if c.Name == name {
			err = devfileObj.Data.DeleteComponent(c.Name)
			if err != nil {
				return err
			}

			if c.Kubernetes.Uri != "" {
				parsedURL, err := url.Parse(c.Kubernetes.Uri)
				if err != nil {
					return err
				}
				if len(parsedURL.Host) == 0 || len(parsedURL.Scheme) == 0 {
					err := fs.Remove(filepath.Join(componentContext, c.Kubernetes.Uri))
					if err != nil {
						return err
					}
				}
			}
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("could not find the service %q in devfile", name)
	}

	return devfileObj.WriteYamlDevfile()
}
