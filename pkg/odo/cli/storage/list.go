package storage

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/openshift/odo/pkg/localConfigProvider"
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
	# List all storage attached to the current component
  %[1]s
	`)
)

type ListOptions struct {
	componentContext string
	*genericclioptions.Context

	client storage.Client
}

// NewStorageListOptions creates a new ListOptions instance
func NewStorageListOptions() *ListOptions {
	return &ListOptions{}
}

// Complete completes ListOptions after they've been created
func (o *ListOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	o.Context, err = genericclioptions.New(genericclioptions.CreateParameters{
		Cmd:              cmd,
		Devfile:          true,
		ComponentContext: o.componentContext,
	})

	if err != nil {
		return err
	}

	o.client = storage.NewClient(storage.ClientOptions{
		LocalConfigProvider: o.Context.LocalConfigProvider,
		OCClient:            *o.Context.Client,
	})

	return
}

// Validate validates the ListOptions based on completed values
func (o *ListOptions) Validate() (err error) {
	return nil
}

func (o *ListOptions) Run(cmd *cobra.Command) (err error) {
	storageList, err := o.client.List()
	if err != nil {
		return err
	}

	if log.IsJSON() {
		machineoutput.OutputSuccess(storageList)
	} else {
		localContainers, err := o.Context.LocalConfigProvider.GetContainers()
		if err != nil {
			return err
		}
		if isContainerDisplay(storageList, localContainers) {
			printStorageWithContainer(storageList, o.Context.LocalConfigProvider.GetName())
		} else {
			printStorage(storageList, o.Context.LocalConfigProvider.GetName())
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
func isContainerDisplay(storageList storage.StorageList, components []localConfigProvider.LocalContainer) bool {

	// get all the container names
	componentsMap := make(map[string]bool)
	for _, comp := range components {
		componentsMap[comp.Name] = true
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
