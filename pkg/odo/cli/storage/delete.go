package storage

import (
	"fmt"

	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/odo/cli/ui"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/odo/util/completion"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

const deleteRecommendedCommandName = "delete"

var (
	storageDeleteShortDesc = `Delete storage from component`
	storageDeleteLongDesc  = ktemplates.LongDesc(`Delete storage from component`)
	storageDeleteExample   = ktemplates.Examples(`
	# Delete storage mystorage from the currently active component
  %[1]s mystorage
`)
)

type StorageDeleteOptions struct {
	storageName            string
	storageForceDeleteFlag bool
	componentContext       string
	*genericclioptions.Context
}

// NewStorageDeleteOptions creates a new StorageDeleteOptions instance
func NewStorageDeleteOptions() *StorageDeleteOptions {
	return &StorageDeleteOptions{}
}

// Complete completes StorageDeleteOptions after they've been created
func (o *StorageDeleteOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	// this initializes the LocalConfigInfo as well
	o.Context = genericclioptions.NewContext(cmd)
	o.storageName = args[0]
	return
}

// Validate validates the StorageDeleteOptions based on completed values
func (o *StorageDeleteOptions) Validate() (err error) {
	exists := o.LocalConfigInfo.StorageExists(o.storageName)
	if !exists {
		return fmt.Errorf("the storage %v does not exists in the application %v, cause %v", o.storageName, o.Application, err)
	}

	return
}

// Run contains the logic for the odo storage delete command
func (o *StorageDeleteOptions) Run() (err error) {
	var deleteMsg string

	mPath := o.LocalConfigInfo.GetMountPath(o.storageName)

	deleteMsg = fmt.Sprintf("Are you sure you want to delete the storage %v mounted to %v in %v component", o.storageName, mPath, o.LocalConfigInfo.GetName())

	if o.storageForceDeleteFlag || ui.Proceed(deleteMsg) {
		err = o.LocalConfigInfo.StorageDelete(o.storageName)
		if err != nil {
			return fmt.Errorf("failed to delete storage, cause %v", err)
		}

		log.Infof("Deleted storage %v from %v", o.storageName, o.LocalConfigInfo.GetName())
		log.Italic("\nPlease use `odo push` command to delete the storage from the cluster")
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

	genericclioptions.AddContextFlag(storageDeleteCmd, &o.componentContext)
	completion.RegisterCommandFlagHandler(storageDeleteCmd, "context", completion.FileCompletionHandler)

	return storageDeleteCmd
}
