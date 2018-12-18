package component

import (
	"fmt"

	appCmd "github.com/redhat-developer/odo/pkg/odo/cli/application"
	projectCmd "github.com/redhat-developer/odo/pkg/odo/cli/project"

	"net/url"
	"os"
	"runtime"

	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"

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

		context := genericclioptions.NewContext(cmd)
		client := context.Client
		applicationName := context.Application

		var argComponent string
		if len(args) == 1 {
			argComponent = args[0]
		}
		componentName := context.Component(argComponent)

		// if the componentName is blank then there is no active component set
		if len(componentName) == 0 {
			log.Error("No component is set as active. Use 'odo component set' to set an active component.")
			os.Exit(1)
		}

		log.Namef("Pushing changes to component: %v", componentName)

		sourceType, sourcePath, err := component.GetComponentSource(client, componentName, applicationName)
		odoutil.CheckError(err, "unable to get component source")
		switch sourceType {
		case "local", "binary":
			// use value of '--dir' as source if it was used
			if len(componentLocal) != 0 {
				if sourceType == "binary" {
					log.Errorf("Unable to push local directory:%s to component %s that uses binary %s.", componentLocal, componentName, sourcePath)
					os.Exit(1)
				}
				sourcePath = util.GenFileURL(componentLocal, runtime.GOOS)
			}

			u, err := url.Parse(sourcePath)
			odoutil.CheckError(err, fmt.Sprintf("unable to parse source %s from component %s", sourcePath, componentName))

			if u.Scheme != "" && u.Scheme != "file" {
				log.Errorf("Component %s has invalid source path %s", componentName, u.Scheme)
				os.Exit(1)
			}

			localLocation := util.ReadFilePath(u, runtime.GOOS)

			_, err = os.Stat(localLocation)
			if err != nil {
				odoutil.CheckError(err, "")
			}

			if sourceType == "local" {
				glog.V(4).Infof("Copying directory %s to pod", localLocation)
				err = component.PushLocal(client, componentName, applicationName, localLocation, os.Stdout, []string{})
			} else {
				dir := filepath.Dir(localLocation)
				glog.V(4).Infof("Copying file %s to pod", localLocation)
				err = component.PushLocal(client, componentName, applicationName, dir, os.Stdout, []string{localLocation})
			}
			odoutil.CheckError(err, fmt.Sprintf("Failed to push component: %v", componentName))

		case "git":
			// currently we don't support changing build type
			// it doesn't make sense to use --dir with git build
			if len(componentLocal) != 0 {
				log.Errorf("Unable to push local directory:%s to component %s that uses Git repository:%s.", componentLocal, componentName, sourcePath)
				os.Exit(1)
			}
			err := component.Build(client, componentName, applicationName, true, stdout)
			odoutil.CheckError(err, fmt.Sprintf("failed to push component: %v", componentName))
		}

		log.Successf("Changes successfully pushed to component: %v", componentName)
	},
}

// NewCmdPush implements the push odo command
func NewCmdPush() *cobra.Command {
	pushCmd.Flags().StringVarP(&componentLocal, "local", "l", "", "Use given local directory as a source for component. (It must be a local component)")

	// Add a defined annotation in order to appear in the help menu
	pushCmd.Annotations = map[string]string{"command": "component"}
	pushCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)

	//Adding `--project` flag
	projectCmd.AddProjectFlag(pushCmd)
	//Adding `--application` flag
	appCmd.AddApplicationFlag(pushCmd)

	return pushCmd
}
