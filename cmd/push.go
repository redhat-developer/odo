package cmd

import (
	"fmt"
	util2 "github.com/redhat-developer/odo/pkg/odo/util"
	"net/url"
	"os"
	"runtime"

	"github.com/fatih/color"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/util"

	"path/filepath"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push [component name]",
	Short: "Push source code to a component",
	Long:  `Push source code to a component.`,
	Example: `  # Push source code to the current component
  odo push

  # Push data to the current component from the original source.
  odo push

  # Push source code in ~/mycode to component called my-component
  odo push my-component --local ~/mycode
	`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		stdout := color.Output
		client := util2.GetOcClient()

		util2.GetAndSetNamespace(client)
		applicationName := util2.GetAppName(client)

		var inputName string
		if len(args) == 0 {
			inputName = ""
		} else {
			inputName = args[0]
		}
		componentName := util2.GetComponent(client, inputName, applicationName)
		fmt.Printf("Pushing changes to component: %v\n", componentName)

		sourceType, sourcePath, err := component.GetComponentSource(client, componentName, applicationName)
		util2.CheckError(err, "unable to get component source")
		switch sourceType {
		case "local", "binary":
			// use value of '--dir' as source if it was used
			if len(componentLocal) != 0 {
				if sourceType == "binary" {
					fmt.Printf("Unable to push local directory:%s to component %s that uses binary %s.\n", componentLocal, componentName, sourcePath)
					os.Exit(1)
				}
				sourcePath = util.GenFileUrl(componentLocal, runtime.GOOS)
			}

			u, err := url.Parse(sourcePath)
			util2.CheckError(err, fmt.Sprintf("unable to parse source %s from component %s", sourcePath, componentName))

			if u.Scheme != "" && u.Scheme != "file" {
				fmt.Printf("Component %s has invalid source path %s", componentName, u.Scheme)
				os.Exit(1)
			}

			localLocation := util.ReadFilePath(u, runtime.GOOS)

			_, err = os.Stat(localLocation)
			if err != nil {
				util2.CheckError(err, "")
			}

			if sourceType == "local" {
				glog.V(4).Infof("Copying directory %s to pod", localLocation)
				err = component.PushLocal(client, componentName, applicationName, localLocation, os.Stdout, []string{})
			} else {
				dir := filepath.Dir(localLocation)
				glog.V(4).Infof("Copying file %s to pod", localLocation)
				err = component.PushLocal(client, componentName, applicationName, dir, os.Stdout, []string{localLocation})
			}
			util2.CheckError(err, fmt.Sprintf("failed to push component: %v", componentName))

		case "git":
			// currently we don't support changing build type
			// it doesn't make sense to use --dir with git build
			if len(componentLocal) != 0 {
				fmt.Printf("Unable to push local directory:%s to component %s that uses Git repository:%s.\n", componentLocal, componentName, sourcePath)
				os.Exit(1)
			}
			err := component.Build(client, componentName, applicationName, true, true, stdout)
			util2.CheckError(err, fmt.Sprintf("failed to push component: %v", componentName))
		}

		fmt.Printf("changes successfully pushed to component: %v\n", componentName)
	},
}

func init() {
	pushCmd.Flags().StringVarP(&componentLocal, "local", "l", "", "Use given local directory as a source for component. (It must be a local component)")

	// Add a defined annotation in order to appear in the help menu
	pushCmd.Annotations = map[string]string{"command": "component"}
	pushCmd.SetUsageTemplate(cmdUsageTemplate)

	//Adding `--project` flag
	addProjectFlag(pushCmd)
	//Adding `--application` flag
	addApplicationFlag(pushCmd)

	rootCmd.AddCommand(pushCmd)
}
