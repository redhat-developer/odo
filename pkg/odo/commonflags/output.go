package commonflags

import (
	"errors"
	"flag"
	"os"

	"github.com/redhat-developer/odo/pkg/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	// OutputFlagName is the name of the flag allowing user to specify output format
	OutputFlagName = "o"
)

func UseOutputFlag(cmd *cobra.Command) {
	if cmd.Annotations == nil {
		cmd.Annotations = map[string]string{}
	}
	cmd.Annotations["machineoutput"] = "json"
}

// AddOutputFlag adds the machine readable output flag to all commands
// We use "flag" in order to make this accessible throughtout ALL of odo, rather than the
// above traditional "persistentflags" usage that does not make it a pointer within the 'pflag'
// package
func AddOutputFlag() {
	flag.CommandLine.String(OutputFlagName, "", "Specify output format, supported format: json")
	_ = pflag.CommandLine.MarkHidden(OutputFlagName)
}

// CheckMachineReadableOutputCommand performs machine-readable output functions required to
// have it work correctly
func CheckMachineReadableOutputCommand(cmd *cobra.Command) error {

	// Get the needed values
	outputFlag := pflag.Lookup(OutputFlagName)
	hasFlagChanged := outputFlag != nil && outputFlag.Changed
	machineOutput := cmd.Annotations["machineoutput"]

	// Check the valid output
	if hasFlagChanged && outputFlag.Value.String() != "json" {
		//revive:disable:error-strings This is a top-level error message displayed as is to the end user
		return errors.New("Please input a valid output format for -o, available format: json")
		//revive:enable:error-strings
	}

	// Check that if -o json has been passed, that the command actually USES json.. if not, error out.
	if hasFlagChanged && outputFlag.Value.String() == "json" && machineOutput == "" {

		// By default we "disable" logging, so activate it so that the below error can be shown.
		_ = flag.Set(OutputFlagName, "")

		// Return the error
		//revive:disable:error-strings This is a top-level error message displayed as is to the end user
		return errors.New("Machine readable output is not yet implemented for this command")
		//revive:enable:error-strings
	}

	// Before running anything, we will make sure that no verbose output is made
	// This is a HACK to manually override `-v 4` to `-v 0` (in which we have no klog.V(0) in our code...
	// in order to have NO verbose output when combining both `-o json` and `-v 4` so json output
	// is not malformed / mixed in with normal logging
	if log.IsJSON() {
		_ = flag.Set("v", "0")
	} else {
		// Override the logging level by the value (if set) by the ODO_LOG_LEVEL env
		// The "-v" flag set on command line will take precedence over ODO_LOG_LEVEL env
		v := flag.CommandLine.Lookup("v").Value.String()
		if level, ok := os.LookupEnv("ODO_LOG_LEVEL"); ok && v == "0" {
			_ = flag.CommandLine.Set("v", level)
		}
	}
	return nil
}
