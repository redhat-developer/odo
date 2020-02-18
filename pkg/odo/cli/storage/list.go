package storage

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/machineoutput"
	"github.com/openshift/odo/pkg/odo/util/completion"
	"github.com/openshift/odo/pkg/storage"

	"github.com/openshift/odo/pkg/odo/genericclioptions"
	ktemplates "k8s.io/kubernetes/pkg/kubectl/util/templates"

	"github.com/spf13/cobra"
)

const listRecommendedCommandName = "list"

var (
	storageListShortDesc = `List storage attached to a component`
	storageListLongDesc  = ktemplates.LongDesc(`List storage attached to a component`)
	storageListExample   = ktemplates.Examples(`
	# List all storage attached or mounted to the current component and 
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
	return nil
}

func (o *StorageListOptions) Run() (err error) {

	storageList, err := storage.ListStorageWithState(o.Client, o.localConfig, o.Component(), o.Application)
	if err != nil {
		return err
	}

	if log.IsJSON() {
		machineoutput.OutputSuccess(storageList)
	} else {
		printStorage(storageList, o.localConfig.GetName())
	}

	return
}

func printStorage(storageList storage.StorageList, compName string) {

	if len(storageList.Items) > 0 {

		tabWriterMounted := tabwriter.NewWriter(os.Stdout, 5, 2, 3, ' ', tabwriter.TabIndent)

		// create headers of mounted storage table
		fmt.Fprintln(tabWriterMounted, "NAME", "\t", "SIZE", "\t", "PATH", "\t", "STATE")
		// iterating over all mounted storage and put in the mount storage table
		for _, mStorage := range storageList.Items {
			fmt.Fprintln(tabWriterMounted, mStorage.Name, "\t", mStorage.Spec.Size, "\t", mStorage.Status.Path, "\t", mStorage.State)
		}

		// print all mounted storage of the given component
		log.Infof("The component '%v' has the following storage attached:", compName)
		tabWriterMounted.Flush()
	} else {
		log.Infof("The component '%v' has no storage attached", compName)
	}

	fmt.Println("")
}

// NewCmdStorageList implements the odo storage list command.
func NewCmdStorageList(name, fullName string) *cobra.Command {
	o := NewStorageListOptions()
	storageListCmd := &cobra.Command{
		Use:         name,
		Short:       storageListShortDesc,
		Long:        storageListLongDesc,
		Example:     fmt.Sprintf(storageListExample, fullName),
		Args:        cobra.MaximumNArgs(1),
		Annotations: map[string]string{"machineoutput": "json"},
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	genericclioptions.AddContextFlag(storageListCmd, &o.componentContext)
	completion.RegisterCommandFlagHandler(storageListCmd, "context", completion.FileCompletionHandler)

	return storageListCmd
}
