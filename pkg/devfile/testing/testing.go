package testing

import (
	"os"
	"path/filepath"
	"testing"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser/data"
	devfileFileSystem "github.com/devfile/library/pkg/testingutil/filesystem"
	"github.com/redhat-developer/odo/pkg/devfile/consts"
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
	err = fs.MkdirAll(filepath.Join(testFolderName, consts.UriFolder), os.ModePerm)
	if err != nil {
		return nil, err
	}
	testFileName, err := fs.Create(filepath.Join(testFolderName, consts.UriFolder, "example.yaml"))
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
		err = devfileData.AddComponents([]devfilev1.Component{{
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
		err = devfileData.AddComponents([]devfilev1.Component{{
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
