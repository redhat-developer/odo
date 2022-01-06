package component

import (
	"github.com/redhat-developer/odo/pkg/envinfo"
)

type Client interface {
	NewComponentFullDescriptionFromClientAndLocalConfigProvider(envInfo *envinfo.EnvSpecificInfo, componentName string, applicationName string, projectName string, context string) (*ComponentFullDescription, error)
	List(applicationSelector string) (ComponentList, error)
	ListDevfileComponentsInPath(paths []string) ([]Component, error)
	Exists(componentName, applicationName string) (bool, error)
	GetComponentNames(applicationName string) ([]string, error)
	CheckDefaultProject(name string) error
}
