package cmd

import (
	"fmt"
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
		client := getOcClient()
		current := project.GetCurrent(client)

		exists, err := project.Exists(client, projectName)
		checkError(err, "")
		if !exists {
			fmt.Printf("The project %s does not exist\n", projectName)
			os.Exit(1)
		}

		err = project.SetCurrent(client, projectName)
		checkError(err, "")
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
		client := getOcClient()
		project := project.GetCurrent(client)

		if projectShortFlag {
			fmt.Println(project)
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
		client := getOcClient()
		err := project.Create(client, projectName)
		checkError(err, "")
		err = project.SetCurrent(client, projectName)
		checkError(err, "")
		fmt.Printf("New project created and now using project : %v\n", projectName)
	},
}

var projectDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a project",
	Long:  "Delete a project and all resources deployed in the project being deleted",
	Example: `  # Create a new project
  odo project delete myproject
	`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]
		client := getOcClient()

		var confirmDeletion string
		if projectForceDeleteFlag {
			confirmDeletion = "y"
		} else {
			fmt.Printf("Are you sure you want to delete project %v? [y/N] ", projectName)
			fmt.Scanln(&confirmDeletion)
		}

		if strings.ToLower(confirmDeletion) != "y" {
			fmt.Printf("Aborting deletion of project: %v\n", projectName)
		}

		err := project.Delete(client, projectName)
		if err != nil {
			checkError(err, "")
		}
		fmt.Printf("Deleted project : %v\n", projectName)
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
		client := getOcClient()
		projects, err := project.List(client)
		checkError(err, "")
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
	projectGetCmd.Flags().BoolVarP(&projectShortFlag, "short", "q", false, "If true, display only the application name")
	projectSetCmd.Flags().BoolVarP(&projectShortFlag, "short", "q", false, "If true, display only the application name")
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
