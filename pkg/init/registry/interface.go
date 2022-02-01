package registry

import (
	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/registry-support/registry-library/library"

	"github.com/redhat-developer/odo/pkg/util"
)

type Client interface {
	PullStackFromRegistry(registry string, stack string, destDir string, options library.RegistryOptions) error
	DownloadFileInMemory(params util.HTTPRequestParams) ([]byte, error)
	DownloadStarterProject(starterProject *devfilev1.StarterProject, decryptedToken string, contextDir string, verbose bool) error
}
