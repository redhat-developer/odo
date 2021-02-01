package envinfo

import (
	"fmt"
	devfile "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
)

// AddServiceToDevfile adds service definition to devfile as an inlined Kubernetes component
func (esi *EnvSpecificInfo) AddServiceToDevfile(crd, name string) error {
	err := esi.devfileObj.Data.AddComponents([]devfile.Component{
		{
			Name: name,
			ComponentUnion: devfile.ComponentUnion{
				Kubernetes: &devfile.KubernetesComponent{
					K8sLikeComponent: devfile.K8sLikeComponent{
						BaseComponent: devfile.BaseComponent{},
						K8sLikeComponentLocation: devfile.K8sLikeComponentLocation{
							Inlined: crd,
						},
					},
				},
			},
		},
	})
	if err != nil {
		return err
	}

	return esi.devfileObj.WriteYamlDevfile()
}

// DeleteServiceFromDevfile deletes an inlined Kubernetes component from devfile, if one exists
func (esi *EnvSpecificInfo) DeleteServiceFromDevfile(name string) error {
	components, err := esi.devfileObj.Data.GetComponents(common.DevfileOptions{})
	if err != nil {
		return err
	}

	found := false
	for _, c := range components {
		if c.Name == name {
			err = esi.devfileObj.Data.DeleteComponent(c.Name)
			if err != nil {
				return err
			}
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("could not find the service %q in devfile", name)
	}

	return esi.devfileObj.WriteYamlDevfile()
}
