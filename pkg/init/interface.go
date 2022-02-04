package init

import (
	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"

	"github.com/redhat-developer/odo/pkg/init/backend"
)

type Client interface {
	// Validate checks for each backend if flags are valid
	Validate(flags map[string]string) error

	// SelectDevfile returns information about a devfile selected based on the flags, or
	// interactively if flags is empty
	SelectDevfile(flags map[string]string) (*backend.DevfileLocation, error)

	// DownloadDevfile downloads a devfile given its location information and a destination directory
	// and returns the path of the downloaded file
	DownloadDevfile(devfileLocation *backend.DevfileLocation, destDir string) (string, error)

	// SelectStarterProject selects a starter project from the devfile and returns information about the starter project,
	// depending on the flags
	SelectStarterProject(devfile parser.DevfileObj, flags map[string]string) (*v1alpha2.StarterProject, error)

	// DownloadStarterProject downloads the starter project referenced in devfile and stores it in dest directory
	// WARNING: This will first remove all the content of dest.
	DownloadStarterProject(project *v1alpha2.StarterProject, dest string) error

	// PersonalizeName updates a devfile name, depending on the flags
	PersonalizeName(devfile parser.DevfileObj, flags map[string]string) error
}
