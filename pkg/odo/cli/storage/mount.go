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
)

const mountRecommendedCommandName = "mount"

var (
	storageMountShortDesc = `mount storage to a component`
	storageMountLongDesc  = ktemplates.LongDesc(`mount storage to a component`)
	storageMountExample   = ktemplates.Examples(` # Mount storage 'dbstorage' to current component
  %[1]s dbstorage --path /data

  # Mount storage 'database' to component 'mongodb'
  %[1]s database --component mongodb --path /data`)
)

type StorageMountOptions struct {
	storageName string
	*genericclioptions.Context
}

// NewStorageCreateOptions creates a new UrlCreateOptions instance
func NewStorageMountOptions() *StorageMountOptions {
	return &StorageMountOptions{}
}

// Complete completes StorageMountOptions after they've been Created
func (o *StorageMountOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	o.Context = genericclioptions.NewContext(cmd)
	o.storageName = args[0]
	return
}

// Validate validates the StorageMountOptions based on completed values
func (o *StorageMountOptions) Validate() (err error) {
	exists, err := storage.Exists(o.Client, o.storageName, o.Application)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the storage %v does not exists in the current application '%v'", o.storageName, o.Application)
	}
	isMounted, err := storage.IsMounted(o.Client, o.storageName, o.Component(), o.Application)
	if err != nil {
		return fmt.Errorf("unable to check if the component is already mounted or not, cause %v", err)
	}
	if isMounted {
		return fmt.Errorf("the storage %v is already mounted on the current component '%v'", o.storageName, o.Component())
	}
	return
}

// Run contains the logic for the odo storage list command
func (o *StorageMountOptions) Run() (err error) {
	err = storage.Mount(o.Client, storagePath, o.storageName, o.Component(), o.Application)
	if err != nil {
		return
	}
	log.Infof("The storage %v is successfully mounted to the current component '%v'", o.storageName, o.Component())
	return
}

// NewCmdStorageCreate implements the odo storage create command.
func NewCmdStorageMount(name, fullName string) *cobra.Command {
	o := NewStorageMountOptions()
	storageMountCmd := &cobra.Command{
		Use:     name + " [storage name]",
		Short:   storageMountShortDesc,
		Long:    storageMountLongDesc,
		Example: fmt.Sprintf(storageMountExample, fullName),
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			odoutil.LogErrorAndExit(o.Complete(name, cmd, args), "")
			odoutil.LogErrorAndExit(o.Validate(), "")
			odoutil.LogErrorAndExit(o.Run(), "")
		},
	}

	storageMountCmd.Flags().StringVar(&storagePath, "path", "", "Path to mount the storage on")
	_ = storageMountCmd.MarkFlagRequired("path")

	projectCmd.AddProjectFlag(storageMountCmd)
	appCmd.AddApplicationFlag(storageMountCmd)
	componentCmd.AddComponentFlag(storageMountCmd)

	completion.RegisterCommandHandler(storageMountCmd, completion.StorageMountCompletionHandler)

	return storageMountCmd
}
