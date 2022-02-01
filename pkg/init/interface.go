package init

import "github.com/devfile/library/pkg/devfile/parser"

type Client interface {
	DownloadDirect(URL string, dest string) error
	DownloadFromRegistry(registryName string, devfile string, dest string) error
	DownloadStarterProject(devfile parser.DevfileObj, project string, dest string) error
}
