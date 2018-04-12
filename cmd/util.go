package cmd

import (
	"fmt"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/occlient"
	"os"
)

// getComponent returns the component to be used for the operation. If an input
// component is specified, then it is returned, if not, the current component
// is fetched and returned
func getComponent(client *occlient.Client, inputComponent, applicationName, projectName string) string {
	if len(inputComponent) == 0 {
		c, err := component.GetCurrent(client, applicationName, projectName)
		checkError(err, "Could not get current component")
		return c
	}
	exists, err := component.Exists(client, inputComponent, applicationName, projectName)
	checkError(err, "")
	if !exists {
		fmt.Printf("Component %v does not exist", inputComponent)
		os.Exit(1)
	}
	return inputComponent
}
