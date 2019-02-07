package storage

import (
	"fmt"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/storage"
	"github.com/redhat-developer/odo/pkg/util"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubernetes/pkg/kubectl/cmd/templates"

	appCmd "github.com/redhat-developer/odo/pkg/odo/cli/application"
	componentCmd "github.com/redhat-developer/odo/pkg/odo/cli/component"
	projectCmd "github.com/redhat-developer/odo/pkg/odo/cli/project"
)

const createRecommendedCommandName = "create"

var (
	urlCreateShortDesc = `Create storage and mount to a component`
	urlCreateLongDesc  = ktemplates.LongDesc(`Create storage and mount to a component`)
	urlCreateExample   = ktemplates.Examples(`  # Create storage of size 1Gb to a component
  %[1]s mystorage --path=/opt/app-root/src/storage/ --size=1Gi
	`)
)

type StorageCreateOptions struct {
	storageName string
	storageSize string
	storagePath string
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
	return validateStoragePath(o.Client, o.storageName, o.Component(), o.Application)
}

// Run contains the logic for the odo storage create command
func (o *StorageCreateOptions) Run() (err error) {
	_, err = storage.Create(o.Client, o.storageName, o.storageSize, o.storagePath, o.Component(), o.Application)
	if err == nil {
		log.Successf("Added storage %v to %v", o.storageName, o.Component())
	}
	return
}

// NewCmdStorageCreate implements the odo storage create command.
func NewCmdStorageCreate(name, fullName string) *cobra.Command {
	o := NewStorageCreateOptions()
	storageCreateCmd := &cobra.Command{
		Use:     name,
		Short:   urlCreateShortDesc,
		Long:    urlCreateLongDesc,
		Example: fmt.Sprintf(urlCreateExample, fullName),
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	storageCreateCmd.Flags().StringVar(&o.storageSize, "size", "", "Size of storage to add")
	storageCreateCmd.Flags().StringVar(&o.storagePath, "path", "", "Path to mount the storage on")
	_ = storageCreateCmd.MarkFlagRequired("path")
	_ = storageCreateCmd.MarkFlagRequired("size")

	projectCmd.AddProjectFlag(storageCreateCmd)
	appCmd.AddApplicationFlag(storageCreateCmd)
	componentCmd.AddComponentFlag(storageCreateCmd)

	return storageCreateCmd
}
