// package registry wraps various package level functions into a Client interface to be able to mock them
package registry

import (
	"context"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	dfutil "github.com/devfile/library/v2/pkg/util"
	"github.com/devfile/registry-support/registry-library/library"
	"github.com/redhat-developer/odo/pkg/api"
)

type Client interface {
	PullStackFromRegistry(registry string, stack string, destDir string, options library.RegistryOptions) error
	DownloadFileInMemory(params dfutil.HTTPRequestParams) ([]byte, error)
	DownloadStarterProject(starterProject *devfilev1.StarterProject, decryptedToken string, contextDir string, verbose bool) error
	GetDevfileRegistries(registryName string) ([]api.Registry, error)
	ListDevfileStacks(ctx context.Context, registryName, devfileFlag, filterFlag string, detailsFlag bool) (DevfileStackList, error)
}
