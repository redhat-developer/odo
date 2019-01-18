package newproject

import (
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/project"
	"github.com/spf13/cobra"
)

var newProjectCmd = &cobra.Command{
	Use:   "new-project",
	Short: "none",
	Long:  "none",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]
		client := genericclioptions.Client(cmd)
		err := project.Create(client, projectName)
		odoutil.LogErrorAndExit(err, "")
		err = project.SetCurrent(client, projectName)
		odoutil.LogErrorAndExit(err, "")
		log.Successf("New project created and now using project : %v", projectName)
	},
	Hidden: true,
}

// NewCmdNewProject Creates an alias for odo new-project
// Ref https://github.com/redhat-developer/odo/issues/1017
// This is a tempoary fix to get around the fact that we can intercept or manipulate oc logic
// message carried forward from oc login immplementation
func NewCmdNewProject() *cobra.Command {
	return newProjectCmd
}
