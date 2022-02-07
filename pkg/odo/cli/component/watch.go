package component

import (
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/devfile"
	"github.com/redhat-developer/odo/pkg/devfile/adapters"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/common"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/kubernetes"
	"github.com/redhat-developer/odo/pkg/envinfo"
	projectCmd "github.com/redhat-developer/odo/pkg/odo/cli/project"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/watch"

	dfutil "github.com/devfile/library/pkg/util"

	"k8s.io/klog"
	ktemplates "k8s.io/kubectl/pkg/util/templates"
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
	// Context
	*genericclioptions.Context

	// Flags
	ignoreFlag  []string
	delayFlag   int
	showLogFlag bool
	contextFlag string

	// devfile commands flags
	buildCommandFlag string
	runCommandFlag   string
	debugCommandFlag string

	sourcePath string

	// initialDevfileHandler is only used to do initial validation on the devfile.
	// All subsequent uses of the devfile adapter are generated in regenerateAdapterAndPush.
	initialDevfileHandler common.ComponentAdapter
}

// NewWatchOptions returns new instance of WatchOptions
func NewWatchOptions() *WatchOptions {
	return &WatchOptions{}
}

func (o *WatchOptions) SetClientset(clientset *clientset.Clientset) {
}

// Complete completes watch args
func (wo *WatchOptions) Complete(cmdline cmdline.Cmdline, args []string) (err error) {
	wo.Context, err = genericclioptions.New(genericclioptions.NewCreateParameters(cmdline).NeedDevfile(wo.contextFlag))
	if err != nil {
		return err
	}
	// Set the source path to either the context or current working directory (if context not set)
	wo.sourcePath, err = dfutil.GetAbsPath(wo.contextFlag)
	if err != nil {
		return errors.Wrap(err, "unable to get source path")
	}

	// Apply ignore information
	err = genericclioptions.ApplyIgnore(&wo.ignoreFlag, wo.sourcePath)
	if err != nil {
		return errors.Wrap(err, "unable to apply ignore information")
	}

	// The namespace was retrieved from the --project flag (or from the kube client if not set) and stored in kclient when initializing the context
	platformContext := kubernetes.KubernetesContext{
		Namespace: wo.KClient.GetCurrentNamespace(),
	}

	wo.initialDevfileHandler, err = adapters.NewComponentAdapter(wo.EnvSpecificInfo.GetName(), wo.contextFlag, wo.GetApplication(), wo.EnvSpecificInfo.GetDevfileObj(), platformContext)

	return err
}

// Validate validates the watch parameters
func (wo *WatchOptions) Validate() (err error) {

	// Delay interval cannot be -ve
	if wo.delayFlag < 0 {
		return fmt.Errorf("Delay cannot be lesser than 0 and delay=0 means changes will be pushed as soon as they are detected which can cause performance issues")
	}
	// Print a debug message warning user if delay is set to 0
	if wo.delayFlag == 0 {
		klog.V(4).Infof("delay=0 means changes will be pushed as soon as they are detected which can cause performance issues")
	}

	if wo.debugCommandFlag != "" && wo.EnvSpecificInfo != nil && wo.EnvSpecificInfo.GetRunMode() != envinfo.Debug {
		return fmt.Errorf("please start the component in debug mode using `odo push --debug` to use the --debug-command flag")
	}
	exists, err := component.ComponentExists(wo.KClient, wo.EnvSpecificInfo.GetName(), wo.GetApplication())
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("component does not exist. Please use `odo push` to create your component")
	}
	return nil
}

// Run has the logic to perform the required actions as part of command
func (wo *WatchOptions) Run() (err error) {
	err = watch.DevfileWatchAndPush(
		os.Stdout,
		watch.WatchParameters{
			//ComponentName: wo.EnvSpecificInfo.GetName(),
			//ApplicationName:     wo.Context.GetApplication(),
			Path:                wo.sourcePath,
			FileIgnores:         dfutil.GetAbsGlobExps(wo.sourcePath, wo.ignoreFlag),
			PushDiffDelay:       wo.delayFlag,
			StartChan:           nil,
			ExtChan:             make(chan bool),
			DevfileWatchHandler: wo.regenerateAdapterAndPush,
			Show:                wo.showLogFlag,
			DevfileBuildCmd:     strings.ToLower(wo.buildCommandFlag),
			DevfileRunCmd:       strings.ToLower(wo.runCommandFlag),
			DevfileDebugCmd:     strings.ToLower(wo.debugCommandFlag),
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

	watchCmd.Flags().BoolVar(&wo.showLogFlag, "show-log", false, "If enabled, logs will be shown when built")
	watchCmd.Flags().StringSliceVar(&wo.ignoreFlag, "ignore", []string{}, "Files or folders to be ignored via glob expressions.")
	watchCmd.Flags().IntVar(&wo.delayFlag, "delay", 1, "Time in seconds between a detection of code change and push.delay=0 means changes will be pushed as soon as they are detected which can cause performance issues")

	watchCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)

	watchCmd.Flags().StringVar(&wo.buildCommandFlag, "build-command", "", "Devfile Build Command to execute")
	watchCmd.Flags().StringVar(&wo.runCommandFlag, "run-command", "", "Devfile Run Command to execute")
	watchCmd.Flags().StringVar(&wo.debugCommandFlag, "debug-command", "", "Devfile Debug Command to execute")

	// Adding context flag
	odoutil.AddContextFlag(watchCmd, &wo.contextFlag)

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
	devObj, err := devfile.ParseAndValidateFromFile(wo.GetDevfilePath())
	if err != nil {
		return nil, err
	}

	platformContext := kubernetes.KubernetesContext{
		Namespace: wo.KClient.GetCurrentNamespace(),
	}

	return adapters.NewComponentAdapter(parameters.ComponentName, parameters.Path, parameters.ApplicationName, devObj, platformContext)

}
