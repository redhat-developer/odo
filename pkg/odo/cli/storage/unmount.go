package storage

import (
	"fmt"
	"os"
	"strings"

	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/odo/util/completion"
	"github.com/openshift/odo/pkg/storage"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

//const unMountRecommendedCommandName = "unmount"

var (
	storageUnMountShortDesc = `Unmount storage from the given path or identified by its name, from the current component`
	storageUnMountLongDesc  = ktemplates.LongDesc(`Unmount storage from the given path or identified by its name, from the current component.

  The storage and the contents are not deleted, the storage or storage with the given path, is only unmounted from the component, and hence is no longer accessible by the component.`)
	storageUnMountExample = ktemplates.Examples(`  # Unmount storage 'dbstorage' from current component
  %[1]s dbstorage

  # Unmount storage mounted to path '/data' from current component
  %[1]s /data
`)
)

type StorageUnMountOptions struct {
	storageName string
	storagePath string
	*genericclioptions.Context
}

// NewStorageUnMountOptions creates a new StorageUnMountOptions instance
func NewStorageUnMountOptions() *StorageUnMountOptions {
	return &StorageUnMountOptions{}
}

// Complete completes StorageUnMountOptions after they've been created
func (o *StorageUnMountOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	o.Context = genericclioptions.NewContext(cmd)
	// checking if the first character in the argument is a "/", indicating a path or not, indicating a storage name
	if strings.HasPrefix(args[0], "/") {
		o.storagePath = args[0]
	} else {
		o.storageName = args[0]
	}
	return
}

// Validate validates the StorageUnMountOptions based on completed values
func (o *StorageUnMountOptions) Validate() (err error) {
	// checking if the first character in the argument is a "/", indicating a path or not, indicating a storage name
	if len(o.storagePath) > 0 {
		o.storageName, err = storage.GetStorageNameFromMountPath(o.Client, o.storagePath, o.Component(), o.Application)
		if err != nil {
			return fmt.Errorf("unable to get storage name from mount path, cause %v", err)
		}
		if o.storageName == "" {
			return fmt.Errorf("no storage is mounted to %s in the component %s", o.storagePath, o.Component())
		}
	} else {
		exists, err := storage.IsMounted(o.Client, o.storageName, o.Component(), o.Application)
		if err != nil {
			return fmt.Errorf("unable to check if storage is mounted or not, cause %v", err)
		}
		if !exists {
			log.Errorf("Storage %v does not exist in component %v", o.storageName, o.Component())
			os.Exit(1)
		}
	}
	return
}

// Run contains the logic for the odo storage unmount command
func (o *StorageUnMountOptions) Run() (err error) {
	err = storage.Unmount(o.Client, o.storageName, o.Component(), o.Application, true)
	if err != nil {
		return fmt.Errorf("unable to unmount storage %v from component %v", o.storageName, o.Component())
	}

	log.Infof("Unmounted storage %v from %v", o.storageName, o.Component())
	return
}

// NewCmdStorageUnMount implements the odo storage unmount command.
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
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	completion.RegisterCommandHandler(storageUnMountCmd, completion.StorageUnMountCompletionHandler)

	return storageUnMountCmd
}
