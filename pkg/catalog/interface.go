package catalog

import (
	"github.com/redhat-developer/odo/pkg/kclient"
)

type Client interface {
	GetDevfileRegistries(registryName string) ([]Registry, error)
	ListDevfileComponents(registryName string) (DevfileComponentTypeList, error)
	GetStarterProjectsNames(details DevfileComponentType) ([]string, error)
	SearchComponent(client kclient.ClientInterface, name string) ([]string, error)
}
