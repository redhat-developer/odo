package component

import (
	"fmt"
	"net/url"
	"os"
	"runtime"

	"github.com/pkg/errors"
	"github.com/redhat-developer/odo/pkg/log"
	appCmd "github.com/redhat-developer/odo/pkg/odo/cli/application"
	projectCmd "github.com/redhat-developer/odo/pkg/odo/cli/project"
	"github.com/redhat-developer/odo/pkg/odo/util/completion"
	ktemplates "k8s.io/kubernetes/pkg/kubectl/cmd/templates"

	"github.com/golang/glog"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"

	"github.com/redhat-developer/odo/pkg/component"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/util"
	"github.com/spf13/cobra"
)

// WatchRecommendedCommandName is the recommended watch command name
const WatchRecommendedCommandName = "watch"

var watchLongDesc = ktemplates.LongDesc(`Watch for changes, update component on change.`)
var watchExample = ktemplates.Examples(`  # Watch for changes in directory for current component
%[1]s

# Watch for changes in directory for component called frontend 
%[1]s frontend
  `)

// WatchOptions contains attributes of the watch command
type WatchOptions struct {
	ignores             []string
	delay               int
	componentName       string
	componentSourceType string
	watchPath           string
	*genericclioptions.Context
}

// NewWatchOptions returns new instance of WatchOptions
func NewWatchOptions() *WatchOptions {
	return &WatchOptions{}
}

// Complete completes watch args
func (wo *WatchOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	wo.Context = genericclioptions.NewContext(cmd)

	if len(args) == 0 {
		glog.V(4).Info("No component name passed, assuming current component")
		wo.componentName = wo.Context.Component()
	} else {
		wo.componentName = args[0]
	}

	sourceType, sourcePath, err := component.GetComponentSource(wo.Context.Client, wo.componentName, wo.Context.Application)
	if err != nil {
		return errors.Wrapf(err, "Unable to get source for %s component.", wo.componentName)
	}

	u, err := url.Parse(sourcePath)
	if err != nil {
		return errors.Wrapf(err, "Unable to parse source %s from component %s.", sourcePath, wo.componentName)
	}

	if u.Scheme != "" && u.Scheme != "file" {
		log.Errorf("Component %s has invalid source path %s.", wo.componentName, u.Scheme)
		os.Exit(1)
	}

	wo.watchPath = util.ReadFilePath(u, runtime.GOOS)
	wo.componentSourceType = sourceType

	if len(wo.ignores) == 0 {
		rules, err := util.GetIgnoreRulesFromDirectory(wo.watchPath)
		if err != nil {
			odoutil.LogErrorAndExit(err, "")
		}
		wo.ignores = append(wo.ignores, rules...)
	}

	return
}

// Validate validates the watch parameters
func (wo *WatchOptions) Validate() (err error) {
	// Validate component name is non-empty
	if wo.componentName == "" {
		return fmt.Errorf(`No component is set as active.
Use 'odo component set <component name> to set and existing component as active or call this command with component name as and argument.
		`)
	}

	// Validate component path existence and accessibility permissions for odo
	if _, err := os.Stat(wo.watchPath); err != nil {
		return errors.Wrapf(err, "Cannot watch %s", wo.watchPath)
	}

	// Validate source of component is either local source or binary path until git watch is supported
	if wo.componentSourceType != "binary" && wo.componentSourceType != "local" {
		return fmt.Errorf("Watch is supported by binary and local components only and source type of component %s is %s", wo.componentName, wo.componentSourceType)
	}

	// Delay interval cannot be -ve
	if wo.delay < 0 {
		return fmt.Errorf("Delay cannot be lesser than 0 and delay=0 means changes will be pushed as soon as they are detected which can cause performance issues")
	}
	// Print a debug message warning user if delay is set to 0
	if wo.delay == 0 {
		glog.V(4).Infof("delay=0 means changes will be pushed as soon as they are detected which can cause performance issues")
	}

	return
}

// Run has the logic to perform the required actions as part of command
func (wo *WatchOptions) Run() (err error) {
	err = component.WatchAndPush(
		wo.Context.Client,
		os.Stdout,
		component.WatchParameters{
			ComponentName:   wo.componentName,
			ApplicationName: wo.Context.Application,
			Path:            wo.watchPath,
			FileIgnores:     wo.ignores,
			PushDiffDelay:   wo.delay,
			StartChan:       nil,
			ExtChan:         make(chan bool),
			WatchHandler:    component.PushLocal,
		},
	)
	if err != nil {
		return errors.Wrapf(err, "Error while trying to watch %s", wo.watchPath)
	}
	return
}

// NewCmdWatch implements the watch odo command
func NewCmdWatch(name, fullName string) *cobra.Command {
	wo := NewWatchOptions()

	var watchCmd = &cobra.Command{
		Use:     fmt.Sprintf("%s [component name]", name),
		Short:   "Watch for changes, update component on change",
		Long:    watchLongDesc,
		Example: fmt.Sprintf(watchExample, fullName),
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(wo, cmd, args)
		},
	}

	watchCmd.Flags().StringSliceVar(&wo.ignores, "ignore", []string{}, "Files or folders to be ignored via glob expressions.")
	watchCmd.Flags().IntVar(&wo.delay, "delay", 1, "Time in seconds between a detection of code change and push.delay=0 means changes will be pushed as soon as they are detected which can cause performance issues")

	// Add a defined annotation in order to appear in the help menu
	watchCmd.Annotations = map[string]string{"command": "component"}
	watchCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)

	//Adding `--application` flag
	appCmd.AddApplicationFlag(watchCmd)

	//Adding `--project` flag
	projectCmd.AddProjectFlag(watchCmd)

	completion.RegisterCommandHandler(watchCmd, completion.ComponentNameCompletionHandler)

	return watchCmd
}
