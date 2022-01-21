package registry

import (
	"github.com/devfile/registry-support/registry-library/library"
	"github.com/redhat-developer/odo/pkg/util"
)

type Client interface {
	PullStackFromRegistry(registry string, stack string, destDir string, options library.RegistryOptions) error
	DownloadFileInMemory(params util.HTTPRequestParams) ([]byte, error)
}
