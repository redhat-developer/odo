package project

import (
	"errors"
	"fmt"

	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/odo/cli/ui"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	odoutil "github.com/openshift/odo/pkg/odo/util"
	"github.com/openshift/odo/pkg/project"
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

// Complete completes ProjectDeleteOptions after they've been created
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
		errorMsg := fmt.Sprintf("The project %s does not exist. Please check the list of projects using `odo project list`", pdo.projectName)

		// If json has been selected
		if pdo.OutputFlag == "json" {
			odoutil.ErrorMachineOutput(pdo.projectName, errors.New(errorMsg))
		}

		return fmt.Errorf(errorMsg)
	}
	return
}

// Run runs the project delete command
func (pdo *ProjectDeleteOptions) Run() (err error) {

	// Set the current project
	err = project.SetCurrent(pdo.Context.Client, pdo.projectName)
	if err != nil {
		return
	}

	// If machine readable output is selected, we will skip printing the project information
	// as well as the force flag.

	if pdo.OutputFlag == "json" {

		// Delete the project without confirmation
		err := project.Delete(pdo.Context.Client, pdo.projectName)
		if err != nil {
			// Error out with Json...
			odoutil.ErrorMachineOutput(pdo.projectName, err)
		}

		// Create "machine-readable" output
		odoutil.SuccessMachineOutput(pdo.projectName,
			fmt.Sprintf("Deleted project %s", pdo.projectName),
			"Project")

		return nil

	} else {

		// Print the artifacts that will be deleted as the result of project deletion
		err = printDeleteProjectInfo(pdo.Context.Client, pdo.projectName)
		if err != nil {
			return err
		}

		// Check to see if the "force" flag is being used.
		if pdo.projectForceDeleteFlag || ui.Proceed(fmt.Sprintf("Are you sure you want to delete project %v", pdo.projectName)) {
			err := project.Delete(pdo.Context.Client, pdo.projectName)
			if err != nil {
				return err
			}

			log.Infof("Deleted project : %v", pdo.projectName)
			return nil
		}

	}

	return fmt.Errorf("Aborting deletion of project: %v", pdo.projectName)
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
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	genericclioptions.AddOutputFlag(projectDeleteCmd)

	projectDeleteCmd.Flags().BoolVarP(&o.projectForceDeleteFlag, "force", "f", false, "Delete project without prompting")

	return projectDeleteCmd
}
