package cmd

import (
	"fmt"
	"strings"

	"os"
	"text/tabwriter"

	"github.com/redhat-developer/odo/pkg/application"
	"github.com/redhat-developer/odo/pkg/project"
	"github.com/redhat-developer/odo/pkg/storage"
	"github.com/spf13/cobra"
)

var (
	storageComponent       string
	storageSize            string
	storagePath            string
	storageForceDeleteflag bool
)

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
		componentName := getComponent(client, storageComponent, applicationName, projectName)

		_, err = storage.Create(client, args[0], storageSize, storagePath, componentName, applicationName)
		checkError(err, "")
		fmt.Printf("Added storage %v to %v\n", args[0], componentName)
	},
}

var storageUnmountCmd = &cobra.Command{
	Use:   "unmount [storage name]",
	Short: "unmount storage from the current component",
	Long: `unmount storage from the current component.
  The storage and the contents are not deleted, the storage is only unmounted
  from the component, and hence is no longer accessible by the component.`,
	Example: `  # Unmount storage 'dbstorage' from current component
  odo storage umount dbstorage

  # Unmount storage 'database' from component 'mongodb'
  odo storage umount database --component mongodb`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := getOcClient()
		applicationName, err := application.GetCurrent(client)
		checkError(err, "")
		projectName := project.GetCurrent(client)
		componentName := getComponent(client, storageComponent, applicationName, projectName)

		storageName := args[0]
		exists, err := storage.IsMounted(client, storageName, componentName, applicationName)
		checkError(err, "")
		if !exists {
			fmt.Printf("Storage %v does not exist in component %v\n", storageName, componentName)
			os.Exit(1)
		}

		err = storage.Unmount(client, storageName, componentName, applicationName)
		checkError(err, "Unable to unmount storage %v from component %v", storageName, componentName)

		fmt.Printf("Unmounted storage %v from %v\n", storageName, componentName)
	},
}

var storageDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete storage from component",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := getOcClient()

		storageName := args[0]
		applicationName, err := application.GetCurrent(client)
		checkError(err, "")
		exists, err := storage.Exists(client, storageName, applicationName)
		checkError(err, "")
		if !exists {

			fmt.Printf("The storage %v does not exists in the application %v\n", storageName, applicationName)
			os.Exit(1)
		}

		componentName, err := storage.GetComponentNameFromStorageName(client, storageName)
		if err != nil {
			checkError(err, "Unable to get component associated with %s storage.", storageName)
		}

		var confirmDeletion string
		if storageForceDeleteflag {
			confirmDeletion = "y"
		} else {
			if componentName != "" {
				mPath := storage.GetMountPath(client, storageName, componentName, applicationName)
				fmt.Printf("Are you sure you want to delete the storage %v mounted to %v in %v component? [y/N] ", storageName, mPath, componentName)
			} else {
				fmt.Printf("Are you sure you want to delete the storage %v that is not currently mounted to any component? [y/N] ", storageName)
			}
			fmt.Scanln(&confirmDeletion)
		}
		if strings.ToLower(confirmDeletion) == "y" {
			componentName, err = storage.Delete(client, storageName, applicationName)
			checkError(err, "failed to delete storage")
			if componentName != "" {
				fmt.Printf("Deleted storage %v from %v\n", storageName, componentName)
			} else {
				fmt.Printf("Deleted storage %v\n", storageName)
			}
		} else {
			fmt.Printf("Aborting deletion of storage: %v\n", storageName)
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
		componentName := getComponent(client, storageComponent, applicationName, projectName)

		storageList, err := storage.List(client, componentName, applicationName)
		checkError(err, "Failed to list storage")

		hasMounted := false
		hasUnmounted := false
		tabWriterMounted := tabwriter.NewWriter(os.Stdout, 5, 2, 3, ' ', tabwriter.TabIndent)
		tabWriterUnmounted := tabwriter.NewWriter(os.Stdout, 5, 2, 3, ' ', tabwriter.TabIndent)

		//create headers
		fmt.Fprintln(tabWriterMounted, "NAME", "\t", "SIZE", "\t", "PATH")
		fmt.Fprintln(tabWriterUnmounted, "NAME", "\t", "SIZE")

		for _, storage := range storageList {
			if storage.Path != "" {
				if !hasMounted {
					hasMounted = true
				}
				fmt.Fprintln(tabWriterMounted, storage.Name, "\t", storage.Size, "\t", storage.Path)
			} else {
				if !hasUnmounted {
					hasUnmounted = true
				}
				fmt.Fprintln(tabWriterUnmounted, storage.Name, "\t", storage.Size)
			}
		}
		if hasMounted {
			fmt.Printf("The component '%v' has the following storage attached -\n", componentName)
			tabWriterMounted.Flush()
		} else {
			fmt.Printf("The component '%v' has no storage attached\n", componentName)
		}
		fmt.Println("")
		if hasUnmounted {
			fmt.Printf("The following unmounted storages can be mounted to '%v' - \n", componentName)
			tabWriterUnmounted.Flush()
		} else {
			fmt.Printf("No unmounted storage exists to mount to '%v' \n", componentName)
		}
	},
}

