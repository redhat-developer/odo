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

	"github.com/redhat-developer/ocdev/pkg/application"
	"github.com/spf13/cobra"
)

const defaultApplication = "app"

// applicationCmd represents the app command
var applicationCmd = &cobra.Command{
	Use:     "application",
	Short:   "application",
	Aliases: []string{"app"},
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create an application",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var name string
		// Set default application name if not set
		if len(args) == 0 {
			name = defaultApplication
		} else {
			name = args[0]
		}
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
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Error: specify the application to delete")
			os.Exit(-1)
		}
		if len(args) > 1 {
			fmt.Println("Error: delete accepts only 1 argument")
			os.Exit(-1)
		}
		err := application.Delete(args[0])
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
		fmt.Printf("Deleting application: %v\n", args[0])
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "lists all the applications",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		app, err := application.List()
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
		fmt.Print(app)
	},
}

func init() {
	getCmd.Flags().BoolVarP(&isQuiet, "short", "q", false, "If true, display only the application name")

	applicationCmd.AddCommand(listCmd)
	applicationCmd.AddCommand(deleteCmd)
	applicationCmd.AddCommand(getCmd)
	applicationCmd.AddCommand(createCmd)
	rootCmd.AddCommand(applicationCmd)
}
