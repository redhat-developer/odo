package storage

import (
	"fmt"

	"github.com/openshift/odo/pkg/localConfigProvider"
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

type CreateOptions struct {
	storageName      string
	storageSize      string
	storagePath      string
	componentContext string

	container string // container to which this storage belongs
	storage   localConfigProvider.LocalStorage
	*genericclioptions.Context
}

// NewStorageCreateOptions creates a new CreateOptions instance
func NewStorageCreateOptions() *CreateOptions {
	return &CreateOptions{}
}

// Complete completes CreateOptions after they've been created
func (o *CreateOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	o.Context, err = genericclioptions.New(genericclioptions.NewCreateParameters(cmd).NeedDevfile().SetComponentContext(o.componentContext))
	if err != nil {
		return err
	}

	if len(args) != 0 {
		o.storageName = args[0]
	} else {
		o.storageName = fmt.Sprintf("%s-%s", o.Context.LocalConfigProvider.GetName(), util.GenerateRandomString(4))
	}

	o.storage = localConfigProvider.LocalStorage{
		Name:      o.storageName,
		Size:      o.storageSize,
		Path:      o.storagePath,
		Container: o.container,
	}

	o.Context.LocalConfigProvider.CompleteStorage(&o.storage)

	return
}

// Validate validates the CreateOptions based on completed values
func (o *CreateOptions) Validate() (err error) {
	// validate the storage
	return o.LocalConfigProvider.ValidateStorage(o.storage)
}

// Run contains the logic for the odo storage create command
func (o *CreateOptions) Run(cmd *cobra.Command) (err error) {
	err = o.Context.LocalConfigProvider.CreateStorage(o.storage)
	if err != nil {
		return err
	}

	if log.IsJSON() {
		storageResultMachineReadable := storage.NewStorage(o.storage.Name, o.storage.Size, o.storage.Path)
		machineoutput.OutputSuccess(storageResultMachineReadable)
	} else {
		log.Successf("Added storage %v to %v", o.storageName, o.Context.LocalConfigProvider.GetName())

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
	storageCreateCmd.Flags().StringVar(&o.container, "container", "", "Name of container to attach the storage to in devfile")

	genericclioptions.AddContextFlag(storageCreateCmd, &o.componentContext)
	completion.RegisterCommandFlagHandler(storageCreateCmd, "context", completion.FileCompletionHandler)

	return storageCreateCmd
}
