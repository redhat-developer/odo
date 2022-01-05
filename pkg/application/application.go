package application

import "github.com/redhat-developer/odo/pkg/component"

type Client interface {
	List() ([]string, error)
	Exists(app string) (bool, error)
	Delete(name string) error
	ComponentList(name string) ([]component.Component, error)
	GetMachineReadableFormat(appName string, projectName string) App
	GetMachineReadableFormatForList(apps []App) AppList
}
