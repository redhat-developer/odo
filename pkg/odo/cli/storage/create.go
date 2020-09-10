package storage

import (
	"fmt"
	"github.com/openshift/odo/pkg/devfile"
	adapterCommon "github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/machineoutput"
	"github.com/openshift/odo/pkg/odo/cli/component"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/odo/util/completion"
	"github.com/openshift/odo/pkg/storage"
	"github.com/openshift/odo/pkg/util"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"
	"path/filepath"
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

	devfilePath   string
	isDevfile     bool
	componentName string
	*genericclioptions.Context
}

// NewStorageCreateOptions creates a new StorageCreateOptions instance
func NewStorageCreateOptions() *StorageCreateOptions {
	return &StorageCreateOptions{devfilePath: component.DevfilePath}
}

// Complete completes StorageCreateOptions after they've been created
func (o *StorageCreateOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	o.devfilePath = filepath.Join(o.componentContext, o.devfilePath)
	o.isDevfile = util.CheckPathExists(o.devfilePath)
	if o.isDevfile {
		o.Context = genericclioptions.NewDevfileContext(cmd)

		o.componentName = o.EnvSpecificInfo.GetName()
		if o.storageSize == "" {
			o.storageSize = adapterCommon.DefaultVolumeSize
		}
	} else {
		o.Context = genericclioptions.NewContext(cmd)
		o.componentName = o.LocalConfigInfo.GetName()

		if o.storageSize == "" {
			return fmt.Errorf("\"size\" flag is required for s2i components")
		}
	}

	if len(args) != 0 {
		o.storageName = args[0]
	} else {
		o.storageName = fmt.Sprintf("%s-%s", o.componentName, util.GenerateRandomString(4))
	}
	return
}

// Validate validates the StorageCreateOptions based on completed values
func (o *StorageCreateOptions) Validate() (err error) {
	if o.isDevfile {
		return
	}
	// validate storage path
	return o.LocalConfigInfo.ValidateStorage(o.storageName, o.storagePath)
}

func (o *StorageCreateOptions) devfileRun() error {
	devFile, err := devfile.ParseAndValidate(o.devfilePath)
	if err != nil {
		return err
	}

	err = devFile.Data.AddVolume(common.DevfileComponent{
		Name: o.storageName,
		Volume: &common.Volume{
			Size: o.storageSize,
		},
	}, o.storagePath)

	if err != nil {
		return err
	}
	err = devFile.WriteYamlDevfile()
	if err != nil {
		return err
	}
	return nil
}

// Run contains the logic for the odo storage create command
func (o *StorageCreateOptions) Run() (err error) {
	if o.isDevfile {
		err := o.devfileRun()
		if err != nil {
			return err
		}
	} else {
		_, err := o.LocalConfigInfo.StorageCreate(o.storageName, o.storageSize, o.storagePath)
		if err != nil {
			return err
		}
	}

	storageResultMachineReadable := storage.GetMachineReadableFormat(o.storageName, o.storageSize, o.storagePath)

	if log.IsJSON() {
		machineoutput.OutputSuccess(storageResultMachineReadable)
	} else {
		log.Successf("Added storage %v to %v", o.storageName, o.componentName)

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

	genericclioptions.AddContextFlag(storageCreateCmd, &o.componentContext)
	completion.RegisterCommandFlagHandler(storageCreateCmd, "context", completion.FileCompletionHandler)

	return storageCreateCmd
}
