package storage

import (
	"fmt"
	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/odo/util/completion"
	"github.com/openshift/odo/pkg/storage"
	"github.com/openshift/odo/pkg/util"
	"os"

	"text/tabwriter"

	"github.com/openshift/odo/pkg/log"
	appCmd "github.com/openshift/odo/pkg/odo/cli/application"
	componentCmd "github.com/openshift/odo/pkg/odo/cli/component"
	projectCmd "github.com/openshift/odo/pkg/odo/cli/project"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	ktemplates "k8s.io/kubernetes/pkg/kubectl/cmd/templates"

	"github.com/spf13/cobra"
)

const listRecommendedCommandName = "list"

var (
	storageListShortDesc = `List storage attached to a component`
	storageListLongDesc  = ktemplates.LongDesc(`List storage attached to a component`)
	storageListExample   = ktemplates.Examples(`  # List all storage attached or mounted to the current component and 
  # all unattached or unmounted storage in the current application
  %[1]s
	`)
)

type StorageListOptions struct {
	componentContext string
	localConfig      *config.LocalConfigInfo
	*genericclioptions.Context
}

// NewStorageListOptions creates a new StorageListOptions instance
func NewStorageListOptions() *StorageListOptions {
	return &StorageListOptions{}
}

// Complete completes StorageListOptions after they've been created
func (o *StorageListOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	o.Context = genericclioptions.NewContext(cmd)
	o.localConfig, err = config.NewLocalConfigInfo(o.componentContext)
	if err != nil {
		return err
	}
	return
}

// Validate validates the StorageListOptions based on completed values
func (o *StorageListOptions) Validate() (err error) {
	// check the machine readable format
	if !util.CheckOutputFlag(o.OutputFlag) {
		return fmt.Errorf("given output format %s is not supported", o.OutputFlag)
	}
	return nil
}

// Run contains the logic for the odo storage list command
func (o *StorageListOptions) Run() (err error) {
	storageListConfig, err := o.localConfig.StorageList()
	if err != nil {
		return err
	}

	storageListMachineReadable := []storage.Storage{}

	for _, storageConfig := range storageListConfig {
		storageListMachineReadable = append(storageListMachineReadable, storage.GetMachineReadableFormat(storageConfig.Name, storageConfig.Size, storageConfig.Path))
	}
	storageListResultMachineReadable := storage.GetMachineReadableFormatForList(storageListMachineReadable)
	out, err := util.MachineOutput(o.OutputFlag, storageListResultMachineReadable)
	if err != nil {
		return err
	}

	if out != "" {
		fmt.Println(out)
	} else {
		// defining the column structure of the table
		tabWriterMounted := tabwriter.NewWriter(os.Stdout, 5, 2, 3, ' ', tabwriter.TabIndent)

		// create headers of mounted storage table
		fmt.Fprintln(tabWriterMounted, "NAME", "\t", "SIZE", "\t", "PATH")
		// iterating over all mounted storage and put in the mount storage table
		if len(storageListConfig) > 0 {
			for _, mStorage := range storageListConfig {
				fmt.Fprintln(tabWriterMounted, mStorage.Name, "\t", mStorage.Size, "\t", mStorage.Path)
			}

			// print all mounted storage of the given component
			log.Infof("The component '%v' has the following storage attached:", o.localConfig.GetName())
			tabWriterMounted.Flush()
		} else {
			log.Infof("The component '%v' has no storage attached", o.localConfig.GetName())
		}
		fmt.Println("")
	}
	return
}

// NewCmdStorageList implements the odo storage list command.
func NewCmdStorageList(name, fullName string) *cobra.Command {
	o := NewStorageListOptions()
	storageListCmd := &cobra.Command{
		Use:     name,
		Short:   storageListShortDesc,
		Long:    storageListLongDesc,
		Example: fmt.Sprintf(storageListExample, fullName),
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	projectCmd.AddProjectFlag(storageListCmd)
	appCmd.AddApplicationFlag(storageListCmd)
	componentCmd.AddComponentFlag(storageListCmd)

	genericclioptions.AddOutputFlag(storageListCmd)
	genericclioptions.AddContextFlag(storageListCmd, &o.componentContext)
	completion.RegisterCommandFlagHandler(storageListCmd, "context", completion.FileCompletionHandler)

	return storageListCmd
}
