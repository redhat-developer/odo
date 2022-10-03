package commonflags

import (
	"errors"
	"flag"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	// RunOnFlagName is the name of the flag allowing user to specify target platform
	RunOnFlagName = "run-on"
)

// UseRunOnFlag indicates that a command accepts the --run-on flag
func UseRunOnFlag(cmd *cobra.Command) {
	if cmd.Annotations == nil {
		cmd.Annotations = map[string]string{}
	}
	cmd.Annotations["runOn"] = "true"
}

// AddRunOnFlag adds the --run-on output flag to all commands
// We use "flag" in order to make this accessible throughtout ALL of odo, rather than the
// above traditional "persistentflags" usage that does not make it a pointer within the 'pflag'
// package
func AddRunOnFlag() {
	flag.CommandLine.String(RunOnFlagName, "", "Specify target platform, supported platforms: cluster(default), podman")
	_ = pflag.CommandLine.MarkHidden(RunOnFlagName)
}

// CheckRunOnCommand checks if commands enabling run-on flag are used correctly
func CheckRunOnCommand(cmd *cobra.Command) error {

	// Get the needed values
	runOnFlag := pflag.Lookup(RunOnFlagName)
	hasFlagChanged := runOnFlag != nil && runOnFlag.Changed
	runOn := cmd.Annotations["runOn"]

	// Check the valid output
	if hasFlagChanged && runOnFlag.Value.String() != "podman" && runOnFlag.Value.String() != "cluster" {
		return fmt.Errorf("%s is not a valid target platform for --run-on, please select either cluster (default) or podman", runOnFlag.Value.String())
	}

	// Check that if -o json has been passed, that the command actually USES json.. if not, error out.
	if hasFlagChanged && runOnFlag.Value.String() != "" && runOn == "" {
		return errors.New("--run-on flag is not supported for this command")
	}

	return nil
}
