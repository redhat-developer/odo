package storage

import (
	"encoding/json"
	"fmt"

	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/occlient"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"os"
	"text/tabwriter"

	appCmd "github.com/redhat-developer/odo/pkg/odo/cli/application"
	componentCmd "github.com/redhat-developer/odo/pkg/odo/cli/component"
	projectCmd "github.com/redhat-developer/odo/pkg/odo/cli/project"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/storage"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubernetes/pkg/kubectl/cmd/templates"
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
	storageListAllFlag bool
	componentName      string
	outputFlag         string
	*genericclioptions.Context
}

// NewStorageListOptions creates a new UrlCreateOptions instance
func NewStorageListOptions() *StorageListOptions {
	return &StorageListOptions{}
}

// Complete completes StorageListOptions after they've been Created
func (o *StorageListOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	o.Context = genericclioptions.NewContext(cmd)
	if o.storageListAllFlag {
		if len(genericclioptions.FlagValueIfSet(cmd, genericclioptions.ComponentFlagName)) > 0 {
			return fmt.Errorf("invalid arguments. Component name is not needed")
		}
	} else {
		o.componentName = o.Component()
	}
	return
}

// Validate validates the StorageListOptions based on completed values
func (o *StorageListOptions) Validate() (err error) {
	return
}

// Run contains the logic for the odo storage list command
func (o *StorageListOptions) Run() (err error) {
	if o.outputFlag == "json" {
		var storeList []storage.Storage
		if storageAllListflag {
			componentList, err := component.List(o.Client, o.Application)
			odoutil.LogErrorAndExit(err, "could not get component list")
			for _, component := range componentList {
				mountedStorages, err := storage.ListMounted(o.Client, component.ComponentName, o.Application)
				odoutil.LogErrorAndExit(err, "")
				for _, storage := range mountedStorages {
					mounted := getMachineReadableFormat(true, storage)
					storeList = append(storeList, mounted)
				}
			}

		} else {
			componentName := o.Component()
			mountedStorages, err := storage.ListMounted(o.Client, componentName, o.Application)
			odoutil.LogErrorAndExit(err, "")
			for _, storage := range mountedStorages {
				mounted := getMachineReadableFormat(true, storage)
				storeList = append(storeList, mounted)

			}
		}
		unmountedStorages, err := storage.ListUnmounted(o.Client, o.Application)
		odoutil.LogErrorAndExit(err, "")
		for _, storage := range unmountedStorages {
			unmounted := getMachineReadableFormat(false, storage)
			storeList = append(storeList, unmounted)
		}
		storageList := storage.StorageList{
			TypeMeta: metav1.TypeMeta{
				Kind:       "List",
				APIVersion: "odo.openshift.io/v1aplha1",
			},
			ListMeta: metav1.ListMeta{},
			Items:    storeList,
		}
		out, err := json.Marshal(storageList)
		odoutil.LogErrorAndExit(err, "")
		fmt.Println(string(out))
	} else {

		if storageAllListflag {
			printMountedStorageInAllComponent(o.Client, o.Application)
		} else {
			// storageComponent is the input component name
			componentName := o.Component()
			printMountedStorageInComponent(o.Client, componentName, o.Application)
		}
		printUnmountedStorage(o.Client, o.Application)
	}
	return
}

// getMachineReadableFormat returns resource information in machine readable format
func getMachineReadableFormat(mounted bool, stor storage.StorageInfo) storage.Storage {
	return storage.Storage{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Storage",
			APIVersion: "odo.openshift.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: stor.Name,
		},
		Spec: storage.StorageSpec{
			Size: stor.Size,
			Path: stor.Path,
		},
		Status: storage.StorageStatus{
			Mounted: mounted,
		},
	}

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
			odoutil.LogErrorAndExit(o.Complete(name, cmd, args), "")
			odoutil.LogErrorAndExit(o.Validate(), "")
			odoutil.LogErrorAndExit(o.Run(), "")
		},
	}

	storageListCmd.Flags().BoolVarP(&o.storageListAllFlag, "all", "a", false, "List all storage in all components")
	storageListCmd.Flags().StringVarP(&o.outputFlag, "output", "o", "", "output in json format")

	projectCmd.AddProjectFlag(storageListCmd)
	appCmd.AddApplicationFlag(storageListCmd)
	componentCmd.AddComponentFlag(storageListCmd)

	return storageListCmd
}

// printMountedStorageInComponent prints all the mounted storage in a given component of the application
func printMountedStorageInComponent(client *occlient.Client, componentName string, applicationName string) {

	// defining the column structure of the table
	tabWriterMounted := tabwriter.NewWriter(os.Stdout, 5, 2, 3, ' ', tabwriter.TabIndent)

	// create headers of mounted storage table
	fmt.Fprintln(tabWriterMounted, "NAME", "\t", "SIZE", "\t", "PATH")

	storageListMounted, err := storage.ListMounted(client, componentName, applicationName)
	odoutil.LogErrorAndExit(err, "could not get mounted storage list")

	// iterating over all mounted storage and put in the mount storage table
	if len(storageListMounted) > 0 {
		for _, mStorage := range storageListMounted {
			fmt.Fprintln(tabWriterMounted, mStorage.Name, "\t", mStorage.Size, "\t", mStorage.Path)
		}

		// print all mounted storage of the given component
		log.Infof("The component '%v' has the following storage attached:", componentName)
		tabWriterMounted.Flush()
	} else {
		log.Infof("The component '%v' has no storage attached", componentName)
	}
	fmt.Println("")
}

// printMountedStorageInAllComponent prints all the mounted storage in all the components of the application and project
func printMountedStorageInAllComponent(client *occlient.Client, applicationName string) {
	componentList, err := component.List(client, applicationName)
	odoutil.LogErrorAndExit(err, "could not get component list")

	// iterating over all the components in the given aplication and project
	for _, component := range componentList {
		printMountedStorageInComponent(client, component.ComponentName, applicationName)
	}
}

// printUnmountedStorage prints all the unmounted storage in the application
func printUnmountedStorage(client *occlient.Client, applicationName string) {

	// defining the column structure of the unmounted storage table
	tabWriterUnmounted := tabwriter.NewWriter(os.Stdout, 5, 2, 3, ' ', tabwriter.TabIndent)

	// create header of unmounted storage in all the components of the given application and project
	fmt.Fprintln(tabWriterUnmounted, "NAME", "\t", "SIZE")

	storageListUnmounted, err := storage.ListUnmounted(client, applicationName)
	odoutil.LogErrorAndExit(err, "could not get unmounted storage list")

	// iterating over all unmounted storage and put in the unmount storage table
	if len(storageListUnmounted) > 0 {
		for _, uStorage := range storageListUnmounted {
			fmt.Fprintln(tabWriterUnmounted, uStorage.Name, "\t", uStorage.Size)
		}

		// print unmounted storage of all the application
		log.Info("Storage that are not mounted to any component:")
		tabWriterUnmounted.Flush()
	}
	fmt.Println("")
}