var storageMountCmd = &cobra.Command{
	Use:   "mount [storage name]",
	Short: "mount storage to a component",
	Example: `  # Mount storage 'dbstorage' to current component
  odo storage mount dbstorage

  # Mount storage 'database' to component 'mongodb'
  odo storage mount database --component mongodb`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := getOcClient()

		storageName := args[0]
		applicationName, err := application.GetCurrent(client)
		checkError(err, "")
		projectName := project.GetCurrent(client)
		componentName := getComponent(client, storageComponent, applicationName, projectName)

		exists, err := storage.Exists(client, storageName, applicationName)
		checkError(err, "unable to check if the storage exists in the current application")
		if !exists {
			fmt.Printf("The storage %v does not exists in the current application '%v'", storageName, applicationName)
			os.Exit(1)
		}
		isMounted, err := storage.IsMounted(client, storageName, componentName, applicationName)
		checkError(err, "unable to check if the component is already mounted or not")
		if isMounted {
			fmt.Printf("The storage %v is already mounted on the current component '%v'\n", storageName, componentName)
			os.Exit(1)
		}
		err = storage.Mount(client, storagePath, storageName, componentName, applicationName)
		checkError(err, "")
		fmt.Printf("The storage %v is successfully mounted to the current component '%v'\n", storageName, componentName)
	},
}

func init() {
	storageDeleteCmd.Flags().BoolVarP(&storageForceDeleteflag, "force", "f", false, "Delete storage without prompting")
	storageCreateCmd.Flags().StringVar(&storageSize, "size", "", "Size of storage to add")
	storageCreateCmd.MarkFlagRequired("size")
	storageCreateCmd.Flags().StringVar(&storagePath, "path", "", "Path to mount the storage on")
	storageCreateCmd.MarkFlagRequired("path")
	storageMountCmd.Flags().StringVar(&storagePath, "path", "", "Path to mount the storage on")
	storageMountCmd.MarkFlagRequired("path")

	storageCreateCmd.Flags().StringVar(&storageComponent, "component", "", "Component to add storage to. Defaults to active component.")
	storageUnmountCmd.Flags().StringVar(&storageComponent, "component", "", "Component from which the storage will be unmounted. Defaults to active component.")
	storageListCmd.Flags().StringVar(&storageComponent, "component", "", "List storage for given component. Defaults to active component.")

	storageCmd.AddCommand(storageCreateCmd)
	storageCmd.AddCommand(storageDeleteCmd)
	storageCmd.AddCommand(storageUnmountCmd)
	storageCmd.AddCommand(storageListCmd)
	storageCmd.AddCommand(storageMountCmd)

	rootCmd.AddCommand(storageCmd)
}
