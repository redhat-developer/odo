package util

import (
	"fmt"
	"os"
	"strings"

	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/machineoutput"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LogErrorAndExit prints the cause of the given error and exits the code with an
// exit code of 1.
// If the context is provided, then that is printed, if not, then the cause is
// detected using errors.Cause(err)
// *If* we are using the global json parameter, we instead output the json output
func LogErrorAndExit(err error, context string, a ...interface{}) {

	if err != nil {

		// If it's JSON, we'll output  the error
		if log.IsJSON() {

			// Machine readble error output
			machineOutput := machineoutput.GenericError{
				TypeMeta: metav1.TypeMeta{
					Kind:       machineoutput.ErrorKind,
					APIVersion: machineoutput.APIVersion,
				},
				Message: err.Error(),
			}

			// Output the error
			machineoutput.OutputError(machineOutput)

		} else {
			if context == "" {
				log.Error(errors.Cause(err))
			} else {
				printstring := fmt.Sprintf("%s%s", strings.Title(context), "\nError: %v")
				log.Errorf(printstring, err)
			}
		}

		// Always exit 1 anyways
		os.Exit(1)

	}
}

// CheckOutputFlag validates the -o flag
func CheckOutputFlag(outputFlag string) error {
	switch outputFlag {
	case "", "json":
		return nil
	default:
		return fmt.Errorf("Please input valid output format. available format: json")
	}

}

// GetFullName generates a command's full name based on its parent's full name and its own name
func GetFullName(parentName, name string) string {
	return parentName + " " + name
}

// CmdUsageTemplate is the main template used for all command line usage
var CmdUsageTemplate = `Usage:{{if .Runnable}}
  {{if .HasAvailableFlags}}{{appendIfNotPresent .UseLine "[flags]"}}{{else}}{{.UseLine}}{{end}}{{end}}{{if .HasAvailableSubCommands}}
  {{ .CommandPath}} [command]{{end}}{{if gt .Aliases 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{ .Example }}{{end}}{{ if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if .IsAvailableCommand}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{ if .HasAvailableLocalFlags}}

Flags:
{{CapitalizeFlagDescriptions .LocalFlags | trimRightSpace}}{{end}}{{ if .HasAvailableInheritedFlags}}

Additional Flags:
{{ModifyAdditionalFlags . | trimRightSpace}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsHelpCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{ if .HasAvailableSubCommands }}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
