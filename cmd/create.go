package cmd

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/redhat-developer/odo/pkg/component"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	componentBinary string
	componentGit    string
	componentLocal  string
)

var componentCreateCmd = &cobra.Command{
	Use:   "create <component_type> [component_name] [flags]",
	Short: "Create new component",
	Long: `Create new component to deploy on OpenShift.

If component name is not provided, component type value will be used for the name.

A full list of component types that can be deployed is available using: 'odo component list'`,
	Example: `  # Create new Node.js component with the source in current directory. 
  odo create nodejs

  # Create new Node.js component named 'frontend' with the source in './frontend' directory
  odo create nodejs frontend --local ./frontend

  # Create new Node.js component with source from remote git repository.
  odo create nodejs --git https://github.com/openshift/nodejs-ex.git

  # Create a Ruby component
  odo create ruby
	
  # Create a Python component
  odo create python`,
	Args: cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("Component create called with args: %#v, flags: binary=%s, git=%s, local=%s", strings.Join(args, " "), componentBinary, componentGit, componentLocal)

		client := getOcClient()
		if len(componentBinary) != 0 {
			log.Error("--binary is not implemented yet\n\n")
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
			log.Error("--binary is not implemented yet\n\n")
			os.Exit(1)
		}

		exists, err := component.Exists(client, componentName)
		if err != nil {
			checkError(err, "")
		}
		if exists {
			log.Errorf("Component with the name %s already exists in the current application\n", componentName)
			os.Exit(1)
		}

		if len(componentGit) != 0 {
			err := component.CreateFromGit(client, componentName, componentType, componentGit)
			checkError(err, "")
		} else if len(componentLocal) != 0 {
			// we want to use and save absolute path for component
			dir, err := filepath.Abs(componentLocal)
			checkError(err, "")
			err = component.CreateFromDir(client, componentName, componentType, dir)
			checkError(err, "")
		} else {
			// we want to use and save absolute path for component
			dir, err := filepath.Abs("./")
			checkError(err, "")
			err = component.CreateFromDir(client, componentName, componentType, dir)
			checkError(err, "")
		}
		// after component is successfully created, set is as active
		err = component.SetCurrent(client, componentName)
		checkError(err, "")
	},
}

func init() {
	componentCreateCmd.Flags().StringVar(&componentBinary, "binary", "", "Binary artifact")
	componentCreateCmd.Flags().StringVar(&componentGit, "git", "", "Git source")
	componentCreateCmd.Flags().StringVar(&componentLocal, "local", "", "Use local directory as a source for component")

	rootCmd.AddCommand(componentCreateCmd)
}
