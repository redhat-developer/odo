package backend

import (
	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
)

// InitBackend is a specialized backend for steps of initiating a project, based on various input (either from CLI flags or interactively from user)
type InitBackend interface {
	// Validate returns an error if the backend can operate with the flags and do not validate their values
	Validate(flags map[string]string) error

	// SelectDevfile selects a devfile and returns its location information, depending on the flags
	// ok is false if the backend cannot operate
	SelectDevfile(flags map[string]string) (ok bool, location *DevfileLocation, err error)

	// SelectStarterProject selects a starter project from the devfile and returns information about the starter project,
	// depending on the flags
	// ok is false if the backend cannot operate
	SelectStarterProject(devfile parser.DevfileObj, flags map[string]string) (ok bool, starter *v1alpha2.StarterProject, err error)

	// PersonalizeName updates a devfile name, depending on the flags
	// ok is false if the backend cannot operate
	PersonalizeName(devfile parser.DevfileObj, flags map[string]string) (bool, error)
}
