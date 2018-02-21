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

	"github.com/spf13/cobra"
)

var (
	// VERSION  is version number that will be displayed when running ./ocdev version
	VERSION = "v0.0.1"

	// GITCOMMIT is hash of the commit that wil be displayed when running ./ocdev version
	// this will be overwritten when running  build like this: go build -ldflags="-X github.com/redhat-developer/ocdev/cmd.GITCOMMIT=$(GITCOMMIT)"
	// HEAD is default indicating that this was not set during build
	GITCOMMIT = "HEAD"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of ocdev",
	Run: func(cmd *cobra.Command, args []string) {

		// If verbose mode is enabled, dump all KUBECLT_* env variables
		// this is usefull for debuging oc plugin integration
		if GlobalVerbose {
			for _, v := range os.Environ() {
				if strings.HasPrefix(v, "KUBECTL_") {
					fmt.Println(v)
				}
			}
		}

		fmt.Println(VERSION + " (" + GITCOMMIT + ")")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
