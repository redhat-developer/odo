package storage

import (
	"fmt"

	odoutil "github.com/openshift/odo/pkg/odo/util"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

// RecommendedCommandName is the recommended command name
const RecommendedCommandName = "storage"

var (
	storageShortDesc = `Perform storage operations`
	storageLongDesc  = ktemplates.LongDesc(`Perform storage operations`)
)

// NewCmdStorage implements the odo storage command
func NewCmdStorage(name, fullName string) *cobra.Command {
	storageCreateCmd := NewCmdStorageCreate(createRecommendedCommandName, odoutil.GetFullName(fullName, createRecommendedCommandName))
	storageDeleteCmd := NewCmdStorageDelete(deleteRecommendedCommandName, odoutil.GetFullName(fullName, deleteRecommendedCommandName))
	storageListCmd := NewCmdStorageList(listRecommendedCommandName, odoutil.GetFullName(fullName, listRecommendedCommandName))

	var storageCmd = &cobra.Command{
		Use:   name,
		Short: storageShortDesc,
		Long:  storageLongDesc,
		Example: fmt.Sprintf("%s\n\n%s\n\n%s",
			storageCreateCmd.Example,
			storageDeleteCmd.Example,
			storageListCmd.Example),
	}

	storageCmd.AddCommand(storageCreateCmd)
	storageCmd.AddCommand(storageDeleteCmd)
	storageCmd.AddCommand(storageListCmd)

	// Add a defined annotation in order to appear in the help menu
	storageCmd.Annotations = map[string]string{"command": "main"}
	storageCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)

	return storageCmd
}
