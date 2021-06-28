package util

import (
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/openshift/odo/pkg/component"
	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/localConfigProvider"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/machineoutput"
	"github.com/openshift/odo/pkg/occlient"
	"github.com/openshift/odo/pkg/storage"
	urlPkg "github.com/openshift/odo/pkg/url"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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
					Kind:       machineoutput.Kind,
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

// PrintComponentInfo prints Component Information like path, URL & storage
func PrintComponentInfo(client *occlient.Client, currentComponentName string, componentDesc component.Component, applicationName string, project string) error {

	log.Describef("Component Name: ", currentComponentName)
	log.Describef("Type: ", componentDesc.Spec.Type)

	// Source
	if componentDesc.Spec.Source != "" {
		log.Describef("Source: ", componentDesc.Spec.Source)
	}

	// Env
	if componentDesc.Spec.Env != nil {

		// Retrieve all the environment variables
		var output string
		for _, env := range componentDesc.Spec.Env {
			output += fmt.Sprintf(" · %v=%v\n", env.Name, env.Value)
		}

		// Cut off the last newline and output
		if len(output) > 0 {
			output = output[:len(output)-1]
			log.Describef("Environment Variables:\n", output)
		}

	}

	// Storage
	if len(componentDesc.Spec.Storage) > 0 {

		var storages storage.StorageList

		if componentDesc.Status.State == "Pushed" {
			// Retrieve the storage list
			storages = storage.StorageList{Items: componentDesc.Spec.StorageSpec}
		} else {
			localConfig, err := config.New()
			LogErrorAndExit(err, "")
			storageLocal, err := localConfig.ListStorage()
			if err != nil {
				return err
			}
			storages = storage.ConvertListLocalToMachine(storageLocal)

		}

		// Gather the output
		var output string
		for _, store := range storages.Items {
			output += fmt.Sprintf(" · %v of size %v mounted to %v\n", store.Name, store.Spec.Size, store.Spec.Path)
		}

		// Cut off the last newline and output
		if len(output) > 0 {
			output = output[:len(output)-1]
			log.Describef("Storage:\n", output)
		}

	}

	// URL
	if componentDesc.Spec.URL != nil {
		var output string

		// S2I Only (odo describe exits by default on Devfile by default anyways..)
		// if the component is not pushed
		if componentDesc.Status.State == component.StateTypeNotPushed {
			// Gather the output
			for i, componentURL := range componentDesc.Spec.URL {
				output += fmt.Sprintf(" · URL named %s will be exposed via %v\n", componentURL, componentDesc.Spec.Ports[i])
			}
		} else {
			// Retrieve the URLs
			urls := urlPkg.URLList{Items: componentDesc.Spec.URLSpec}

			// Gather the output
			for _, componentURL := range componentDesc.Spec.URL {
				url := urls.Get(componentURL)

				var urlString string

				switch url.Spec.Kind {
				case localConfigProvider.ROUTE:
					urlString = urlPkg.GetURLString(url.Spec.Protocol, url.Spec.Host, "", false)
				case localConfigProvider.INGRESS:
					urlString = urlPkg.GetURLString(url.Spec.Protocol, "", url.Spec.Host, false)
				default:
					continue
				}

				output += fmt.Sprintf(" · %v exposed via %v\n", urlString, url.Spec.Port)
			}

		}

		// Cut off the last newline and output
		if len(output) > 0 {
			output = output[:len(output)-1]
			log.Describef("URLs:\n", output)
		}

	}

	// Linked services
	if len(componentDesc.Status.LinkedServices) > 0 {

		// Gather the output
		var output string
		for _, linkedService := range componentDesc.Status.LinkedServices {

			// Let's also get the secrets / environment variables that are being passed in.. (if there are any)
			secrets, err := client.GetKubeClient().GetSecret(linkedService.SecretName, project)
			LogErrorAndExit(err, "")

			if len(secrets.Data) > 0 {
				// Iterate through the secrets to throw in a string
				var secretOutput string
				for i := range secrets.Data {
					secretOutput += fmt.Sprintf("    · %v\n", i)
				}

				if len(secretOutput) > 0 {
					// Cut off the last newline
					secretOutput = secretOutput[:len(secretOutput)-1]
					output += fmt.Sprintf(" · %s\n   Environment Variables:\n%s\n", linkedService.SecretName, secretOutput)
				}

			} else {
				output += fmt.Sprintf(" · %s\n", linkedService.SecretName)
			}

		}

		if len(output) > 0 {
			// Cut off the last newline and output
			output = output[:len(output)-1]
			log.Describef("Linked Services:\n", output)

		}

	}
	return nil
}

// GetFullName generates a command's full name based on its parent's full name and its own name
func GetFullName(parentName, name string) string {
	return parentName + " " + name
}

// VisitCommands visits each command within Cobra.
// Adapted from: https://github.com/cppforlife/knctl/blob/612840d3c9729b1c57b20ca0450acab0d6eceeeb/pkg/knctl/cobrautil/misc.go#L23
func VisitCommands(cmd *cobra.Command, f func(*cobra.Command)) {
	f(cmd)
	for _, child := range cmd.Commands() {
		VisitCommands(child, f)
	}
}

// ModifyAdditionalFlags modifies the flags and updates the descriptions
// as well as changes whether or not machine readable output
// has been passed in..
//
// Return the flag usages for the help outout
func ModifyAdditionalFlags(cmd *cobra.Command) string {

	// Hide the machine readable output if the command
	// does not have the annotation.
	machineOutput := cmd.Annotations["machineoutput"]
	f := cmd.InheritedFlags()

	f.VisitAll(func(f *pflag.Flag) {
		// Remove json flag if machineoutput has not been passed in
		if f.Name == "o" && machineOutput == "json" {
			f.Hidden = false
		}
	})

	return CapitalizeFlagDescriptions(f)
}

// CapitalizeFlagDescriptions adds capitalizations
func CapitalizeFlagDescriptions(f *pflag.FlagSet) string {
	f.VisitAll(func(f *pflag.Flag) {
		cap := []rune(f.Usage)
		cap[0] = unicode.ToUpper(cap[0])
		f.Usage = string(cap)
	})
	return f.FlagUsages()
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

// ThrowContextError prints a context error if application/project is not found
func ThrowContextError() error {
	return errors.Errorf(`Please specify the application name and project name
Or use the command from inside a directory containing an odo component.`)
}
