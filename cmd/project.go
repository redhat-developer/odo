package cmd

import (
	"fmt"
	"github.com/redhat-developer/odo/pkg/odo/util"
	"os"
	"strings"

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
		client := util.GetOcClient()
		current := project.GetCurrent(client)

		exists, err := project.Exists(client, projectName)
		util.CheckError(err, "")
		if !exists {
			fmt.Printf("The project %s does not exist\n", projectName)
			os.Exit(1)
		}

		err = project.SetCurrent(client, projectName)
		util.CheckError(err, "")
		if projectShortFlag {
			fmt.Print(projectName)
		} else {
			if current == projectName {
				fmt.Printf("Already on project : %v\n", projectName)
			} else {
				fmt.Printf("Switched to project : %v\n", projectName)
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
		client := util.GetOcClient()
		project := project.GetCurrent(client)

		if projectShortFlag {
			fmt.Print(project)
		} else {
			fmt.Printf("The current project is: %v\n", project)
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
		client := util.GetOcClient()
		err := project.Create(client, projectName)
		util.CheckError(err, "")
		err = project.SetCurrent(client, projectName)
		util.CheckError(err, "")
		fmt.Printf("New project created and now using project : %v\n", projectName)
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
		client := util.GetOcClient()

		// Validate existence of the project to be deleted
		isValidProject, err := project.Exists(client, projectName)
		util.CheckError(err, "Failed to delete project %s", projectName)
		if !isValidProject {
			fmt.Printf("The project %s does not exist. Please check the list of projects using `odo project list`\n", projectName)
			os.Exit(1)
		}

		var confirmDeletion string
		if projectForceDeleteFlag {
			confirmDeletion = "y"
		} else {
			fmt.Printf("Are you sure you want to delete project %v? [y/N] ", projectName)
			fmt.Scanln(&confirmDeletion)
		}

		if strings.ToLower(confirmDeletion) != "y" {
			fmt.Printf("Aborting deletion of project: %v\n", projectName)
			os.Exit(1)
		}

		fmt.Printf("Deleting project %s...\n(this operation may take some time)\n", projectName)
		err = project.Delete(client, projectName)
		if err != nil {
			util.CheckError(err, "")
		}
		fmt.Printf("Deleted project : %v\n", projectName)

		// Get Current Project
		currProject := project.GetCurrent(client)

		// Check if List returns empty, if so, the currProject is showing old currentProject
		// In openshift, when the project is deleted, it does not reset the current project in kube config file which is used by odo for current project
		projects, err := project.List(client)
		util.CheckError(err, "")
		if len(projects) != 0 {
			fmt.Printf("%s has been set as the active project\n", currProject)
		} else {
			// oc errors out as "error: you do not have rights to view project "$deleted_project"."
			fmt.Printf("You are not a member of any projects. You can request a project to be created using the `odo project create <project_name>` command\n")
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
		client := util.GetOcClient()
		projects, err := project.List(client)
		util.CheckError(err, "")
		if len(projects) == 0 {
			fmt.Println("You are not a member of any projects. You can request a project to be created using the `odo project create <project_name>` command")
			return
		}
		fmt.Printf("ACTIVE   NAME\n")
		for _, app := range projects {
			activeMark := " "
			if app.Active {
				activeMark = "*"
			}
			fmt.Printf("  %s      %s\n", activeMark, app.Name)
		}
	},
}

func init() {

	projectGetCmd.Flags().BoolVarP(&projectShortFlag, "short", "q", false, "If true, display only the project name")
	projectSetCmd.Flags().BoolVarP(&projectShortFlag, "short", "q", false, "If true, display only the project name")
	projectDeleteCmd.Flags().BoolVarP(&projectForceDeleteFlag, "force", "f", false, "Delete project without prompting")

	projectCmd.Flags().AddFlagSet(projectGetCmd.Flags())
	projectCmd.AddCommand(projectGetCmd)
	projectCmd.AddCommand(projectSetCmd)
	projectCmd.AddCommand(projectCreateCmd)
	projectCmd.AddCommand(projectDeleteCmd)
	projectCmd.AddCommand(projectListCmd)

	// Add a defined annotation in order to appear in the help menu
	projectCmd.Annotations = map[string]string{"command": "other"}
	projectCmd.SetUsageTemplate(cmdUsageTemplate)

	rootCmd.AddCommand(projectCmd)
}
