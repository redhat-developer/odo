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
	"github.com/redhat-developer/ocdev/pkg/application"
	"github.com/spf13/cobra"
	"os"
)

// applicationCmd represents the app command
var applicationCmd = &cobra.Command{
	Use:     "application",
	Short:   "application",
	Aliases: []string{"app"},
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create an application",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := application.Create(args[0]); err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
	},
}

func init() {
	applicationCmd.AddCommand(createCmd)
	rootCmd.AddCommand(applicationCmd)
}
