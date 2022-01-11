package component

import (
	"github.com/redhat-developer/odo/pkg/envinfo"
)

type Client interface {
	List(applicationSelector string) (ComponentList, error)
	ListComponentsInPath(paths []string) ([]Component, error)
	Exists(componentName, applicationName string) (bool, error)
	GetComponentNames(applicationName string) ([]string, error)
	GetComponentFullDescription(envInfo *envinfo.EnvSpecificInfo, componentName string, applicationName string, projectName string, context string) (*ComponentFullDescription, error)
	GetLinkedServicesSecretData(namespace, secretName string) (map[string][]byte, error)
	CheckDefaultProject(name string) error
}
