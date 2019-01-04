package storage

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	appCmd "github.com/redhat-developer/odo/pkg/odo/cli/application"
	componentCmd "github.com/redhat-developer/odo/pkg/odo/cli/component"
	projectCmd "github.com/redhat-developer/odo/pkg/odo/cli/project"

	"github.com/pkg/errors"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/occlient"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/odo/util/completion"

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
		odoutil.LogErrorAndExit(err, "")

		_, err = storage.Create(client, storageName, storageSize, storagePath, componentName, applicationName)
		odoutil.LogErrorAndExit(err, "")
		log.Successf("Added storage %v to %v", storageName, componentName)
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
			odoutil.LogErrorAndExit(err, "Unable to get storage name from mount path")
			if storageName == "" {
				log.Errorf("No storage is mounted to %s in the component %s", path, componentName)
				os.Exit(1)
			}
		} else {
			storageName = args[0]
			exists, err := storage.IsMounted(client, storageName, componentName, applicationName)
			odoutil.LogErrorAndExit(err, "Unable to check if storage is mounted or not")
			if !exists {
				log.Errorf("Storage %v does not exist in component %v", storageName, componentName)
				os.Exit(1)
			}
		}

		err = storage.Unmount(client, storageName, componentName, applicationName, true)
		odoutil.LogErrorAndExit(err, "Unable to unmount storage %v from component %v", storageName, componentName)

		log.Infof("Unmounted storage %v from %v", storageName, componentName)
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
		odoutil.LogErrorAndExit(err, "")
		if !exists {

			log.Errorf("The storage %v does not exists in the application %v", storageName, applicationName)
			os.Exit(1)
		}

		componentName, err := storage.GetComponentNameFromStorageName(client, storageName)
		if err != nil {
			odoutil.LogErrorAndExit(err, "Unable to get component associated with %s storage.", storageName)
		}

		var confirmDeletion string
		if storageForceDeleteflag {
			confirmDeletion = "y"
		} else {
			if componentName != "" {
				mPath := storage.GetMountPath(client, storageName, componentName, applicationName)
				log.Askf("Are you sure you want to delete the storage %v mounted to %v in %v component? [y/N]: ", storageName, mPath, componentName)
			} else {
				log.Askf("Are you sure you want to delete the storage %v that is not currently mounted to any component? [y/N]: ", storageName)
			}
			fmt.Scanln(&confirmDeletion)
		}
		if strings.ToLower(confirmDeletion) == "y" {
			componentName, err = storage.Delete(client, storageName, applicationName)
			odoutil.LogErrorAndExit(err, "failed to delete storage")
			if componentName != "" {
				log.Infof("Deleted storage %v from %v", storageName, componentName)
			} else {
				log.Infof("Deleted storage %v", storageName)
			}
		} else {
			log.Errorf("Aborting deletion of storage: %v", storageName)
			os.Exit(1)
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
			if len(genericclioptions.FlagValueIfSet(cmd, genericclioptions.ComponentFlagName)) > 0 {
				log.Error("Invalid arguments. Component name is not needed")
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
		odoutil.LogErrorAndExit(err, "unable to check if the storage exists in the current application")
		if !exists {
			log.Errorf("The storage %v does not exists in the current application '%v'", storageName, applicationName)
			os.Exit(1)
		}
		isMounted, err := storage.IsMounted(client, storageName, componentName, applicationName)
		odoutil.LogErrorAndExit(err, "unable to check if the component is already mounted or not")
		if isMounted {
			log.Errorf("The storage %v is already mounted on the current component '%v'", storageName, componentName)
			os.Exit(1)
		}
		err = storage.Mount(client, storagePath, storageName, componentName, applicationName)
		odoutil.LogErrorAndExit(err, "")
		log.Infof("The storage %v is successfully mounted to the current component '%v'", storageName, componentName)
	},
}

// NewCmdStorage implements the odo storage command
func NewCmdStorage() *cobra.Command {
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
	projectCmd.AddProjectFlag(storageCreateCmd)
	projectCmd.AddProjectFlag(storageDeleteCmd)
	projectCmd.AddProjectFlag(storageListCmd)
	projectCmd.AddProjectFlag(storageMountCmd)
	projectCmd.AddProjectFlag(storageUnmountCmd)

	//Adding `--application` flag
	appCmd.AddApplicationFlag(storageCreateCmd)
	appCmd.AddApplicationFlag(storageDeleteCmd)
	appCmd.AddApplicationFlag(storageListCmd)
	appCmd.AddApplicationFlag(storageMountCmd)
	appCmd.AddApplicationFlag(storageUnmountCmd)

	//Adding `--component` flag
	componentCmd.AddComponentFlag(storageCreateCmd)
	componentCmd.AddComponentFlag(storageDeleteCmd)
	componentCmd.AddComponentFlag(storageListCmd)
	componentCmd.AddComponentFlag(storageMountCmd)
	componentCmd.AddComponentFlag(storageUnmountCmd)

	// Add a defined annotation in order to appear in the help menu
	storageCmd.Annotations = map[string]string{"command": "other"}
	storageCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)

	completion.RegisterCommandHandler(storageDeleteCmd, completion.StorageDeleteCompletionHandler)
	completion.RegisterCommandHandler(storageMountCmd, completion.StorageMountCompletionHandler)
	completion.RegisterCommandHandler(storageUnmountCmd, completion.StorageUnMountCompletionHandler)

	return storageCmd
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
	odoutil.LogErrorAndExit(err, "could not get mounted storage list")

	// iterating over all mounted storage and put in the mount storage table
	if len(storageListMounted) > 0 {
		for _, mStorage := range storageListMounted {
			fmt.Fprintln(tabWriterMounted, mStorage.Name, "\t", mStorage.Size, "\t", mStorage.Path)
		}

		// print all mounted storage of the given component
		log.Infof("The component '%v' has the following storage attached:", componentName)
		tabWriterMounted.Flush()
	} else {
		log.Infof("The component '%v' has no storage attached", componentName)
	}
	fmt.Println("")
}

// printMountedStorageInAllComponent prints all the mounted storage in all the components of the application and project
func printMountedStorageInAllComponent(client *occlient.Client, applicationName string) {
	componentList, err := component.List(client, applicationName)
	odoutil.LogErrorAndExit(err, "could not get component list")

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
	odoutil.LogErrorAndExit(err, "could not get unmounted storage list")

	// iterating over all unmounted storage and put in the unmount storage table
	if len(storageListUnmounted) > 0 {
		for _, uStorage := range storageListUnmounted {
			fmt.Fprintln(tabWriterUnmounted, uStorage.Name, "\t", uStorage.Size)
		}

		// print unmounted storage of all the application
		log.Info("Storage that are not mounted to any component:")
		tabWriterUnmounted.Flush()
	}
	fmt.Println("")
}
