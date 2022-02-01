// cmdline package provides an abstration of a cmdline utility
// and an implementation with the corba library
package cmdline

import (
	"context"

	"github.com/redhat-developer/odo/pkg/kclient"
)

type Cmdline interface {
	// GetWorkingdirectory returns tehe directory on which the command should execute
	GetWorkingDirectory() (string, error)

	// GetFlags returns a map of flags set
	GetFlags() map[string]string

	// FlagValue returns the value for a flag
	FlagValue(flagName string) (string, error)

	// FlagValueIfSet returns the value for a flag, or an empty string if not set
	FlagValueIfSet(flagName string) string

	// IsFlagSet returns true if the flag is explicitely set
	IsFlagSet(flagName string) bool

	// CheckIfConfigurationNeeded checks against a set of commands that do *NOT* need configuration.
	CheckIfConfigurationNeeded() (bool, error)

	// Context returns the context attached to the command
	Context() context.Context

	// GetArgsAfterDashes returns the sub-array of args after `--`
	// returns an error if no args were passed after --
	GetArgsAfterDashes(args []string) ([]string, error)

	// GetParentName returns the name of the parent command or an empty string is there is no parent
	GetParentName() string

	// GetRootName returns the name of the root command
	GetRootName() string

	// GetName returns the name of the command
	GetName() string

	GetKubeClient() (kclient.ClientInterface, error)
}
