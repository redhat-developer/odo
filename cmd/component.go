// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/redhat-developer/ocdev/pkg/component"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	componentBinary string
	componentGit    string
	componentDir    string
)

// componentCmd represents the component command
var componentCmd = &cobra.Command{
	Use:   "component",
	Short: "components of application",
	Long:  "components of application",
}

var componentCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "component create <component_type> [component_name]",
	Long:  "create new component",
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
		log.Debugf("component create called")
		log.Debugf("args: %#v", strings.Join(args, " "))
		log.Debugf("flags: binary=%s, git=%s, dir=%s", componentBinary, componentGit, componentDir)

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
			output, err := component.CreateFromGit(componentName, componentType, componentGit)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Println(output)
		} else if len(componentDir) != 0 {
			output, err := component.CreateFromDir(componentName, componentType, componentDir)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Println(output)
		} else {
			// no flag was set, create empty component
			output, err := component.CreateEmpty(componentName, componentType)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Println(output)
		}

	},
}

func init() {
	componentCreateCmd.Flags().StringVar(&componentBinary, "binary", "", "binary artifact")
	componentCreateCmd.Flags().StringVar(&componentGit, "git", "", "git source")
	componentCreateCmd.Flags().StringVar(&componentDir, "dir", "", "local directory as source")

	componentCmd.AddCommand(componentCreateCmd)
	rootCmd.AddCommand(componentCmd)
}
