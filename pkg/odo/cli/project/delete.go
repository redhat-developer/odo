package project

import (
	"fmt"
	"strings"

	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/project"
	"github.com/spf13/cobra"

	ktemplates "k8s.io/kubernetes/pkg/kubectl/cmd/templates"
)

const deleteRecommendedCommandName = "delete"

var (
	deleteExample = ktemplates.Examples(`
	# Delete a project
	%[1]s myproject  
	`)

	deleteLongDesc = ktemplates.LongDesc(`Delete a project and all resources deployed in the project being deleted`)

	deleteShortDesc = `Delete a project`
)

// ProjectDeleteOptions encapsulates the options for the odo project delete command
type ProjectDeleteOptions struct {

	// name of the project
	projectName string

	// force delete doesn't ask the user for confirmation
	projectForceDeleteFlag bool

	// generic context options common to all commands
	*genericclioptions.Context
}

// NewProjectDeleteOptions creates a ProjectDeleteOptions instance
func NewProjectDeleteOptions() *ProjectDeleteOptions {
	return &ProjectDeleteOptions{}
}

// Complete completes ProjectDeleteOptions after they've been deleted
func (pdo *ProjectDeleteOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	pdo.projectName = args[0]
	pdo.Context = genericclioptions.NewContext(cmd)
	return
}

// Validate validates the parameters of the ProjectDeleteOptions
func (pdo *ProjectDeleteOptions) Validate() (err error) {
	// Validate existence of the project to be deleted
	isValidProject, err := project.Exists(pdo.Context.Client, pdo.projectName)
	if !isValidProject {
		return fmt.Errorf("The project %s does not exist. Please check the list of projects using `odo project list`", pdo.projectName)
	}
	return
}

// Run runs the project delete command
func (pdo *ProjectDeleteOptions) Run() (err error) {
	var confirmDeletion string
	if pdo.projectForceDeleteFlag {
		confirmDeletion = "y"
	} else {
		log.Askf("Are you sure you want to delete project %v? [y/N]: ", pdo.projectName)
		fmt.Scanln(&confirmDeletion)
	}

	if strings.ToLower(confirmDeletion) != "y" {
		return fmt.Errorf("Aborting deletion of project: %v", pdo.projectName)
	}

	currentProject, err := project.Delete(pdo.Context.Client, pdo.projectName)
	if err != nil {
		return err
	}

	log.Infof("Deleted project : %v", pdo.projectName)

	if currentProject != "" {
		log.Infof("%s has been set as the active project\n", currentProject)
	} else {
		// oc errors out as "error: you do not have rights to view project "$deleted_project"."
		log.Infof("You are not a member of any projects. You can request a project to be created using the `odo project create <project_name>` command")
	}

	return
}

// NewCmdProjectDelete creates the project delete command
func NewCmdProjectDelete(name, fullName string) *cobra.Command {
	o := NewProjectDeleteOptions()

	projectDeleteCmd := &cobra.Command{
		Use:     name,
		Short:   deleteShortDesc,
		Long:    deleteLongDesc,
		Example: fmt.Sprintf(deleteExample, fullName),
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			util.LogErrorAndExit(o.Complete(name, cmd, args), "")
			util.LogErrorAndExit(o.Validate(), "")
			util.LogErrorAndExit(o.Run(), "")
		},
	}

	projectDeleteCmd.Flags().BoolVarP(&o.projectForceDeleteFlag, "force", "f", false, "Delete project without prompting")

	return projectDeleteCmd
}
