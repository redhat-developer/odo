package registry

import (
	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/registry-support/registry-library/library"

	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/util"
)

type RegistryClient struct{}

func NewRegistryClient() RegistryClient {
	return RegistryClient{}
}

// PullStackFromRegistry pulls stack from registry with all stack resources (all media types) to the destination directory
func (o RegistryClient) PullStackFromRegistry(registry string, stack string, destDir string, options library.RegistryOptions) error {
	return library.PullStackFromRegistry(registry, stack, destDir, options)
}

// DownloadFileInMemory uses the url to download the file and return bytes
func (o RegistryClient) DownloadFileInMemory(params util.HTTPRequestParams) ([]byte, error) {
	return util.DownloadFileInMemory(params)
}

// DownloadStarterProject downloads a starter project referenced in devfile
// This will first remove the content of the contextDir
func (o RegistryClient) DownloadStarterProject(starterProject *devfilev1.StarterProject, decryptedToken string, contextDir string, verbose bool) error {
	return component.DownloadStarterProject(starterProject, decryptedToken, contextDir, verbose)
}
