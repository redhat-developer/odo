package storage

import (
	"fmt"

	"github.com/redhat-developer/odo/pkg/log"
	appCmd "github.com/redhat-developer/odo/pkg/odo/cli/application"
	componentCmd "github.com/redhat-developer/odo/pkg/odo/cli/component"
	projectCmd "github.com/redhat-developer/odo/pkg/odo/cli/project"
	"github.com/redhat-developer/odo/pkg/odo/cli/ui"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/util/completion"
	"github.com/redhat-developer/odo/pkg/storage"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubernetes/pkg/kubectl/cmd/templates"
)

const deleteRecommendedCommandName = "delete"

var (
	storageDeleteShortDesc = `Delete storage from component`
	storageDeleteLongDesc  = ktemplates.LongDesc(`Delete storage from component`)
	storageDeleteExample   = ktemplates.Examples(`  # Delete storage mystorage from the currently active component
  %[1]s mystorage

  # Delete storage mystorage from component 'mongodb'
  %[1]s mystorage --component mongodb
`)
)

type StorageDeleteOptions struct {
	storageName            string
	storageForceDeleteFlag bool
	componentName          string
	*genericclioptions.Context
}

// NewStorageDeleteOptions creates a new StorageDeleteOptions instance
func NewStorageDeleteOptions() *StorageDeleteOptions {
	return &StorageDeleteOptions{}
}

// Complete completes StorageDeleteOptions after they've been created
func (o *StorageDeleteOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	o.Context = genericclioptions.NewContext(cmd)
	o.storageName = args[0]
	return
}

// Validate validates the StorageDeleteOptions based on completed values
func (o *StorageDeleteOptions) Validate() (err error) {
	exists, err := storage.Exists(o.Client, o.storageName, o.Application)

	if err != nil {
		return
	}
	if !exists {
		return fmt.Errorf("the storage %v does not exists in the application %v, cause %v", o.storageName, o.Application, err)
	}

	o.componentName, err = storage.GetComponentNameFromStorageName(o.Client, o.storageName)
	if err != nil {
		return fmt.Errorf("unable to get component associated with %s storage, cause %v", o.storageName, err)
	}

	return
}

// Run contains the logic for the odo storage delete command
func (o *StorageDeleteOptions) Run() (err error) {
	var deleteMsg string
	if o.componentName != "" {
		mPath := storage.GetMountPath(o.Client, o.storageName, o.componentName, o.Application)
		deleteMsg = fmt.Sprintf("Are you sure you want to delete the storage %v mounted to %v in %v component", o.storageName, mPath, o.componentName)
	} else {
		deleteMsg = fmt.Sprintf("Are you sure you want to delete the storage %v that is not currently mounted to any component", o.storageName)
	}
	if o.storageForceDeleteFlag || ui.Proceed(deleteMsg) {
		o.componentName, err = storage.Delete(o.Client, o.storageName, o.Application)
		if err != nil {
			return fmt.Errorf("failed to delete storage, cause %v", err)
		}
		if o.componentName != "" {
			log.Infof("Deleted storage %v from %v", o.storageName, o.componentName)
		} else {
			log.Infof("Deleted storage %v", o.storageName)
		}
	} else {
		return fmt.Errorf("aborting deletion of storage: %v", o.storageName)
	}

	return
}

// NewCmdStorageDelete implements the odo storage delete command.
func NewCmdStorageDelete(name, fullName string) *cobra.Command {
	o := NewStorageDeleteOptions()
	storageDeleteCmd := &cobra.Command{
		Use:     name,
		Short:   storageDeleteShortDesc,
		Long:    storageDeleteLongDesc,
		Example: fmt.Sprintf(storageDeleteExample, fullName),
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	storageDeleteCmd.Flags().BoolVarP(&o.storageForceDeleteFlag, "force", "f", false, "Delete storage without prompting")
	completion.RegisterCommandHandler(storageDeleteCmd, completion.StorageDeleteCompletionHandler)

	projectCmd.AddProjectFlag(storageDeleteCmd)
	appCmd.AddApplicationFlag(storageDeleteCmd)
	componentCmd.AddComponentFlag(storageDeleteCmd)

	return storageDeleteCmd
}
