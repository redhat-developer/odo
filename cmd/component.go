package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/redhat-developer/ocdev/pkg/application"
	"github.com/redhat-developer/ocdev/pkg/component"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	componentApp             string
	componentBinary          string
	componentGit             string
	componentLocal           string
	componentShortFlag       bool
	componentForceDeleteFlag bool
)

// componentCmd represents the component command
var componentCmd = &cobra.Command{
	Use:   "component",
	Short: "Components of application.",
	// 'ocdev component' is the same as 'ocdev component get'
	Run: componentGetCmd.Run,
}

var componentCreateCmd = &cobra.Command{
	Use:   "create <component_type> [component_name] [flags]",
	Short: "Create new component",
	Long: `Create new component.
If component name is not provided, component type value will be used for name.
	`,
	Example: `  # Create new nodejs component with the source in current directory. 
  ocdev create nodejs

  # Create new nodejs component named 'frontend' with the source in './frontend' directory
  ocdev create nodejs frontend --local ./frontend

  # Create new nodejs component with source from remote git repository.
  ocdev create nodejs --git https://github.com/openshift/nodejs-ex.git
	`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			// TODO: improve this message. It should say something about requiring component name
			return fmt.Errorf("At least one argument is required")
		}
		if len(args) > 2 {
			// TODO: Improve this message
			return fmt.Errorf("Invalid arguments, maximum 2 arguments possible")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("Component create called with args: %#v, flags: binary=%s, git=%s, local=%s", strings.Join(args, " "), componentBinary, componentGit, componentLocal)

		client := getOcClient()
		if len(componentBinary) != 0 {
			fmt.Printf("--binary is not implemented yet\n\n")
			cmd.Help()
			os.Exit(1)
		}

		//TODO: check flags - only one of binary, git, dir can be specified

		//We don't have to check it anymore, Args check made sure that args has at least one item
		// and no more than two
		componentType := args[0]
		componentName := args[0]
		if len(args) == 2 {
			componentName = args[1]
		}

		if len(componentBinary) != 0 {
			fmt.Printf("--binary is not implemented yet\n\n")
			os.Exit(1)
		}

		if len(componentGit) != 0 {
			output, err := component.CreateFromGit(client, componentName, componentType, componentGit)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Println(output)
		} else if len(componentLocal) != 0 {
			// we want to use and save absolute path for component
			dir, err := filepath.Abs(componentLocal)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			output, err := component.CreateFromDir(client, componentName, componentType, dir)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Println(output)
		} else {
			// we want to use and save absolute path for component
			dir, err := filepath.Abs("./")
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			output, err := component.CreateFromDir(client, componentName, componentType, dir)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Println(output)
		}
		// after component is successfully created, set is as active
		if err := component.SetCurrent(client, componentName); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

	},
}

var componentDeleteCmd = &cobra.Command{
	Use:   "delete <component_name>",
	Short: "Delete existing component",
	Long:  "Delete existing component.",
	Example: `  # Delete component named 'frontend'. 
  ocdev delete frontend
	`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("Please specify component name to delete.")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("component delete called")
		log.Debugf("args: %#v", strings.Join(args, " "))
		client := getOcClient()
		componentName := args[0]
		var confirmDeletion string

		currentApp, err := application.GetCurrent(client)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if componentForceDeleteFlag {
			confirmDeletion = "y"
		} else {
			fmt.Printf("Are you sure you want to delete %v from %v? [y/N] ", componentName, currentApp)
			fmt.Scanln(&confirmDeletion)
		}

		if strings.ToLower(confirmDeletion) == "y" {
			output, err := component.Delete(client, componentName)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Println(output)
		} else {
			fmt.Printf("Aborting deletion of component: %v\n", componentName)
		}
	},
}

var componentGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get currently active component",
	Long:  "Get currently active component.",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("component get called")
		client := getOcClient()
		component, err := component.GetCurrent(client)
		if err != nil {
			fmt.Println(errors.Wrap(err, "unable to get current component"))
			os.Exit(1)
		}
		if componentShortFlag {
			fmt.Print(component)
		} else {
			if component == "" {
				fmt.Printf("No component is set as current\n")
				return
			}
			fmt.Printf("The current component is: %v\n", component)
		}
	},
}

var componentSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set active component.",
	Long:  "Set component as active.",
	Example: `  # Set component named 'frontend' as active
  ocdev set component frontend
  `,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("Please provide component name")
		}
		if len(args) > 1 {
			return fmt.Errorf("Only one argument (component name) is allowed")
		}
		return nil
	}, Run: func(cmd *cobra.Command, args []string) {
		client := getOcClient()
		err := component.SetCurrent(client, args[0])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Printf("Switched to component: %v\n", args[0])
	},
}

var componentListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all component in current application",
	Long:  "List all component in current application.",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		client := getOcClient()

		components, err := component.List(client)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if len(components) == 0 {
			fmt.Println("There are no components deployed.")
			return
		}

		fmt.Println("You have deployed:")
		for _, comp := range components {
			fmt.Printf("%s using the %s component\n", comp.Name, comp.Type)
		}

	},
}

func init() {
	componentDeleteCmd.Flags().BoolVarP(&componentForceDeleteFlag, "force", "f", false, "Delete component without prompting")

	componentCreateCmd.Flags().StringVar(&componentBinary, "binary", "", "binary artifact")
	componentCreateCmd.Flags().StringVar(&componentGit, "git", "", "git source")
	componentCreateCmd.Flags().StringVar(&componentLocal, "local", "", "Use local directory as a source for component.")

	componentGetCmd.Flags().BoolVarP(&componentShortFlag, "short", "q", false, "If true, display only the component name")

	// add flags from 'get' to component command
	componentCmd.Flags().AddFlagSet(applicationGetCmd.Flags())

	componentCmd.AddCommand(componentDeleteCmd)
	componentCmd.AddCommand(componentGetCmd)
	componentCmd.AddCommand(componentCreateCmd)
	componentCmd.AddCommand(componentSetCmd)
	componentCmd.AddCommand(componentListCmd)

	rootCmd.AddCommand(componentCmd)
}
