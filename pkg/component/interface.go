package component

import (
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/redhat-developer/odo/pkg/envinfo"
)

type Client interface {
	NewComponentFullDescriptionFromClientAndLocalConfigProvider(envInfo *envinfo.EnvSpecificInfo, componentName string, applicationName string, projectName string, context string) (*ComponentFullDescription, error)
	List(applicationSelector string) (ComponentList, error)
	ListDevfileComponentsInPath(paths []string) ([]Component, error)
	Exists(componentName, applicationName string) (bool, error)
	GetComponentNames(applicationName string) ([]string, error)
	GetComponentFromDevfile(info *envinfo.EnvSpecificInfo) (Component, parser.DevfileObj, error)
	GetComponentState(componentName, applicationName string) State
	GetComponent(componentName string, applicationName string) (component Component, err error)
	GetPushedComponents(applicationName string) (map[string]PushedComponent, error)
	GetPushedComponent(componentName, applicationName string) (PushedComponent, error)
	CheckDefaultProject(name string) error
}
