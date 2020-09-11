package storage

import (
	"fmt"
	"github.com/openshift/odo/pkg/devfile"
	"github.com/openshift/odo/pkg/devfile/parser"
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/odo/cli/component"
	odoutil "github.com/openshift/odo/pkg/util"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/machineoutput"
	"github.com/openshift/odo/pkg/odo/util/completion"
	"github.com/openshift/odo/pkg/storage"

	"github.com/openshift/odo/pkg/odo/genericclioptions"
	ktemplates "k8s.io/kubectl/pkg/util/templates"

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
	*genericclioptions.Context

	isDevfile bool
	parser.DevfileObj
}

// NewStorageListOptions creates a new StorageListOptions instance
func NewStorageListOptions() *StorageListOptions {
	return &StorageListOptions{}
}

// Complete completes StorageListOptions after they've been created
func (o *StorageListOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	devFilePath := filepath.Join(o.componentContext, component.DevfilePath)
	o.isDevfile = odoutil.CheckPathExists(devFilePath)
	if o.isDevfile {
		o.Context = genericclioptions.NewDevfileContext(cmd)

		o.DevfileObj, err = devfile.ParseAndValidate(devFilePath)
		if err != nil {
			return err
		}
	} else {
		// this also initializes the context as well
		o.Context = genericclioptions.NewContext(cmd)
	}
	return
}

// Validate validates the StorageListOptions based on completed values
func (o *StorageListOptions) Validate() (err error) {
	return nil
}

func (o *StorageListOptions) Run() (err error) {
	var storageList storage.StorageList
	var componentName string
	if o.isDevfile {
		componentName = o.EnvSpecificInfo.GetName()
		storageList, err = storage.DevfileList(o.KClient, o.DevfileObj.Data, o.EnvSpecificInfo.GetName())
		if err != nil {
			return err
		}
	} else {
		componentName = o.LocalConfigInfo.GetName()
		storageList, err = storage.ListStorageWithState(o.Client, o.LocalConfigInfo, o.Component(), o.Application)
		if err != nil {
			return err
		}
	}

	if log.IsJSON() {
		machineoutput.OutputSuccess(storageList)
	} else {
		if o.isDevfile && isContainerDisplay(storageList, o.DevfileObj.Data.GetComponents()) {
			printStorageWithContainer(storageList, componentName)
		} else {
			printStorage(storageList, componentName)
		}
	}

	return
}

// printStorage prints the given storageList
func printStorage(storageList storage.StorageList, compName string) {

	if len(storageList.Items) > 0 {

		tabWriterMounted := tabwriter.NewWriter(os.Stdout, 5, 2, 3, ' ', tabwriter.TabIndent)

		storageMap := make(map[string]bool)

		// create headers of mounted storage table
		fmt.Fprintln(tabWriterMounted, "NAME", "\t", "SIZE", "\t", "PATH", "\t", "STATE")
		// iterating over all mounted storage and put in the mount storage table
		for _, mStorage := range storageList.Items {
			_, ok := storageMap[mStorage.Name]
			if !ok {
				storageMap[mStorage.Name] = true
				fmt.Fprintln(tabWriterMounted, mStorage.Name, "\t", mStorage.Spec.Size, "\t", mStorage.Spec.Path, "\t", mStorage.Status)
			}
		}

		// print all mounted storage of the given component
		log.Infof("The component '%v' has the following storage attached:", compName)
		tabWriterMounted.Flush()
	} else {
		log.Infof("The component '%v' has no storage attached", compName)
	}

	fmt.Println("")
}

// printStorageWithContainer prints the given storageList with the corresponding container name
func printStorageWithContainer(storageList storage.StorageList, compName string) {

	if len(storageList.Items) > 0 {

		tabWriterMounted := tabwriter.NewWriter(os.Stdout, 5, 2, 3, ' ', tabwriter.TabIndent)

		// create headers of mounted storage table
		fmt.Fprintln(tabWriterMounted, "NAME", "\t", "SIZE", "\t", "PATH", "\t", "CONTAINER", "\t", "STATE")
		// iterating over all mounted storage and put in the mount storage table
		for _, mStorage := range storageList.Items {
			fmt.Fprintln(tabWriterMounted, mStorage.Name, "\t", mStorage.Spec.Size, "\t", mStorage.Spec.Path, "\t", mStorage.Spec.ContainerName, "\t", mStorage.Status)
		}

		// print all mounted storage of the given component
		log.Infof("The component '%v' has the following storage attached:", compName)
		tabWriterMounted.Flush()
	} else {
		log.Infof("The component '%v' has no storage attached", compName)
	}

	fmt.Println("")
}

// isContainerDisplay checks whether the container name should be included in the output
func isContainerDisplay(storageList storage.StorageList, components []common.DevfileComponent) bool {

	// get all the container names
	componentsMap := make(map[string]bool)
	for _, comp := range components {
		if comp.Container != nil {
			componentsMap[comp.Name] = true
		}
	}

	storageCompMap := make(map[string][]string)
	pathMap := make(map[string]string)
	storageMap := make(map[string]storage.StorageStatus)

	for _, storageItem := range storageList.Items {
		if pathMap[storageItem.Name] == "" {
			pathMap[storageItem.Name] = storageItem.Spec.Path
		}
		if storageMap[storageItem.Name] == "" {
			storageMap[storageItem.Name] = storageItem.Status
		}

		// check if the storage is mounted on the same path in all the containers
		if pathMap[storageItem.Name] != storageItem.Spec.Path {
			return true
		}

		// check if the storage is in the same state for all the containers
		if storageMap[storageItem.Name] != storageItem.Status {
			return true
		}

		// check if the storage is mounted on a valid devfile container
		// this situation can arrive when a container is removed from the devfile
		// but the state is not pushed thus it exists on the cluster
		_, ok := componentsMap[storageItem.Spec.ContainerName]
		if !ok {
			return true
		}
		storageCompMap[storageItem.Name] = append(storageCompMap[storageItem.Name], storageItem.Spec.ContainerName)
	}

	for _, containerNames := range storageCompMap {
		// check if the storage is mounted on all the devfile containers
		if len(containerNames) != len(componentsMap) {
			return true
		}
	}

	return false
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
