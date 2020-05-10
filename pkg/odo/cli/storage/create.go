package storage

import (
	"fmt"

	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/machineoutput"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/odo/util/completion"
	"github.com/openshift/odo/pkg/storage"
	"github.com/openshift/odo/pkg/util"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

const createRecommendedCommandName = "create"

var (
	storageCreateShortDesc = `Create storage and mount to a component`
	storageCreateLongDesc  = ktemplates.LongDesc(`Create storage and mount to a component`)
	storageCreateExample   = ktemplates.Examples(`
	# Create storage of size 1Gb to a component
  %[1]s mystorage --path=/opt/app-root/src/storage/ --size=1Gi
	`)
)

type StorageCreateOptions struct {
	storageName      string
	storageSize      string
	storagePath      string
	componentContext string
	*genericclioptions.Context
}

// NewStorageCreateOptions creates a new StorageCreateOptions instance
func NewStorageCreateOptions() *StorageCreateOptions {
	return &StorageCreateOptions{}
}

// Complete completes StorageCreateOptions after they've been created
func (o *StorageCreateOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	o.Context = genericclioptions.NewContext(cmd)
	if len(args) != 0 {
		o.storageName = args[0]
	} else {
		o.storageName = o.Component() + "-" + util.GenerateRandomString(4)
	}
	return
}

// Validate validates the StorageCreateOptions based on completed values
func (o *StorageCreateOptions) Validate() (err error) {
	// validate storage path
	return o.LocalConfigInfo.ValidateStorage(o.storageName, o.storagePath)
}

// Run contains the logic for the odo storage create command
func (o *StorageCreateOptions) Run() (err error) {
	storageResult, err := o.LocalConfigInfo.StorageCreate(o.storageName, o.storageSize, o.storagePath)
	if err != nil {
		return err
	}

	storageResultMachineReadable := storage.GetMachineReadableFormat(storageResult.Name, storageResult.Size, storageResult.Path)

	if log.IsJSON() {
		machineoutput.OutputSuccess(storageResultMachineReadable)
	} else {
		log.Successf("Added storage %v to %v", o.storageName, o.LocalConfigInfo.GetName())
		log.Italic("\nPlease use `odo push` command to make the storage accessible to the component")
	}
	return
}

// NewCmdStorageCreate implements the odo storage create command.
func NewCmdStorageCreate(name, fullName string) *cobra.Command {
	o := NewStorageCreateOptions()
	storageCreateCmd := &cobra.Command{
		Use:         name,
		Short:       storageCreateShortDesc,
		Long:        storageCreateLongDesc,
		Example:     fmt.Sprintf(storageCreateExample, fullName),
		Args:        cobra.MaximumNArgs(1),
		Annotations: map[string]string{"machineoutput": "json"},
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	storageCreateCmd.Flags().StringVar(&o.storageSize, "size", "", "Size of storage to add")
	storageCreateCmd.Flags().StringVar(&o.storagePath, "path", "", "Path to mount the storage on")
	_ = storageCreateCmd.MarkFlagRequired("path")
	_ = storageCreateCmd.MarkFlagRequired("size")

	genericclioptions.AddContextFlag(storageCreateCmd, &o.componentContext)
	completion.RegisterCommandFlagHandler(storageCreateCmd, "context", completion.FileCompletionHandler)

	return storageCreateCmd
}
