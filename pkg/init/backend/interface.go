// package backend provides different backends to initiate projects.
// - `Flags` backend gets needed information from command line flags.
// - `Interactive` backend interacts with the user to get needed information.
package backend

import (
	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/redhat-developer/odo/pkg/alizer"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
)

// InitBackend is a specialized backend for steps of initiating a project, based on various input (either from CLI flags or interactively from user)
type InitBackend interface {
	// Validate returns an error if it does not validate the flags based on the directory content
	Validate(flags map[string]string, fs filesystem.Filesystem, dir string) error

	// SelectDevfile selects a devfile and returns its location information, depending on the flags
	SelectDevfile(flags map[string]string, fs filesystem.Filesystem, dir string) (location *alizer.DevfileLocation, err error)

	// SelectStarterProject selects a starter project from the devfile and returns information about the starter project,
	// depending on the flags. If not starter project is selected, a nil starter is returned
	SelectStarterProject(devfile parser.DevfileObj, flags map[string]string) (starter *v1alpha2.StarterProject, err error)

	// PersonalizeName returns the customized Devfile Metadata Name.
	// Depending on the flags, it may return a name set interactively or not.
	PersonalizeName(devfile parser.DevfileObj, flags map[string]string) (string, error)

	// PersonalizeDevfileConfig updates the devfile config for ports and environment variables
	PersonalizeDevfileConfig(devfileobj parser.DevfileObj) (parser.DevfileObj, error)
}
