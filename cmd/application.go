package cmd

import (
	"fmt"
	"os"

	"github.com/redhat-developer/ocdev/pkg/application"
	"github.com/spf13/cobra"
)

// applicationCmd represents the app command
var applicationCmd = &cobra.Command{
	Use:     "application",
	Short:   "application",
	Aliases: []string{"app"},
}

var applicationCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "create an application",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("Please provide name for the new application")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Args validation makes sure that there is exactly one argument
		name := args[0]

		fmt.Printf("Creating application: %v\n", name)
		if err := application.Create(name); err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
		fmt.Printf("Switched to application: %v\n", name)
	},
}

var isQuiet bool
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get the active application",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		app, err := application.GetCurrent()
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
		if isQuiet {
			fmt.Print(app)
		} else {
			fmt.Printf("The current application is: %v\n", app)
		}
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete the given application",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("Please provide application name")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		err := application.Delete(args[0])
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
		fmt.Printf("Application %s  was deleted.\n", args[0])
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "lists all the applications",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		apps, err := application.List()
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
		fmt.Printf("ACTIVE   NAME\n")
		for _, app := range apps {
			activeMark := " "
			if app.Active {
				activeMark = "*"
			}
			fmt.Printf("  %s      %s\n", activeMark, app.Name)
		}
	},
}

func init() {
	getCmd.Flags().BoolVarP(&isQuiet, "short", "q", false, "If true, display only the application name")

	applicationCmd.AddCommand(listCmd)
	applicationCmd.AddCommand(deleteCmd)
	applicationCmd.AddCommand(getCmd)
	applicationCmd.AddCommand(applicationCreateCmd)
	rootCmd.AddCommand(applicationCmd)
}
