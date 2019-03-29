package storage

import (
	"fmt"
	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/log"
	appCmd "github.com/openshift/odo/pkg/odo/cli/application"
	componentCmd "github.com/openshift/odo/pkg/odo/cli/component"
	projectCmd "github.com/openshift/odo/pkg/odo/cli/project"
	"github.com/openshift/odo/pkg/odo/cli/ui"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/odo/util/completion"
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
	componentContext       string
	localConfig            *config.LocalConfigInfo
	*genericclioptions.Context
}

// NewStorageDeleteOptions creates a new StorageDeleteOptions instance
func NewStorageDeleteOptions() *StorageDeleteOptions {
	return &StorageDeleteOptions{}
}

// Complete completes StorageDeleteOptions after they've been created
func (o *StorageDeleteOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	o.Context = genericclioptions.NewContext(cmd)
	o.localConfig, err = config.NewLocalConfigInfo(o.componentContext)
	if err != nil {
		return err
	}
	o.Context = genericclioptions.NewContext(cmd)
	o.storageName = args[0]
	return
}

// Validate validates the StorageDeleteOptions based on completed values
func (o *StorageDeleteOptions) Validate() (err error) {
	exists := o.localConfig.StorageExists(o.storageName)
	if !exists {
		return fmt.Errorf("the storage %v does not exists in the application %v, cause %v", o.storageName, o.Application, err)
	}

	return
}

// Run contains the logic for the odo storage delete command
func (o *StorageDeleteOptions) Run() (err error) {
	var deleteMsg string

	mPath := o.localConfig.GetMountPath(o.storageName)

	deleteMsg = fmt.Sprintf("Are you sure you want to delete the storage %v mounted to %v in %v component", o.storageName, mPath, o.localConfig.GetName())

	if o.storageForceDeleteFlag || ui.Proceed(deleteMsg) {
		err = o.localConfig.StorageDelete(o.storageName)
		if err != nil {
			return fmt.Errorf("failed to delete storage, cause %v", err)
		}

		log.Infof("Deleted storage %v from %v", o.storageName, o.localConfig.GetName())
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

	genericclioptions.AddContextFlag(storageDeleteCmd, &o.componentContext)
	completion.RegisterCommandFlagHandler(storageDeleteCmd, "context", completion.FileCompletionHandler)

	return storageDeleteCmd
}
