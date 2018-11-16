package cli

import (
	"fmt"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"net/url"
	"os"
	"runtime"

	odoutil "github.com/redhat-developer/odo/pkg/odo/util"

	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/util"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
)

var (
	ignores []string
	delay   int
)

var watchCmd = &cobra.Command{
	Use:   "watch [component name]",
	Short: "Watch for changes, update component on change",
	Long:  `Watch for changes, update component on change.`,
	Example: `  # Watch for changes in directory for current component
  odo watch

  # Watch for changes in directory for component called frontend 
  odo watch frontend
	`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		stdout := os.Stdout
		context := genericclioptions.NewContext(cmd)
		client := context.Client
		projectName := context.Project
		applicationName := context.Application

		// TODO: check if we can use context.Component() here
		var componentName string
		if len(args) == 0 {
			var err error
			glog.V(4).Info("No component name passed, assuming current component")
			componentName, err = component.GetCurrent(applicationName, projectName)
			odoutil.CheckError(err, "")
			if componentName == "" {
				fmt.Println("No component is set as active.")
				fmt.Println("Use 'odo component set <component name> to set and existing component as active or call this command with component name as and argument.")
				os.Exit(1)
			}
		} else {
			componentName = args[0]
		}

		sourceType, sourcePath, err := component.GetComponentSource(client, componentName, applicationName)
		odoutil.CheckError(err, "Unable to get source for %s component.", componentName)

		if sourceType != "binary" && sourceType != "local" {
			fmt.Printf("Watch is supported by binary and local components only and source type of component %s is %s\n", componentName, sourceType)
			os.Exit(1)
		}

		u, err := url.Parse(sourcePath)
		odoutil.CheckError(err, "Unable to parse source %s from component %s.", sourcePath, componentName)

		if u.Scheme != "" && u.Scheme != "file" {
			fmt.Printf("Component %s has invalid source path %s.", componentName, u.Scheme)
			os.Exit(1)
		}
		watchPath := util.ReadFilePath(u, runtime.GOOS)

		err = component.WatchAndPush(client, componentName, applicationName, watchPath, stdout, ignores, delay, make(chan string), component.PushLocal)
		odoutil.CheckError(err, "Error while trying to watch %s", watchPath)
	},
}

func init() {
	// ignore git as it can change even if no source file changed
	// for example some plugins providing git info in PS1 doing that
	watchCmd.Flags().StringSliceVar(&ignores, "ignore", []string{".*\\.git.*"}, "Files or folders to be ignored via regular expressions.")
	watchCmd.Flags().IntVar(&delay, "delay", 1, "Time in seconds between a detection of code change and push.delay=0 means changes will be pushed as soon as they are detected which can cause performance issues")
	// Add a defined annotation in order to appear in the help menu
	watchCmd.Annotations = map[string]string{"command": "component"}
	watchCmd.SetUsageTemplate(CmdUsageTemplate)

	//Adding `--application` flag
	AddApplicationFlag(watchCmd)

	//Adding `--project` flag
	AddProjectFlag(watchCmd)

	rootCmd.AddCommand(watchCmd)
}
