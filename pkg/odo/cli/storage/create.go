package storage

import (
	"fmt"

	"github.com/redhat-developer/odo/pkg/localConfigProvider"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/machineoutput"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/odo/util/completion"
	"github.com/redhat-developer/odo/pkg/storage"
	"github.com/redhat-developer/odo/pkg/util"
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

	# Create storage with ephemeral volume of size 2Gi to a component
  %[1]s mystorage --path=/opt/app-root/src/storage/ --size=2Gi --ephemeral
	`)
)

type CreateOptions struct {
	// Context
	*genericclioptions.Context

	// Parameters
	storageName string

	// Flags
	sizeFlag      string
	pathFlag      string
	contextFlag   string
	containerFlag string
	ephemeralFlag bool

	storage localConfigProvider.LocalStorage
}

// NewStorageCreateOptions creates a new CreateOptions instance
func NewStorageCreateOptions() *CreateOptions {
	return &CreateOptions{}
}

// Complete completes CreateOptions after they've been created
func (o *CreateOptions) Complete(name string, cmdline cmdline.Cmdline, args []string) (err error) {
	o.Context, err = genericclioptions.New(genericclioptions.NewCreateParameters(cmdline).NeedDevfile(o.contextFlag))
	if err != nil {
		return err
	}

	if len(args) != 0 {
		o.storageName = args[0]
	} else {
		o.storageName = fmt.Sprintf("%s-%s", o.Context.LocalConfigProvider.GetName(), util.GenerateRandomString(4))
	}

	var eph *bool
	if o.ephemeralFlag {
		eph = &o.ephemeralFlag
	}
	o.storage = localConfigProvider.LocalStorage{
		Name:      o.storageName,
		Size:      o.sizeFlag,
		Ephemeral: eph,
		Path:      o.pathFlag,
		Container: o.containerFlag,
	}

	o.Context.LocalConfigProvider.CompleteStorage(&o.storage)

	return nil
}

// Validate validates the CreateOptions based on completed values
func (o *CreateOptions) Validate() (err error) {
	// validate the storage
	return o.LocalConfigProvider.ValidateStorage(o.storage)
}

// Run contains the logic for the odo storage create command
func (o *CreateOptions) Run() (err error) {
	err = o.Context.LocalConfigProvider.CreateStorage(o.storage)
	if err != nil {
		return err
	}

	if log.IsJSON() {
		storageResultMachineReadable := storage.NewStorage(o.storage.Name, o.storage.Size, o.storage.Path, nil)
		machineoutput.OutputSuccess(storageResultMachineReadable)
	} else {
		log.Successf("Added storage %v to %v", o.storageName, o.Context.LocalConfigProvider.GetName())

		log.Italic("\nPlease use `odo push` command to make the storage accessible to the component")
	}
	return nil
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

	storageCreateCmd.Flags().StringVar(&o.sizeFlag, "size", "", "Size of storage to add")
	storageCreateCmd.Flags().StringVar(&o.pathFlag, "path", "", "Path to mount the storage on")
	storageCreateCmd.Flags().StringVar(&o.containerFlag, "container", "", "Name of container to attach the storage to in devfile")
	storageCreateCmd.Flags().BoolVar(&o.ephemeralFlag, "ephemeral", false, "Set volume as ephemeral")

	odoutil.AddContextFlag(storageCreateCmd, &o.contextFlag)
	completion.RegisterCommandFlagHandler(storageCreateCmd, "context", completion.FileCompletionHandler)

	return storageCreateCmd
}
