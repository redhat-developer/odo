package cmd

import (
	"fmt"
	"os"

	"strings"

	"github.com/redhat-developer/ocdev/pkg/application"
	"github.com/spf13/cobra"
)

var (
	applicationShortFlag       bool
	applicationForceDeleteFlag bool
)

// applicationCmd represents the app command
var applicationCmd = &cobra.Command{
	Use:     "application",
	Short:   "Perform application operations",
	Aliases: []string{"app"},
	// 'ocdev application' is the same as 'ocdev application get'
	Run: applicationGetCmd.Run,
}

var applicationCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "create an application",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Args validation makes sure that there is exactly one argument
		name := args[0]
		client := getOcClient()
		fmt.Printf("Creating application: %v\n", name)
		err := application.Create(client, name)
		checkError(err, "")
		err = application.SetCurrent(client, name)
		checkError(err, "")
		fmt.Printf("Switched to application: %v\n", name)
	},
}

var applicationGetCmd = &cobra.Command{
	Use:   "get",
	Short: "get the active application",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		client := getOcClient()
		app, err := application.GetCurrent(client)
		checkError(err, "")
		if applicationShortFlag {
			fmt.Print(app)
			return
		}
		if app == "" {
			fmt.Printf("There's no active application.\nYou can create one by running 'ocdev application create <name>'.")
			return
		}
		fmt.Printf("The current application is: %v\n", app)
	},
}

var applicationDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete the given application",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("Please provide application name")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		client := getOcClient()
		appName := args[0]
		var confirmDeletion string

		if applicationForceDeleteFlag {
			confirmDeletion = "y"
		} else {
			fmt.Printf("Are you sure you want to delete the application: %v? [y/N] ", appName)
			fmt.Scanln(&confirmDeletion)
		}

		if strings.ToLower(confirmDeletion) == "y" {
			err := application.Delete(client, appName)
			checkError(err, "")
			fmt.Printf("Deleted application: %s\n", args[0])
		} else {
			fmt.Printf("Aborting deletion of application: %v\n", appName)
		}
	},
}

var applicationListCmd = &cobra.Command{
	Use:   "list",
	Short: "lists all the applications",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		client := getOcClient()
		apps, err := application.List(client)
		checkError(err, "")
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

var applicationSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set application as active.",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("Please provide application name")
		}
		if len(args) > 1 {
			return fmt.Errorf("Only one argument (application name) is allowed")
		}
		return nil
	}, Run: func(cmd *cobra.Command, args []string) {
		client := getOcClient()
		appName := args[0]
		// error if application does not exist
		exists, err := application.Exists(client, appName)
		checkError(err, "unable to check if application exists")
		if !exists {
			fmt.Printf("Application %v does not exist\n", appName)
			os.Exit(1)
		}

		err = application.SetCurrent(client, appName)
		checkError(err, "")
		fmt.Printf("Switched to application: %v\n", args[0])
	},
}

func init() {
	applicationDeleteCmd.Flags().BoolVarP(&applicationForceDeleteFlag, "force", "f", false, "Delete application without prompting")

	applicationGetCmd.Flags().BoolVarP(&applicationShortFlag, "short", "q", false, "If true, display only the application name")
	// add flags from 'get' to application command
	applicationCmd.Flags().AddFlagSet(applicationGetCmd.Flags())

	applicationCmd.AddCommand(applicationListCmd)
	applicationCmd.AddCommand(applicationDeleteCmd)
	applicationCmd.AddCommand(applicationGetCmd)
	applicationCmd.AddCommand(applicationCreateCmd)
	applicationCmd.AddCommand(applicationSetCmd)

	rootCmd.AddCommand(applicationCmd)
}
