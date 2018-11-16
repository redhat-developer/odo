package cli

import (
	"fmt"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/storage"
	"github.com/spf13/cobra"
)

var describeCmd = &cobra.Command{
	Use:   "describe [component_name]",
	Short: "Describe the given component",
	Long:  `Describe the given component.`,
	Example: `  # Describe nodejs component,
  odo describe nodejs
	`,
	Args: cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		context := genericclioptions.NewContext(cmd)
		client := context.Client
		applicationName := context.Application

		// If no arguments have been passed, get the current component
		// else, use the first argument and check to see if it exists
		var componentName string
		if len(args) == 0 {
			componentName = context.Component()
		} else {
			componentName = context.Component(args[0])
		}
		componentType, path, componentURL, appStore, err := component.GetComponentDesc(client, componentName, applicationName)
		util.CheckError(err, "")
		printComponentInfo(componentName, componentType, path, componentURL, appStore)
	},
}

func init() {
	// Add a defined annotation in order to appear in the help menu
	describeCmd.Annotations = map[string]string{"command": "component"}
	describeCmd.SetUsageTemplate(CmdUsageTemplate)

	//Adding `--project` flag
	AddProjectFlag(describeCmd)
	//Adding `--application` flag
	AddApplicationFlag(describeCmd)

	RootCmd().AddCommand(describeCmd)
}

// printComponentInfo prints Component Information like path, URL & storage
func printComponentInfo(currentComponentName string, componentType string, path string, componentURL string, appStore []storage.StorageInfo) {
	// Source
	if path != "" {
		fmt.Println("Component", currentComponentName, "of type", componentType, "with source in", path)
	}
	// URL
	if componentURL != "" {
		fmt.Println("Externally exposed via", componentURL)
	}
	// Storage
	for _, store := range appStore {
		fmt.Println("Storage", store.Name, "of size", store.Size)
	}
}
