package cmd

import (
	"fmt"
	"os"

	"github.com/redhat-developer/ocdev/pkg/application"
	"github.com/redhat-developer/ocdev/pkg/component"
	"github.com/redhat-developer/ocdev/pkg/occlient"
	"github.com/redhat-developer/ocdev/pkg/storage"
	"github.com/spf13/cobra"
)

var (
	storageComponent string
	storageSize      string
	storagePath      string
)

// Gets the current application

func getApplication(client *occlient.Client) string {

	currentApplication, err := application.GetCurrentOrDefault(client)

	if err != nil {
		fmt.Printf("Could not get application: %v\n", err)
		os.Exit(1)
	}

	return currentApplication
}

// Gets the current component
func getComponent(client *occlient.Client) string {

	if len(storageComponent) == 0 {
		c, err := component.GetCurrent(client)
		checkError(err, "Could not get current component")
		return c
	}

	return storageComponent
}

var storageCmd = &cobra.Command{
	Use:   "storage",
	Short: "Perform storage operations",
	Long:  "Perform storage operations",
}

var storageAddCmd = &cobra.Command{
	Use:   "add",
	Short: "create storage and mount to component",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		client := getOcClient()

		currentComponent := getApplication(client) + "-" + getComponent(client)

		// Get the current application (namespaced apps)

		_, err := storage.Add(client,
			&occlient.VolumeConfig{
				Name:             &args[0],
				DeploymentConfig: &currentComponent,
				Path:             &storagePath,
				Size:             &storageSize,
			})
		checkError(err, "Failed to add storage")
		fmt.Printf("Added storage %v to %v\n", args[0], currentComponent)
	},
}

var storageRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "remove storage from component",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := getOcClient()
		currentComponent := getApplication(client) + "-" + getComponent(client)
		var storageName *string
		if len(args) == 0 {
			storageName = nil
		} else {
			storageName = &args[0]
		}
		_, err := storage.Remove(client,
			&occlient.VolumeConfig{
				Name:             storageName,
				DeploymentConfig: &currentComponent,
			})
		checkError(err, "Failed to remove storage")

		if len(args) == 0 {
			fmt.Printf("Removed all storage from %v\n", currentComponent)
		} else {
			fmt.Printf("Removed %v from %v\n", *storageName, currentComponent)
		}
	},
}

var storageListCmd = &cobra.Command{
	Use:   "list",
	Short: "list storage attached to a component",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := getOcClient()
		currentComponent := getApplication(client) + "-" + getComponent(client)

		output, err := storage.List(client,
			&occlient.VolumeConfig{
				DeploymentConfig: &currentComponent,
			})
		checkError(err, "Failed to list storage")
		fmt.Println(output)
	},
}

func init() {
	storageAddCmd.Flags().StringVar(&storageSize, "size", "", "Size of storage to add")
	storageAddCmd.MarkFlagRequired("size")
	storageAddCmd.Flags().StringVar(&storagePath, "path", "", "Path to mount the storage on")
	storageAddCmd.MarkFlagRequired("path")

	storageCmd.AddCommand(storageAddCmd)
	storageCmd.AddCommand(storageRemoveCmd)
	storageCmd.AddCommand(storageListCmd)

	storageCmd.PersistentFlags().StringVar(&storageComponent, "component", "", "Component to add storage to, defaults to active component")
	rootCmd.AddCommand(storageCmd)
}
