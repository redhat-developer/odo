package storage

import (
	"fmt"

	"github.com/openshift/odo/pkg/devfile/location"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/machineoutput"
	"github.com/openshift/odo/pkg/odo/cli/ui"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/odo/util/completion"
	"github.com/openshift/odo/pkg/storage"
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

type DeleteOptions struct {
	storageName            string
	storageForceDeleteFlag bool
	componentContext       string

	*genericclioptions.Context
}

// NewStorageDeleteOptions creates a new DeleteOptions instance
func NewStorageDeleteOptions() *DeleteOptions {
	return &DeleteOptions{}
}

// Complete completes DeleteOptions after they've been created
func (o *DeleteOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	o.Context, err = genericclioptions.New(genericclioptions.CreateParameters{
		Cmd:              cmd,
		DevfilePath:      location.DevfileFilenamesProvider(o.componentContext),
		ComponentContext: o.componentContext,
	})

	if err != nil {
		return err
	}

	o.storageName = args[0]
	return
}

// Validate validates the DeleteOptions based on completed values
func (o *DeleteOptions) Validate() (err error) {
	gotStorage, err := o.LocalConfigProvider.GetStorage(o.storageName)
	if err != nil {
		return err
	}
	if gotStorage == nil {
		return fmt.Errorf("the storage %v does not exists in the application %v, cause %v", o.storageName, o.Application, err)
	}

	return
}

// Run contains the logic for the odo storage delete command
func (o *DeleteOptions) Run(cmd *cobra.Command) (err error) {
	mPath, err := o.Context.LocalConfigProvider.GetStorageMountPath(o.storageName)
	if err != nil {
		return err
	}

	deleteMsg := fmt.Sprintf("Are you sure you want to delete the storage %v mounted to %v in %v component", o.storageName, mPath, o.Context.LocalConfigProvider.GetName())

	if log.IsJSON() || o.storageForceDeleteFlag || ui.Proceed(deleteMsg) {
		err := o.Context.LocalConfigProvider.DeleteStorage(o.storageName)
		if err != nil {
			return fmt.Errorf("failed to delete storage, cause %v", err)
		}

		successMessage := fmt.Sprintf("Deleted storage %v from %v", o.storageName, o.Context.LocalConfigProvider.GetName())

		log.Infof(successMessage)
		log.Italic("\nPlease use `odo push` command to delete the storage from the cluster")

		if log.IsJSON() {
			machineoutput.SuccessStatus(storage.StorageKind, o.storageName, successMessage)
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
		Use:         name,
		Short:       storageDeleteShortDesc,
		Long:        storageDeleteLongDesc,
		Example:     fmt.Sprintf(storageDeleteExample, fullName),
		Args:        cobra.ExactArgs(1),
		Annotations: map[string]string{"machineoutput": "json"},
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
