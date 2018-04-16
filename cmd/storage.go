package cmd

import (
	"fmt"

	"github.com/redhat-developer/odo/pkg/application"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/occlient"
	"github.com/redhat-developer/odo/pkg/project"
	"github.com/redhat-developer/odo/pkg/storage"
	"github.com/spf13/cobra"
)

var (
	storageComponent string
	storageSize      string
	storagePath      string
)

func getComponent(client *occlient.Client, applicationName, projectName string) string {
	if len(storageComponent) == 0 {
		c, err := component.GetCurrent(client, applicationName, projectName)
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

var storageCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "create storage and mount to component",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := getOcClient()
		applicationName, err := application.GetCurrent(client)
		checkError(err, "")
		projectName := project.GetCurrent(client)
		cmpnt := getComponent(client, applicationName, projectName)
		_, err = storage.Create(client, args[0], storageSize, storagePath, cmpnt, applicationName)
		checkError(err, "")
		fmt.Printf("Added storage %v to %v\n", args[0], cmpnt)
	},
}

var storageRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "remove storage from component",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := getOcClient()

		var storageName string
		if len(args) != 0 {
			storageName = args[0]
		}
		applicationName, err := application.GetCurrent(client)
		checkError(err, "")
		projectName := project.GetCurrent(client)

		cmpnt := getComponent(client, applicationName, projectName)

		err = storage.Remove(client, storageName, applicationName, cmpnt)
		checkError(err, "failed to remove storage")

		switch storageName {
		case "":
			fmt.Printf("Removed all storage from %v\n", cmpnt)
		default:
			fmt.Printf("Removed %v from %v\n", storageName, cmpnt)
		}
	},
}

var storageListCmd = &cobra.Command{
	Use:   "list",
	Short: "list storage attached to a component",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := getOcClient()
		applicationName, err := application.GetCurrent(client)
		checkError(err, "")
		projectName := project.GetCurrent(client)

		cmpnt := getComponent(client, applicationName, projectName)
		storageList, err := storage.List(client, applicationName, cmpnt)
		checkError(err, "failed to list storage")

		if len(storageList) == 0 {
			fmt.Printf("The component '%v' has no storage attached\n", cmpnt)
		} else {
			fmt.Printf("The component '%v' has the following storage attached -\n", cmpnt)
			for _, strg := range storageList {
				fmt.Printf("- %v - %v\n", strg.Name, strg.Size)
			}
		}
	},
}

func init() {
	storageCreateCmd.Flags().StringVar(&storageSize, "size", "", "Size of storage to add")
	storageCreateCmd.MarkFlagRequired("size")
	storageCreateCmd.Flags().StringVar(&storagePath, "path", "", "Path to mount the storage on")
	storageCreateCmd.MarkFlagRequired("path")

	storageCmd.AddCommand(storageCreateCmd)
	storageCmd.AddCommand(storageRemoveCmd)
	storageCmd.AddCommand(storageListCmd)

	storageCmd.PersistentFlags().StringVar(&storageComponent, "component", "", "Component to add storage to, defaults to active component")
	rootCmd.AddCommand(storageCmd)
}
