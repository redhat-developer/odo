package cmd

import (
	"fmt"
	"os"

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

func getComponent(client *occlient.Client) string {
	if len(storageComponent) == 0 {
		c, err := component.GetCurrent(client)
		if err != nil {
			fmt.Printf("Could not get current component: %v\n", err)
			os.Exit(1)
		}
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
	Short: "Create storage and mount to component",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := getOcClient()
		cmpnt := getComponent(client)
		_, err := storage.Add(client,
			&occlient.VolumeConfig{
				Name:             &args[0],
				DeploymentConfig: &cmpnt,
				Path:             &storagePath,
				Size:             &storageSize,
			})
		if err != nil {
			fmt.Printf("Failed to add storage: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Added storage %v to %v\n", args[0], cmpnt)
	},
}

var storageRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove storage from component",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := getOcClient()
		cmpnt := getComponent(client)
		var storageName *string
		if len(args) == 0 {
			storageName = nil
		} else {
			storageName = &args[0]
		}
		_, err := storage.Remove(client,
			&occlient.VolumeConfig{
				Name:             storageName,
				DeploymentConfig: &cmpnt,
			})
		if err != nil {
			fmt.Printf("Failed to remove storage: %v\n", err)
			os.Exit(1)
		}
		if len(args) == 0 {
			fmt.Printf("Removed all storage from %v\n", cmpnt)
		} else {
			fmt.Printf("Removed %v from %v\n", *storageName, cmpnt)
		}
	},
}

var storageListCmd = &cobra.Command{
	Use:   "list",
	Short: "List storage attached to a component",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := getOcClient()
		cmpnt := getComponent(client)
		output, err := storage.List(client,
			&occlient.VolumeConfig{
				DeploymentConfig: &cmpnt,
			})
		if err != nil {
			fmt.Printf("Failed to list storage: %v\n", err)
			os.Exit(1)
		}
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
