package cmd

import (
	"fmt"
	"os"

	"github.com/redhat-developer/ocdev/pkg/project"
	"github.com/spf13/cobra"
)

var (
	projectShortFlag bool
)

var projectCmd = &cobra.Command{
	Use:   "project [options]",
	Short: "Perform project operations",
	Run:   projectGetCmd.Run,
}

var projectSetCmd = &cobra.Command{
	Use:   "set",
	Short: "set the current active project",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]
		client := getOcClient()
		current := project.GetCurrent(client)

		err := project.SetCurrent(client, projectName)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if projectShortFlag {
			fmt.Print(projectName)
		} else {
			if current == projectName {
				fmt.Printf("Already on project : %v\n", projectName)
			} else {
				fmt.Printf("Now using project : %v\n", projectName)
			}
		}
	},
}

var projectGetCmd = &cobra.Command{
	Use:   "get",
	Short: "get the active project",
	Args:  cobra.ExactArgs(0),
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
	Short: "create a new project",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]
		client := getOcClient()
		err := project.Create(client, projectName)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		err = project.SetCurrent(client, projectName)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Printf("New project created and now using project : %v\n", projectName)
	},
}

var projectListCmd = &cobra.Command{
	Use:   "list",
	Short: "list all the projects",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		client := getOcClient()
		projects, err := project.List(client)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
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
	projectGetCmd.Flags().BoolVarP(&projectShortFlag, "short", "q", false, "If true, display only the application name")
	projectSetCmd.Flags().BoolVarP(&projectShortFlag, "short", "q", false, "If true, display only the application name")
	projectCmd.Flags().AddFlagSet(projectGetCmd.Flags())
	projectCmd.AddCommand(projectGetCmd)
	projectCmd.AddCommand(projectSetCmd)
	projectCmd.AddCommand(projectCreateCmd)
	projectCmd.AddCommand(projectListCmd)
	rootCmd.AddCommand(projectCmd)
}
