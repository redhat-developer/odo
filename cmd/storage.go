package cmd

import (
	"fmt"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"os"
	"strings"

	"github.com/redhat-developer/odo/pkg/storage"
	"github.com/redhat-developer/odo/pkg/util"
	"github.com/spf13/cobra"
)

var (
	storageSize            string
	storagePath            string
	storageForceDeleteflag bool
	storageAllListflag     bool
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
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		context := genericclioptions.NewContext(cmd)
		client := context.Client
		applicationName := context.Application
		componentName := context.Component()

		var storageName string
		if len(args) != 0 {
			storageName = args[0]
		} else {
			storageName = componentName + "-" + util.GenerateRandomString(4)
		}
		// validate storage path
		err := validateStoragePath(client, storagePath, componentName, applicationName)
		odoutil.CheckError(err, "")

		_, err = storage.Create(client, storageName, storageSize, storagePath, componentName, applicationName)
		odoutil.CheckError(err, "")
		fmt.Printf("Added storage %v to %v\n", storageName, componentName)
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
		context := genericclioptions.NewContext(cmd)
		client := context.Client
		applicationName := context.Application
		componentName := context.Component()

		var storageName string
		var err error
		// checking if the first character in the argument is a "/", indicating a path or not, indicating a storage name
		if string(args[0][0]) == "/" {
			path := args[0]
			storageName, err = storage.GetStorageNameFromMountPath(client, path, componentName, applicationName)
			odoutil.CheckError(err, "Unable to get storage name from mount path")
			if storageName == "" {
				fmt.Printf("No storage is mounted to %s in the component %s\n", path, componentName)
				os.Exit(1)
			}
		} else {
			storageName = args[0]
			exists, err := storage.IsMounted(client, storageName, componentName, applicationName)
			odoutil.CheckError(err, "Unable to check if storage is mounted or not")
			if !exists {
				fmt.Printf("Storage %v does not exist in component %v\n", storageName, componentName)
				os.Exit(1)
			}
		}

		err = storage.Unmount(client, storageName, componentName, applicationName, true)
		odoutil.CheckError(err, "Unable to unmount storage %v from component %v", storageName, componentName)

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
		context := genericclioptions.NewContext(cmd)
		client := context.Client
		applicationName := context.Application

		storageName := args[0]

		exists, err := storage.Exists(client, storageName, applicationName)
		odoutil.CheckError(err, "")
		if !exists {

			fmt.Printf("The storage %v does not exists in the application %v\n", storageName, applicationName)
			os.Exit(1)
		}

		componentName, err := storage.GetComponentNameFromStorageName(client, storageName)
		if err != nil {
			odoutil.CheckError(err, "Unable to get component associated with %s storage.", storageName)
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
			odoutil.CheckError(err, "failed to delete storage")
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
		context := genericclioptions.NewContext(cmd)
		client := context.Client
		applicationName := context.Application

		if storageAllListflag {
			if cmd.Flag(odoutil.ComponentFlagName) != nil {
				fmt.Println("Invalid arguments. Component name is not needed")
				os.Exit(1)
			}
			printMountedStorageInAllComponent(client, applicationName)
		} else {
			// storageComponent is the input component name
			componentName := context.Component()
			printMountedStorageInComponent(client, componentName, applicationName)
		}
		printUnmountedStorage(client, applicationName)
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
		context := genericclioptions.NewContext(cmd)
		client := context.Client
		applicationName := context.Application
		componentName := context.Component()

		storageName := args[0]

		exists, err := storage.Exists(client, storageName, applicationName)
		odoutil.CheckError(err, "unable to check if the storage exists in the current application")
		if !exists {
			fmt.Printf("The storage %v does not exists in the current application '%v'", storageName, applicationName)
			os.Exit(1)
		}
		isMounted, err := storage.IsMounted(client, storageName, componentName, applicationName)
		odoutil.CheckError(err, "unable to check if the component is already mounted or not")
		if isMounted {
			fmt.Printf("The storage %v is already mounted on the current component '%v'\n", storageName, componentName)
			os.Exit(1)
		}
		err = storage.Mount(client, storagePath, storageName, componentName, applicationName)
		odoutil.CheckError(err, "")
		fmt.Printf("The storage %v is successfully mounted to the current component '%v'\n", storageName, componentName)
	},
}

func init() {
	storageCreateCmd.Flags().StringVar(&storageSize, "size", "", "Size of storage to add")
	storageCreateCmd.Flags().StringVar(&storagePath, "path", "", "Path to mount the storage on")
	storageCreateCmd.MarkFlagRequired("path")
	storageCreateCmd.MarkFlagRequired("size")

	storageDeleteCmd.Flags().BoolVarP(&storageForceDeleteflag, "force", "f", false, "Delete storage without prompting")

	storageListCmd.Flags().BoolVarP(&storageAllListflag, "all", "a", false, "List all storage in all components")

	storageMountCmd.Flags().StringVar(&storagePath, "path", "", "Path to mount the storage on")
	storageMountCmd.MarkFlagRequired("path")

	storageCmd.AddCommand(storageCreateCmd)
	storageCmd.AddCommand(storageDeleteCmd)
	storageCmd.AddCommand(storageUnmountCmd)
	storageCmd.AddCommand(storageListCmd)
	storageCmd.AddCommand(storageMountCmd)

	//Adding `--project` flag
	addProjectFlag(storageCreateCmd)
	addProjectFlag(storageDeleteCmd)
	addProjectFlag(storageListCmd)
	addProjectFlag(storageMountCmd)
	addProjectFlag(storageUnmountCmd)

	//Adding `--application` flag
	addApplicationFlag(storageCreateCmd)
	addApplicationFlag(storageDeleteCmd)
	addApplicationFlag(storageListCmd)
	addApplicationFlag(storageMountCmd)
	addApplicationFlag(storageUnmountCmd)

	//Adding `--component` flag
	addComponentFlag(storageCreateCmd)
	addComponentFlag(storageDeleteCmd)
	addComponentFlag(storageListCmd)
	addComponentFlag(storageMountCmd)
	addComponentFlag(storageUnmountCmd)

	// Add a defined annotation in order to appear in the help menu
	storageCmd.Annotations = map[string]string{"command": "other"}
	storageCmd.SetUsageTemplate(cmdUsageTemplate)

	rootCmd.AddCommand(storageCmd)
}
