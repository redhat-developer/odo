// Package init provides methods to initiate an odo project.
// Most of the methods of the package get a `flags` parameter
// representing the flags passed from the user through the command line.
// Several backends are available to complete the operations, the backend
// being chosen depending on the flags content:
// - if no flags are passed, the `interactive` backend will be used
// - if some flags are passed, the `flags` backend will be used.
package init

import (
	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"

	"github.com/redhat-developer/odo/pkg/alizer"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
)

type Client interface {
	// GetFlags gets the flag specific to init operation so that it can correctly decide on the backend to be used
	// It ignores all the flags except the ones specific to init operation, for e.g. verbosity flag
	GetFlags(flags map[string]string) map[string]string
	// Validate checks for each backend if flags are valid
	Validate(flags map[string]string, fs filesystem.Filesystem, dir string) error

	// InitDevfile allows to initialize a Devfile in cases where this operation is needed as a prerequisite,
	// like if the directory contains no Devfile at all.
	// `preInitHandlerFunc` allows to perform operations prior to triggering the actual Devfile
	// initialization and personalization process.
	// `newDevfileHandlerFunc` is called only when a new Devfile object has been instantiated.
	// It allows to perform operations right after the Devfile has been initialized and personalized.
	// It is not called if the context directory already has a Devfile file.
	InitDevfile(flags map[string]string, contextDir string, preInitHandlerFunc func(interactiveMode bool),
		newDevfileHandlerFunc func(newDevfileObj parser.DevfileObj) error) error

	// SelectDevfile returns information about a devfile selected based on Alizer if the directory content,
	// or based on the flags if the directory is empty, or
	// interactively if flags is empty
	SelectDevfile(flags map[string]string, fs filesystem.Filesystem, dir string) (*alizer.DevfileLocation, error)

	// DownloadDevfile downloads a devfile given its location information and a destination directory
	// and returns the path of the downloaded file
	DownloadDevfile(devfileLocation *alizer.DevfileLocation, destDir string) (string, error)

	// SelectStarterProject selects a starter project from the devfile and returns information about the starter project,
	// depending on the flags. If not starter project is selected, a nil starter is returned
	SelectStarterProject(devfile parser.DevfileObj, flags map[string]string, fs filesystem.Filesystem, dir string) (*v1alpha2.StarterProject, error)

	// DownloadStarterProject downloads the starter project referenced in devfile and stores it in dest directory
	// WARNING: This will first remove all the content of dest.
	DownloadStarterProject(project *v1alpha2.StarterProject, dest string) error

	// PersonalizeName returns the customized Devfile Metadata Name.
	// Depending on the flags, it may return a name set interactively or not.
	PersonalizeName(devfile parser.DevfileObj, flags map[string]string) (string, error)

	// PersonalizeDevfileConfig updates the env vars, and URL endpoints
	PersonalizeDevfileConfig(devfileobj parser.DevfileObj, flags map[string]string, fs filesystem.Filesystem, dir string) (parser.DevfileObj, error)

	// SelectAndPersonalizeDevfile selects a devfile, then downloads, parse and personalize it
	// Returns the devfile object and its path
	SelectAndPersonalizeDevfile(flags map[string]string, contextDir string) (parser.DevfileObj, string, error)
}
