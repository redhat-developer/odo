package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/redhat-developer/ocdev/pkg/component"
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

func init() {
	componentCreateCmd.Flags().StringVar(&componentBinary, "binary", "", "Binary artifact")
	componentCreateCmd.Flags().StringVar(&componentGit, "git", "", "Git source")
	componentCreateCmd.Flags().StringVar(&componentLocal, "local", "", "Use local directory as a source for component")

	rootCmd.AddCommand(componentCreateCmd)
}
