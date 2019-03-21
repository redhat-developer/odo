package util

import (
	"fmt"
	"os"
	"strings"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	"github.com/openshift/odo/pkg/component"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/occlient"
	storagePkg "github.com/openshift/odo/pkg/storage"
	urlPkg "github.com/openshift/odo/pkg/url"
)

// LogErrorAndExit prints the cause of the given error and exits the code with an
// exit code of 1.
// If the context is provided, then that is printed, if not, then the cause is
// detected using errors.Cause(err)
func LogErrorAndExit(err error, context string, a ...interface{}) {
	if err != nil {
		glog.V(4).Infof("Error:\n%v", err)
		if context == "" {
			log.Error(errors.Cause(err))
		} else {
			log.Errorf(fmt.Sprintf("%s", context), a...)
		}
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
{{.LocalFlags.FlagUsages | trimRightSpace}}{{end}}{{ if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimRightSpace}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsHelpCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{ if .HasAvailableSubCommands }}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`

// PrintComponentInfo prints Component Information like path, URL & storage
func PrintComponentInfo(client *occlient.Client, currentComponentName string, componentDesc component.Component, applicationName string) {
	fmt.Printf("Component Name: %v\nType: %v\n", currentComponentName, componentDesc.Spec.Type)
	// Source
	if componentDesc.Spec.Source != "" {
		fmt.Printf("Source: %v\n", componentDesc.Spec.Source)
	}

	// Env
	if componentDesc.Spec.Env != nil {
		fmt.Println("\nEnvironment variables:")
		for _, env := range componentDesc.Spec.Env {
			fmt.Printf(" - %v=%v\n", env.Name, env.Value)
		}
	}
	// Storage
	if len(componentDesc.Spec.Storage) > 0 {
		fmt.Println("\nStorage:")
		storages, err := storagePkg.List(client, currentComponentName, applicationName)
		LogErrorAndExit(err, "")
		for _, storage := range componentDesc.Spec.Storage {
			store := storages.Get(storage)
			fmt.Printf(" - %v of size %v mounted to %v\n", store.Name, store.Spec.Size, store.Status.Path)
		}
	}
	// URL
	if componentDesc.Spec.URL != nil {
		fmt.Println("\nURLs")
		urls, err := urlPkg.List(client, currentComponentName, applicationName)
		LogErrorAndExit(err, "")
		for _, componentURL := range componentDesc.Spec.URL {
			url := urls.Get(componentURL)
			fmt.Printf(" - %v exposed via %v\n", urlPkg.GetURLString(url.Spec.Protocol, url.Spec.Host), url.Spec.Port)
		}

	}
	// Linked services
	if len(componentDesc.Status.LinkedServices) > 0 {
		fmt.Print("Linked Services: ")
		fmt.Printf("%v\n", strings.Join(componentDesc.Status.LinkedServices, ","))
	}
	// Linked components
	if len(componentDesc.Status.LinkedComponents) > 0 {
		fmt.Println("Linked Components")
		for name, ports := range componentDesc.Status.LinkedComponents {
			if len(ports) > 0 {
				fmt.Printf("Name: %v - Port(s): %v\n", name, strings.Join(ports, ","))
			}
		}
	}
}

// GetFullName generates a command's full name based on its parent's full name and its own name
func GetFullName(parentName, name string) string {
	return parentName + " " + name
}
