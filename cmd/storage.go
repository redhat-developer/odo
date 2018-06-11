package cmd

import (
	"fmt"
	"os"
	"strings"
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
	Example: fmt.Sprintf("%s\n%s\n%s\n%s",
		storageCreateCmd.Example,
		storageDeleteCmd.Example,
		storageUnmountCmd.Example,
		storageListCmd.Example),
}

var storageCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create storage and mount to a component",
	Long:  "Create storage and mount to a component",
	Example: `  # Create storage of size 1Gb to a component
  odo storage create mystorage --path=/opt/app-root/src/storage/ --size=1Gi
	`,
	Args: cobra.ExactArgs(1),
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
	Use:   "unmount PATH | STORAGE_NAME",
	Short: "Unmount storage from the given path or identified by its name, from the current component",
	Long: `Unmount storage from the given path or identified by its name, from the current component.

  The storage and the contents are not deleted, the storage or storage with the given path, is only unmounted from the component, and hence is no longer accessible by the component.`,
	Example: `  # Unmount storage 'dbstorage' from current component
  odo storage unmount dbstorage

  # Unmount storage 'database' from component 'mongodb'
  odo storage unmount database --component mongodb

  # Unmount storage mounted to path '/data' from current component
  odo storage unmount /data

  # Unmount storage mounted to path '/data' from component 'mongodb'
  odo storage unmount /data --component mongodb
`,
	Aliases: []string{"umount"},
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := getOcClient()
		applicationName, err := application.GetCurrent(client)
		checkError(err, "")
		projectName := project.GetCurrent(client)
		componentName := getComponent(client, storageComponent, applicationName, projectName)

		var storageName string
		// checking if the first character in the argument is a "/", indicating a path or not, indicating a storage name
		if string(args[0][0]) == "/" {
			path := args[0]
			storageName, err = storage.GetStorageNameFromMountPath(client, path, componentName, applicationName)
			checkError(err, "Unable to get storage name from mount path")
			if storageName == "" {
				fmt.Printf("No storage is mounted to %s in the component %s\n", path, componentName)
				os.Exit(1)
			}
		} else {
			storageName = args[0]
			exists, err := storage.IsMounted(client, storageName, componentName, applicationName)
			checkError(err, "Unable to check if storage is mounted or not")
			if !exists {
				fmt.Printf("Storage %v does not exist in component %v\n", storageName, componentName)
				os.Exit(1)
			}
		}

		err = storage.Unmount(client, storageName, componentName, applicationName)
		checkError(err, "Unable to unmount storage %v from component %v", storageName, componentName)

		fmt.Printf("Unmounted storage %v from %v\n", storageName, componentName)
	},
}

var storageDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete storage from component",
	Example: `  # Delete storage mystorage from the currently active component
  odo storage delete mystorage

  # Delete storage mystorage from component 'mongodb'
  odo storage delete mystorage --component mongodb
`,
	Args: cobra.ExactArgs(1),
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
	Short: "List storage attached to a component",
	Long:  "List storage attached to a component",
	Example: `  # List all storage attached or mounted to the current component and 
  # all unattached or unmounted storage in the current application
  odo storage list
	`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := getOcClient()
		applicationName, err := application.GetCurrent(client)
		checkError(err, "")
		projectName := project.GetCurrent(client)
		componentName := getComponent(client, storageComponent, applicationName, projectName)

		if componentName == "" {
			fmt.Println("No component selected")
			os.Exit(1)
		}

		storageList, err := storage.List(client, componentName, applicationName)
		checkError(err, "")

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
  odo storage mount dbstorage --path /data

  # Mount storage 'database' to component 'mongodb'
  odo storage mount database --component mongodb --path /data`,
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
	storageMountCmd.Flags().StringVar(&storageComponent, "component", "", "Component to which storage will be mounted to.")

	storageCmd.AddCommand(storageCreateCmd)
	storageCmd.AddCommand(storageDeleteCmd)
	storageCmd.AddCommand(storageUnmountCmd)
	storageCmd.AddCommand(storageListCmd)
	storageCmd.AddCommand(storageMountCmd)

	// Add a defined annotation in order to appear in the help menu
	storageCmd.Annotations = map[string]string{"command": "other"}
	storageCmd.SetUsageTemplate(cmdUsageTemplate)

	rootCmd.AddCommand(storageCmd)
}
