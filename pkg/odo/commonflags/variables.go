package commonflags

import (
	"errors"
	"os"

	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
	"github.com/redhat-developer/odo/pkg/vars"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	VarFileFlagName = "var-file"
	VarFlagName     = "var"
)

// UseVariablesFlags indicates that a command accepts the --var-file and --var flags
func UseVariablesFlags(cmd *cobra.Command) {
	if cmd.Annotations == nil {
		cmd.Annotations = map[string]string{}
	}
	cmd.Annotations["variables"] = "true"
}

// AddVariablesFlags adds the --var-file and --var flags to all commands
// We use "flag" in order to make this accessible throughtout ALL of odo, rather than the
// above traditional "persistentflags" usage that does not make it a pointer within the 'pflag'
// package
func AddVariablesFlags() {
	pflag.CommandLine.StringArray(VarFlagName, []string{}, "Variable to override Devfile variable and variables in var-file")
	pflag.CommandLine.String(VarFileFlagName, "", "File containing variables to override Devfile variables")
}

// CheckVariablesCommand checks if commands enabling --var-file and --var flags are used correctly
func CheckVariablesCommand(cmd *cobra.Command) error {
	// Get the needed values
	varFileFlag := pflag.Lookup(VarFileFlagName)
	varFlag := pflag.Lookup(VarFlagName)
	hasFlagChanged := (varFileFlag != nil && varFileFlag.Changed) || (varFlag != nil && varFlag.Changed)
	supportVariablesFlags := cmd.Annotations["variables"] == "true"

	// Check that if flags have been used, the command supports them
	if hasFlagChanged && !supportVariablesFlags {
		return errors.New("--var-file and --var flags are not supported for this command")
	}

	return nil
}

// GetVariablesValues returns variables computed from --var-file and --var values
func GetVariablesValues(cmd cmdline.Cmdline) (map[string]string, error) {
	varFileFlagValue, err := cmd.FlagValue(VarFileFlagName)
	if err != nil {
		return nil, err
	}
	varFlagValue, err := cmd.FlagValues(VarFlagName)
	if err != nil {
		return nil, err
	}
	return vars.GetVariables(filesystem.DefaultFs{}, varFileFlagValue, varFlagValue, os.LookupEnv)
}
