package cmd

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/occlient"
	"github.com/redhat-developer/odo/pkg/storage"
)

// printDeleteAppInfo will print things which will be deleted
func printDeleteAppInfo(client *occlient.Client, appName string, currentProject string) error {
	componentList, err := component.List(client, appName, currentProject)
	if err != nil {
		return errors.Wrap(err, "failed to get Component list")
	}

	for _, cmpnt := range componentList {
		_, _, componentURL, appStore, err := component.GetComponentDesc(client, cmpnt.Name, appName, currentProject)
		if err != nil {
			return errors.Wrap(err, "unable to get component description")
		}
		fmt.Println("Component", cmpnt.Name, "will be deleted.")

		if len(componentURL) != 0 {
			fmt.Println("  This component is externally exposed, and the URL will be removed")
		}

		for _, store := range appStore {
			fmt.Println("  This component uses storage", store.Name, "of size", store.Size, "will be removed")
		}

	}
	return nil
}

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

// printComponentInfo prints Component Information like path, URL & storage
func printComponentInfo(cmpntName string, componentType string, path string, componentURL string, appStore []storage.StorageInfo) {
	// Source
	if path != "" {
		fmt.Println("Component", cmpntName, "of type", componentType, "with source in", path)
	}
	// URL
	if componentURL != "" {
		fmt.Println("This Component is externally exposed via", componentURL)
	}
	// Storage
	for _, store := range appStore {
		fmt.Println("This Component uses storage", store.Name, "of size", store.Size)
	}
}
