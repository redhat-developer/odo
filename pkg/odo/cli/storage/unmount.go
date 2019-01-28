package storage

import (
	"fmt"
	"github.com/redhat-developer/odo/pkg/log"
	appCmd "github.com/redhat-developer/odo/pkg/odo/cli/application"
	componentCmd "github.com/redhat-developer/odo/pkg/odo/cli/component"
	projectCmd "github.com/redhat-developer/odo/pkg/odo/cli/project"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/odo/util/completion"
	"github.com/redhat-developer/odo/pkg/storage"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	"os"
)

const unMountRecommendedCommandName = "unmount"

var (
	storageUnMountShortDesc = `Unmount storage from the given path or identified by its name, from the current component`
	storageUnMountLongDesc  = ktemplates.LongDesc(`Unmount storage from the given path or identified by its name, from the current component.

  The storage and the contents are not deleted, the storage or storage with the given path, is only unmounted from the component, and hence is no longer accessible by the component.`)
	storageUnMountExample = ktemplates.Examples(`  # Unmount storage 'dbstorage' from current component
  %[1]s dbstorage

  # Unmount storage 'database' from component 'mongodb'
  %[1]s database --component mongodb

  # Unmount storage mounted to path '/data' from current component
  %[1]s /data

  # Unmount storage mounted to path '/data' from component 'mongodb'
  %[1]s /data --component mongodb
`)
)

type StorageUnMountOptions struct {
	storageName string
	storagePath string
	*genericclioptions.Context
}

// NewStorageCreateOptions creates a new UrlCreateOptions instance
func NewStorageUnMountOptions() *StorageUnMountOptions {
	return &StorageUnMountOptions{}
}

// Complete completes StorageMountOptions after they've been Created
func (o *StorageUnMountOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	o.Context = genericclioptions.NewContext(cmd)
	// checking if the first character in the argument is a "/", indicating a path or not, indicating a storage name
	if string(args[0][0]) == "/" {
		o.storagePath = args[0]
	} else {
		o.storageName = args[0]
	}
	return
}

// Validate validates the StorageMountOptions based on completed values
func (o *StorageUnMountOptions) Validate() (err error) {
	// checking if the first character in the argument is a "/", indicating a path or not, indicating a storage name
	if len(o.storagePath) > 0 {
		o.storageName, err = storage.GetStorageNameFromMountPath(o.Client, o.storagePath, o.Component(), o.Application)
		odoutil.LogErrorAndExit(err, "Unable to get storage name from mount path")
		if o.storageName == "" {
			log.Errorf("No storage is mounted to %s in the component %s", o.storagePath, o.Component())
			os.Exit(1)
		}
	} else {
		exists, err := storage.IsMounted(o.Client, o.storageName, o.Component(), o.Application)
		odoutil.LogErrorAndExit(err, "Unable to check if storage is mounted or not")
		if !exists {
			log.Errorf("Storage %v does not exist in component %v", o.storageName, o.Component())
			os.Exit(1)
		}
	}
	return
}

// Run contains the logic for the odo storage list command
func (o *StorageUnMountOptions) Run() (err error) {
	err = storage.Unmount(o.Client, o.storageName, o.Component(), o.Application, true)
	odoutil.LogErrorAndExit(err, "Unable to unmount storage %v from component %v", o.storageName, o.Component())

	log.Infof("Unmounted storage %v from %v", o.storageName, o.Component())
	return
}

// NewCmdStorageCreate implements the odo storage create command.
func NewCmdStorageUnMount(name, fullName string) *cobra.Command {
	o := NewStorageUnMountOptions()
	storageUnMountCmd := &cobra.Command{
		Use:     name + " PATH | STORAGE_NAME",
		Short:   storageUnMountShortDesc,
		Long:    storageUnMountLongDesc,
		Example: fmt.Sprintf(storageUnMountExample, fullName),
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"umount"},
		Run: func(cmd *cobra.Command, args []string) {
			odoutil.LogErrorAndExit(o.Complete(name, cmd, args), "")
			odoutil.LogErrorAndExit(o.Validate(), "")
			odoutil.LogErrorAndExit(o.Run(), "")
		},
	}

	projectCmd.AddProjectFlag(storageUnMountCmd)
	appCmd.AddApplicationFlag(storageUnMountCmd)
	componentCmd.AddComponentFlag(storageUnMountCmd)

	completion.RegisterCommandHandler(storageUnMountCmd, completion.StorageUnMountCompletionHandler)

	return storageUnMountCmd
}
