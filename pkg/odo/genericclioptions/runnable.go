package genericclioptions

import (
	"flag"
	"fmt"
	"os"

	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/odo/util"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type Runnable interface {
	Complete(name string, cmd *cobra.Command, args []string) error
	Validate() error
	Run() error
}

func GenericRun(o Runnable, cmd *cobra.Command, args []string) {

	// CheckMachineReadableOutput
	// fixes / checks all related machine readable output functions
	CheckMachineReadableOutputCommand(cmd)

	// LogErrorAndExit is used so that we get -o (jsonoutput) for cmds which have json output implemented
	util.LogErrorAndExit(checkConflictingFlags(cmd), "")
	// Run completion, validation and run.
	util.LogErrorAndExit(o.Complete(cmd.Name(), cmd, args), "")
	util.LogErrorAndExit(o.Validate(), "")
	util.LogErrorAndExit(o.Run(), "")
}

// checkConflictingFlags checks for conflicting flags. Currently --context cannot be provided
// with either --app, --project and --component as that information can be fetched from the local
// config.
func checkConflictingFlags(cmd *cobra.Command) error {

	// we allow providing --context with --app and --project in case of `odo create` or `odo component create`
	if cmd.Name() == "create" {
		if cmd.HasParent() {
			if cmd.Parent().Name() == "odo" || cmd.Parent().Name() == "component" {
				return nil
			}
		}
	}
	app := stringFlagLookup(cmd, "app")
	project := stringFlagLookup(cmd, "project")
	context := stringFlagLookup(cmd, "context")
	component := stringFlagLookup(cmd, "component")
	if (context != "") && (app != "" || project != "" || component != "") {
		return fmt.Errorf("cannot provide --app, --project or --component flag when --context is provided")
	}
	return nil
}
func stringFlagLookup(cmd *cobra.Command, flagName string) string {
	flag := cmd.Flags().Lookup(flagName)
	// a check to make sure if the flag is not defined we return blank
	if flag == nil {
		return ""
	}
	return flag.Value.String()
}

// CheckMachineReadableOutputCommand performs machine-readable output functions required to
// have it work correctly
func CheckMachineReadableOutputCommand(cmd *cobra.Command) {

	// Get the needed values
	outputFlag := pflag.Lookup("o")
	hasFlagChanged := outputFlag != nil && outputFlag.Changed
	machineOutput := cmd.Annotations["machineoutput"]

	// Check the valid output
	if hasFlagChanged && outputFlag.Value.String() != "json" {
		log.Error("Please input a valid output format for -o, available format: json")
		os.Exit(1)
	}

	// Check that if -o json has been passed, that the command actually USES json.. if not, error out.
	if hasFlagChanged && outputFlag.Value.String() == "json" && machineOutput == "" {

		// By default we "disable" logging, so undisable it so that the below error can be shown.
		_ = flag.Set("o", "")

		// Output the error
		log.Error("Machine readable output is not yet implemented for this command")
		os.Exit(1)
	}

	// Before running anything, we will make sure that no verbose output is made
	// This is a HACK to manually override `-v 4` to `-v 0` (in which we have no klog.V(0) in our code...
	// in order to have NO verbose output when combining both `-o json` and `-v 4` so json output
	// is not malformed / mixed in with normal logging
	if log.IsJSON() {
		_ = flag.Set("v", "0")
	}
}
