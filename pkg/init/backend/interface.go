// package backend provides different backends to initiate projects.
// - `Flags` backend gets needed information from command line flags.
// - `Interactive` backend interacts with the user to get needed information.
package backend

import (
	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
)

// InitBackend is a specialized backend for steps of initiating a project, based on various input (either from CLI flags or interactively from user)
type InitBackend interface {
	// Validate returns an error if it does not validate the flags
	Validate(flags map[string]string) error

	// SelectDevfile selects a devfile and returns its location information, depending on the flags
	SelectDevfile(flags map[string]string) (location *DevfileLocation, err error)

	// SelectStarterProject selects a starter project from the devfile and returns information about the starter project,
	// depending on the flags. If not starter project is selected, a nil starter is returned
	SelectStarterProject(devfile parser.DevfileObj, flags map[string]string) (starter *v1alpha2.StarterProject, err error)

	// PersonalizeName updates a devfile name, depending on the flags
	PersonalizeName(devfile parser.DevfileObj, flags map[string]string) error

	// PersonalizeDevfileConfig updates the env vars, and URL endpoints
	PersonalizeDevfileConfig(devfile parser.DevfileObj) error
}
