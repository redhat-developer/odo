package registry

import (
	"github.com/devfile/registry-support/registry-library/library"
	"github.com/redhat-developer/odo/pkg/util"
)

type RegistryClient struct{}

func NewRegistryClient() RegistryClient {
	return RegistryClient{}
}

func (o RegistryClient) PullStackFromRegistry(registry string, stack string, destDir string, options library.RegistryOptions) error {
	return library.PullStackFromRegistry(registry, stack, destDir, options)
}

func (o RegistryClient) DownloadFileInMemory(params util.HTTPRequestParams) ([]byte, error) {
	return util.DownloadFileInMemory(params)
}
