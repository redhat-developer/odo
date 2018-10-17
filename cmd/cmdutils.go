package cmd

import (
	"fmt"
	"os"
	"strings"

	"text/tabwriter"

	"github.com/pkg/errors"
	"github.com/redhat-developer/odo/pkg/application"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/occlient"
	"github.com/redhat-developer/odo/pkg/project"
	"github.com/redhat-developer/odo/pkg/storage"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/validation"
)

// printDeleteAppInfo will print things which will be deleted
func printDeleteAppInfo(client *occlient.Client, appName string) error {
	componentList, err := component.List(client, appName)
	if err != nil {
		return errors.Wrap(err, "failed to get Component list")
	}

	for _, currentComponent := range componentList {
		_, _, componentURL, appStore, err := component.GetComponentDesc(client, currentComponent.Name, appName)
		if err != nil {
			return errors.Wrap(err, "unable to get component description")
		}
		fmt.Println("Component", currentComponent.Name, "will be deleted.")

		if len(componentURL) != 0 {
			fmt.Println("  Externally exposed URL will be removed")
		}

		for _, store := range appStore {
			fmt.Println("  Storage", store.Name, "of size", store.Size, "will be removed")
		}

	}
	return nil
}

// getComponent returns the component to be used for the operation. If an input
// component is specified, then it is returned if it exists, if not,
// the current component is fetched and returned. If no component set, throws error
func getComponent(client *occlient.Client, inputComponent string, applicationName string) string {
	if len(inputComponent) == 0 {
		c, err := component.GetCurrent(applicationName, client.Namespace)
		checkError(err, "Could not get current component")
		if c == "" {
			fmt.Println("There is no component set")
			os.Exit(1)
		}
		return c
	}
	exists, err := component.Exists(client, inputComponent, applicationName)
	checkError(err, "")
	if !exists {
		fmt.Printf("Component %v does not exist\n", inputComponent)
		os.Exit(1)
	}
	return inputComponent
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

// validateName will do validation of application & component names
// Criteria for valid name in kubernetes: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/
func validateName(name string) error {

	errorList := validation.IsDNS1123Label(name)

	if len(errorList) != 0 {
		return errors.New(fmt.Sprintf("%s is not a valid name:  %s", name, strings.Join(errorList, " ")))
	}

	return nil

}

// validateStoragePath will validate storagePath, if there is any existing storage with similar path, it will give an error
func validateStoragePath(client *occlient.Client, storagePath, componentName, applicationName string) error {
	storeList, err := storage.List(client, componentName, applicationName)
	if err != nil {
		return err
	}
	for _, store := range storeList {
		if store.Path == storagePath {
			return errors.Errorf("there already is a storage mounted at %s", storagePath)
		}
	}
	return nil
}

// printMountedStorageInComponent prints all the mounted storage in a given component of the application
func printMountedStorageInComponent(client *occlient.Client, componentName string, applicationName string) {

	// defining the column structure of the table
	tabWriterMounted := tabwriter.NewWriter(os.Stdout, 5, 2, 3, ' ', tabwriter.TabIndent)

	// create headers of mounted storage table
	fmt.Fprintln(tabWriterMounted, "NAME", "\t", "SIZE", "\t", "PATH")

	storageListMounted, err := storage.ListMounted(client, componentName, applicationName)
	checkError(err, "could not get mounted storage list")

	// iterating over all mounted storage and put in the mount storage table
	if len(storageListMounted) > 0 {
		for _, mStorage := range storageListMounted {
			fmt.Fprintln(tabWriterMounted, mStorage.Name, "\t", mStorage.Size, "\t", mStorage.Path)
		}

		// print all mounted storage of the given component
		fmt.Printf("The component '%v' has the following storage attached:\n", componentName)
		tabWriterMounted.Flush()
	} else {
		fmt.Printf("The component '%v' has no storage attached\n", componentName)
	}
	fmt.Println("")
}

// printMountedStorageInAllComponent prints all the mounted storage in all the components of the application and project
func printMountedStorageInAllComponent(client *occlient.Client, applicationName string) {
	componentList, err := component.List(client, applicationName)
	checkError(err, "could not get component list")

	// iterating over all the components in the given aplication and project
	for _, component := range componentList {
		printMountedStorageInComponent(client, component.Name, applicationName)
	}
}

// printUnmountedStorage prints all the unmounted storage in the application
func printUnmountedStorage(client *occlient.Client, applicationName string) {

	// defining the column structure of the unmounted storage table
	tabWriterUnmounted := tabwriter.NewWriter(os.Stdout, 5, 2, 3, ' ', tabwriter.TabIndent)

	// create header of unmounted storage in all the components of the given application and project
	fmt.Fprintln(tabWriterUnmounted, "NAME", "\t", "SIZE")

	storageListUnmounted, err := storage.ListUnmounted(client, applicationName)
	checkError(err, "could not get unmounted storage list")

	// iterating over all unmounted storage and put in the unmount storage table
	if len(storageListUnmounted) > 0 {
		for _, uStorage := range storageListUnmounted {
			fmt.Fprintln(tabWriterUnmounted, uStorage.Name, "\t", uStorage.Size)
		}

		// print unmounted storage of all the application
		fmt.Printf("Storage that are not mounted to any component:\n")
		tabWriterUnmounted.Flush()
	}
	fmt.Println("")
}

// GetAppName returns application name from the provided flag or if flag is not provided, it will return current application name
func getAppName(client *occlient.Client) string {
	// applicationFlag is `--application` flag
	if applicationFlag != "" {
		_, err := application.Exists(client, applicationFlag)
		checkError(err, "")
		return applicationFlag
	}
	applicationName, err := application.GetCurrent(client.Namespace)
	checkError(err, "unable to get current application")

	return applicationName
}

// getAndSetNamespace checks whether project flag is provided,
// if provided, it validates the name and sets it as namespace for further operations
// if not provided, it fetches current namespace and sets it as namespace for further operations
// getAndSetNamespace also return the project name
func getAndSetNamespace(client *occlient.Client) string {
	// projectFlag is `--project` flag
	if projectFlag != "" {
		_, err := project.Exists(client, projectFlag)
		checkError(err, "")
		client.Namespace = projectFlag
		return projectFlag
	}
	client.Namespace = project.GetCurrent(client)
	return project.GetCurrent(client)

}

func addProjectFlag(cmd *cobra.Command) {
	cmd.Flags().StringVar(&projectFlag, "project", "", "Project, defaults to active project")
}

func addComponentFlag(cmd *cobra.Command) {
	cmd.Flags().StringVar(&componentFlag, "component", "", "Component, defaults to active component.")
}

func addApplicationFlag(cmd *cobra.Command) {
	cmd.Flags().StringVar(&applicationFlag, "app", "", "Application, defaults to active application")
}
