package cmd

import (
	"fmt"
	"github.com/redhat-developer/ocdev/pkg/component"
	"github.com/redhat-developer/ocdev/pkg/occlient"
	"github.com/redhat-developer/ocdev/pkg/storage"
	"github.com/spf13/cobra"
	"os"
)

var (
	storageComponent string
	storageSize      string
	storagePath      string
)

func getComponent() string {
	if len(storageComponent) == 0 {
		c, err := component.GetCurrent()
		if err != nil {
			fmt.Printf("Could not get current component: %v\n", err)
			os.Exit(-1)
		}
		return c
	}
	return storageComponent
}

var storageCmd = &cobra.Command{
	Use:   "storage",
	Short: "storage",
	Long:  "perform storage operations",
}

var storageAddCmd = &cobra.Command{
	Use:   "add",
	Short: "create storage and mount to component",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cmpnt := getComponent()
		_, err := storage.Add(&occlient.VolumeConfig{
			Name:             &args[0],
			DeploymentConfig: &cmpnt,
			Path:             &storagePath,
			Size:             &storageSize,
		})
		if err != nil {
			fmt.Printf("Failed to add storage: %v\n", err)
			os.Exit(-1)
		}
		fmt.Printf("Added storage %v to %v\n", args[0], cmpnt)
	},
}

var storageRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "remove storage from component",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cmpnt := getComponent()
		var storageName *string
		if len(args) == 0 {
			storageName = nil
		} else {
			storageName = &args[0]
		}
		_, err := storage.Remove(&occlient.VolumeConfig{
			Name:             storageName,
			DeploymentConfig: &cmpnt,
		})
		if err != nil {
			fmt.Printf("Failed to remove storage: %v\n", err)
			os.Exit(-1)
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
	Short: "list storage attached to a component",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cmpnt := getComponent()
		output, err := storage.List(&occlient.VolumeConfig{
			DeploymentConfig: &cmpnt,
		})
		if err != nil {
			fmt.Printf("Failed to list storage: %v\n", err)
			os.Exit(-1)
		}
		fmt.Println(output)
	},
}

func init() {
	storageAddCmd.Flags().StringVar(&storageSize, "size", "", "size of storage to add")
	storageAddCmd.MarkFlagRequired("size")
	storageAddCmd.Flags().StringVar(&storagePath, "path", "", "path to mount the storage on")
	storageAddCmd.MarkFlagRequired("path")

	storageCmd.AddCommand(storageAddCmd)
	storageCmd.AddCommand(storageRemoveCmd)
	storageCmd.AddCommand(storageListCmd)

	storageCmd.PersistentFlags().StringVar(&storageComponent, "component", "", "component to add storage to, defaults to active component")
	rootCmd.AddCommand(storageCmd)
}
