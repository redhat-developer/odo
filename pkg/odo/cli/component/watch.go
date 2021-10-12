package component

import (
	"fmt"
	"os"
	"strings"

	"github.com/openshift/odo/pkg/devfile"
	"github.com/openshift/odo/pkg/devfile/adapters"
	"github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/devfile/adapters/kubernetes"
	"github.com/openshift/odo/pkg/envinfo"
	appCmd "github.com/openshift/odo/pkg/odo/cli/application"
	projectCmd "github.com/openshift/odo/pkg/odo/cli/project"
	"github.com/pkg/errors"
	ktemplates "k8s.io/kubectl/pkg/util/templates"

	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"k8s.io/klog"

	odoutil "github.com/openshift/odo/pkg/odo/util"
	"github.com/openshift/odo/pkg/util"
	"github.com/openshift/odo/pkg/watch"
	"github.com/spf13/cobra"
)

// WatchRecommendedCommandName is the recommended watch command name
const WatchRecommendedCommandName = "watch"

var watchLongDesc = ktemplates.LongDesc(`Watch for changes, update component on change. Watch doesn't support git components.`)
var watchExampleWithDevfile = ktemplates.Examples(`  # Watch for changes in directory for current component
%[1]s

# Watch source code changes with custom devfile commands using --build-command, --run-command and --debug-command for devfile based components
%[1]s --build-command="mybuild" --run-command="myrun" --debug-command="mydebug"
  `)

// WatchOptions contains attributes of the watch command
type WatchOptions struct {
	ignores []string
	delay   int
	show    bool

	sourcePath       string
	componentContext string

	componentName string
	devfilePath   string
	namespace     string

	// initialDevfileHandler is only used to do initial validation on the devfile.
	// All subsequent uses of the devfile adapter are generated in regenerateAdapterAndPush.
	initialDevfileHandler common.ComponentAdapter

	// devfile commands
	devfileBuildCommand string
	devfileRunCommand   string
	devfileDebugCommand string

	*genericclioptions.Context
}

// NewWatchOptions returns new instance of WatchOptions
func NewWatchOptions() *WatchOptions {
	return &WatchOptions{}
}

// Complete completes watch args
func (wo *WatchOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	wo.devfilePath = devfile.DevfileLocation(wo.componentContext)

	wo.Context, err = genericclioptions.NewContext(cmd)
	if err != nil {
		return err
	}
	// Set the source path to either the context or current working directory (if context not set)
	wo.sourcePath, err = util.GetAbsPath(wo.componentContext)
	if err != nil {
		return errors.Wrap(err, "unable to get source path")
	}

	// Apply ignore information
	err = genericclioptions.ApplyIgnore(&wo.ignores, wo.sourcePath)
	if err != nil {
		return errors.Wrap(err, "unable to apply ignore information")
	}

	// Get the component name
	wo.componentName = wo.EnvSpecificInfo.GetName()

	// Parse devfile and validate
	devObj, err := devfile.ParseFromFile(wo.devfilePath)
	if err != nil {
		return err
	}

	var platformContext interface{}
	// The namespace was retrieved from the --project flag (or from the kube client if not set) and stored in kclient when initializing the context
	wo.namespace = wo.KClient.GetCurrentNamespace()
	platformContext = kubernetes.KubernetesContext{
		Namespace: wo.namespace,
	}

	wo.initialDevfileHandler, err = adapters.NewComponentAdapter(wo.componentName, wo.componentContext, wo.Application, devObj, platformContext)

	return err
}

// Validate validates the watch parameters
func (wo *WatchOptions) Validate() (err error) {

	// Delay interval cannot be -ve
	if wo.delay < 0 {
		return fmt.Errorf("Delay cannot be lesser than 0 and delay=0 means changes will be pushed as soon as they are detected which can cause performance issues")
	}
	// Print a debug message warning user if delay is set to 0
	if wo.delay == 0 {
		klog.V(4).Infof("delay=0 means changes will be pushed as soon as they are detected which can cause performance issues")
	}

	if wo.devfileDebugCommand != "" && wo.EnvSpecificInfo != nil && wo.EnvSpecificInfo.GetRunMode() != envinfo.Debug {
		return fmt.Errorf("please start the component in debug mode using `odo push --debug` to use the --debug-command flag")
	}
	exists, err := wo.initialDevfileHandler.DoesComponentExist(wo.componentName, wo.Application)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("component does not exist. Please use `odo push` to create your component")
	}
	return nil
}

