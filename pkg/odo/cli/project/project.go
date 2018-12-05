package project

import (
	"fmt"
	"os"
	"strings"

	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/odo/util/completion"

	"github.com/redhat-developer/odo/pkg/project"
	"github.com/spf13/cobra"
)

var (
	projectShortFlag       bool
	projectForceDeleteFlag bool
)

var projectCmd = &cobra.Command{
	Use:   "project [options]",
	Short: "Perform project operations",
	Long:  "Perform project operations",
	Example: fmt.Sprintf("%s\n%s\n%s\n%s\n%s",
		projectSetCmd.Example,
		projectCreateCmd.Example,
		projectListCmd.Example,
		projectDeleteCmd.Example,
		projectGetCmd.Example),
	// 'odo project' is the same as 'odo project get'
	// 'odo project <project_name>' is the same as 'odo project set <project_name>'
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 && args[0] != "get" && args[0] != "set" {
			projectSetCmd.Run(cmd, args)
		} else {
			projectGetCmd.Run(cmd, args)
		}
	},
}

var projectSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set the current active project",
	Long:  "Set the current active project",
	Example: `  # Set the current active project
  odo project set myproject
	`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]
		context := genericclioptions.NewContext(cmd)
		client := context.Client
		current := context.Project

		exists, err := project.Exists(client, projectName)
		odoutil.CheckError(err, "")
		if !exists {
			log.Errorf("The project %s does not exist", projectName)
			os.Exit(1)
		}

		err = project.SetCurrent(client, projectName)
		odoutil.CheckError(err, "")
		if projectShortFlag {
			fmt.Print(projectName)
		} else {
			if current == projectName {
				log.Infof("Already on project : %v", projectName)
			} else {
				log.Infof("Switched to project : %v", projectName)
			}
		}
	},
}

var projectGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get the active project",
	Long:  "Get the active project",
	Example: `  # Get the active project
  odo project get
	`,
	Args: cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		context := genericclioptions.NewContext(cmd)
		project := context.Project

		if projectShortFlag {
			fmt.Print(project)
		} else {
			log.Infof("The current project is: %v", project)
		}
	},
}

var projectCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new project",
	Long:  "Create a new project",
	Example: `  # Create a new project
  odo project create myproject
	`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]
		client := genericclioptions.Client(cmd)
		err := project.Create(client, projectName)
		odoutil.CheckError(err, "")
		err = project.SetCurrent(client, projectName)
		odoutil.CheckError(err, "")
		log.Successf("New project created and now using project : %v", projectName)
	},
}

var projectDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a project",
	Long:  "Delete a project and all resources deployed in the project being deleted",
	Example: `  # Delete a project
  odo project delete myproject
	`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]
		client := genericclioptions.Client(cmd)

		// Validate existence of the project to be deleted
		isValidProject, err := project.Exists(client, projectName)
		odoutil.CheckError(err, "Failed to delete project %s", projectName)
		if !isValidProject {
			log.Errorf("The project %s does not exist. Please check the list of projects using `odo project list`", projectName)
			os.Exit(1)
		}

		var confirmDeletion string
		if projectForceDeleteFlag {
			confirmDeletion = "y"
		} else {
			log.Askf("Are you sure you want to delete project %v? [y/N]: ", projectName)
			fmt.Scanln(&confirmDeletion)
		}

		if strings.ToLower(confirmDeletion) != "y" {
			log.Errorf("Aborting deletion of project: %v", projectName)
			os.Exit(1)
		}

		currentProject, err := project.Delete(client, projectName)
		if err != nil {
			odoutil.CheckError(err, "")
		}

		fmt.Printf("Deleted project : %v\n", projectName)

		if currentProject != "" {
			log.Infof("%s has been set as the active project\n", currentProject)
		} else {
			// oc errors out as "error: you do not have rights to view project "$deleted_project"."
			log.Infof("You are not a member of any projects. You can request a project to be created using the `odo project create <project_name>` command")
		}

	},
}

var projectListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all the projects",
	Long:  "List all the projects",
	Example: `  # List all the projects
  odo project list
	`,
	Args: cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		client := genericclioptions.Client(cmd)
		projects, err := project.List(client)
		odoutil.CheckError(err, "")
		if len(projects) == 0 {
			log.Errorf("You are not a member of any projects. You can request a project to be created using the `odo project create <project_name>` command")
			os.Exit(1)
		}
		fmt.Printf("ACTIVE   NAME\n")
		for _, project := range projects {
			activeMark := " "
			if project.Active {
				activeMark = "*"
			}
			fmt.Printf("  %s      %s\n", activeMark, project.Name)
		}
	},
}

// NewCmdProject implements the project odo command
func NewCmdProject() *cobra.Command {

	projectGetCmd.Flags().BoolVarP(&projectShortFlag, "short", "q", false, "If true, display only the project name")
	projectSetCmd.Flags().BoolVarP(&projectShortFlag, "short", "q", false, "If true, display only the project name")
	projectDeleteCmd.Flags().BoolVarP(&projectForceDeleteFlag, "force", "f", false, "Delete project without prompting")
	projectDeleteCmd.Flags().BoolVarP(&projectShortFlag, "short", "q", false, "Delete project without prompting")

	projectCmd.Flags().AddFlagSet(projectGetCmd.Flags())
	projectCmd.AddCommand(projectGetCmd)
	projectCmd.AddCommand(projectSetCmd)
	projectCmd.AddCommand(projectCreateCmd)
	projectCmd.AddCommand(projectDeleteCmd)
	projectCmd.AddCommand(projectListCmd)

	// Add a defined annotation in order to appear in the help menu
	projectCmd.Annotations = map[string]string{"command": "other"}
	projectCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)

	completion.RegisterCommandHandler(projectSetCmd, completion.ProjectNameCompletionHandler)
	completion.RegisterCommandHandler(projectDeleteCmd, completion.ProjectNameCompletionHandler)

	return projectCmd
}

// AddProjectFlag adds a `project` flag to the given cobra command
// Also adds a completion handler to the flag
func AddProjectFlag(cmd *cobra.Command) {
	cmd.Flags().String(genericclioptions.ProjectFlagName, "", "Project, defaults to active project")
	completion.RegisterCommandFlagHandler(cmd, "project", completion.ProjectNameCompletionHandler)
}