// Run has the logic to perform the required actions as part of command
func (wo *WatchOptions) Run(cmd *cobra.Command) (err error) {
	err = watch.DevfileWatchAndPush(
		os.Stdout,
		watch.WatchParameters{
			ComponentName:       wo.componentName,
			ApplicationName:     wo.Context.Application,
			Path:                wo.sourcePath,
			FileIgnores:         util.GetAbsGlobExps(wo.sourcePath, wo.ignores),
			PushDiffDelay:       wo.delay,
			StartChan:           nil,
			ExtChan:             make(chan bool),
			DevfileWatchHandler: wo.regenerateAdapterAndPush,
			Show:                wo.show,
			DevfileBuildCmd:     strings.ToLower(wo.devfileBuildCommand),
			DevfileRunCmd:       strings.ToLower(wo.devfileRunCommand),
			DevfileDebugCmd:     strings.ToLower(wo.devfileDebugCommand),
			EnvSpecificInfo:     wo.EnvSpecificInfo,
		},
	)
	if err != nil {
		return errors.Wrapf(err, "Error while trying to watch %s", wo.sourcePath)
	}
	return err
}

// NewCmdWatch implements the watch odo command
func NewCmdWatch(name, fullName string) *cobra.Command {
	wo := NewWatchOptions()

	usage := name

	// Add information on Devfile
	example := fmt.Sprintf(watchExampleWithDevfile, fullName)

	var watchCmd = &cobra.Command{
		Use:         usage,
		Short:       "Watch for changes, update component on change. Watch doesn't support git components.",
		Long:        watchLongDesc,
		Example:     example,
		Args:        cobra.NoArgs,
		Annotations: map[string]string{"command": "component"},
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(wo, cmd, args)
		},
	}

	watchCmd.Flags().BoolVar(&wo.show, "show-log", false, "If enabled, logs will be shown when built")
	watchCmd.Flags().StringSliceVar(&wo.ignores, "ignore", []string{}, "Files or folders to be ignored via glob expressions.")
	watchCmd.Flags().IntVar(&wo.delay, "delay", 1, "Time in seconds between a detection of code change and push.delay=0 means changes will be pushed as soon as they are detected which can cause performance issues")

	watchCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)

	watchCmd.Flags().StringVar(&wo.devfileBuildCommand, "build-command", "", "Devfile Build Command to execute")
	watchCmd.Flags().StringVar(&wo.devfileRunCommand, "run-command", "", "Devfile Run Command to execute")
	watchCmd.Flags().StringVar(&wo.devfileDebugCommand, "debug-command", "", "Devfile Debug Command to execute")

	// Adding context flag
	genericclioptions.AddContextFlag(watchCmd, &wo.componentContext)

	//Adding `--application` flag
	appCmd.AddApplicationFlag(watchCmd)

	//Adding `--project` flag
	projectCmd.AddProjectFlag(watchCmd)

	return watchCmd
}

// regenerateAdapterAndPush is used as a DevfileWatchHandler in WatchParameters; it is a wrapper around adapter.Push()
// that first regenerates the component adapter before calling push. This ensures that it has picked up the latest
// devfile.yaml changes
func (wo *WatchOptions) regenerateAdapterAndPush(pushParams common.PushParameters, watchParams watch.WatchParameters) error {
	var adapter common.ComponentAdapter

	adapter, err := wo.regenerateComponentAdapterFromWatchParams(watchParams)
	if err != nil {
		return errors.Wrapf(err, "unable to generate component from watch parameters")
	}

	err = adapter.Push(pushParams)
	if err != nil {
		return errors.Wrapf(err, "watch command was unable to push component")
	}

	return err
}

// regenerateComponentAdapterFromWatchParams (re)generates a component adapter from the given watch parameters.
func (wo *WatchOptions) regenerateComponentAdapterFromWatchParams(parameters watch.WatchParameters) (common.ComponentAdapter, error) {

	// Parse devfile and validate
	devObj, err := devfile.ParseFromFile(wo.devfilePath)
	if err != nil {
		return nil, err
	}

	platformContext := kubernetes.KubernetesContext{
		Namespace: wo.namespace,
	}

	return adapters.NewComponentAdapter(parameters.ComponentName, parameters.Path, parameters.ApplicationName, devObj, platformContext)

}
