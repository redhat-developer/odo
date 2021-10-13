package component

import (
	"os"
	"reflect"
	"strings"

	"github.com/openshift/odo/pkg/devfile"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/openshift/odo/pkg/envinfo"
	"github.com/openshift/odo/pkg/machineoutput"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/util"
	"github.com/pkg/errors"

	componentlabels "github.com/openshift/odo/pkg/component/labels"
	"github.com/openshift/odo/pkg/devfile/adapters"
	"github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/devfile/adapters/kubernetes"
	"github.com/openshift/odo/pkg/log"
)

// DevfilePush has the logic to perform the required actions for a given devfile
func (po *PushOptions) DevfilePush() error {

	// Wrap the push so that we can capture the error in JSON-only mode
	err := po.devfilePushInner()

	if err != nil && log.IsJSON() {
		eventLoggingClient := machineoutput.NewConsoleMachineEventLoggingClient()
		eventLoggingClient.ReportError(err, machineoutput.TimestampNow())

		// Suppress the error to prevent it from being output by the generic machine-readable handler (which will produce invalid JSON for our purposes)
		err = nil

		// os.Exit(1) since we are suppressing the generic machine-readable handler's exit code logic
		os.Exit(1)
	}

	if err != nil {
		return err
	}

	// push is successful, save the runMode used
	runMode := envinfo.Run
	if po.debugRun {
		runMode = envinfo.Debug
	}

	return po.EnvSpecificInfo.SetRunMode(runMode)
}

func (po *PushOptions) devfilePushInner() (err error) {
	devObj, err := devfile.ParseFromFile(po.DevfilePath)
	if err != nil {
		return err
	}
	componentName := po.EnvSpecificInfo.GetName()

	// Set the source path to either the context or current working directory (if context not set)
	po.sourcePath, err = util.GetAbsPath(po.componentContext)
	if err != nil {
		return errors.Wrap(err, "unable to get source path")
	}

	// Apply ignore information
	err = genericclioptions.ApplyIgnore(&po.ignores, po.sourcePath)
	if err != nil {
		return errors.Wrap(err, "unable to apply ignore information")
	}

	var platformContext interface{}
	kc := kubernetes.KubernetesContext{
		Namespace: po.KClient.GetCurrentNamespace(),
	}
	platformContext = kc

	devfileHandler, err := adapters.NewComponentAdapter(componentName, po.sourcePath, po.Application, devObj, platformContext)
	if err != nil {
		return err
	}

	pushParams := common.PushParameters{
		Path:            po.sourcePath,
		IgnoredFiles:    po.ignores,
		ForceBuild:      po.forceBuild,
		Show:            po.show,
		EnvSpecificInfo: *po.EnvSpecificInfo,
		DevfileBuildCmd: strings.ToLower(po.devfileBuildCommand),
		DevfileRunCmd:   strings.ToLower(po.devfileRunCommand),
		DevfileDebugCmd: strings.ToLower(po.devfileDebugCommand),
		Debug:           po.debugRun,
		DebugPort:       po.EnvSpecificInfo.GetDebugPort(),
	}

	_, err = po.EnvSpecificInfo.ListURLs()
	if err != nil {
		return err
	}

	// Start or update the component
	err = devfileHandler.Push(pushParams)
	if err != nil {
		err = errors.Errorf("Failed to start component with name %q. Error: %v",
			componentName,
			err,
		)
	} else {
		log.Infof("\nPushing devfile component %q", componentName)
		log.Success("Changes successfully pushed to component")
	}

	return
}

// DevfileComponentLog fetch and display log from devfile components
func (lo LogOptions) DevfileComponentLog() error {
	devObj, err := devfile.ParseFromFile(lo.devfilePath)
	if err != nil {
		return err
	}

	componentName := lo.Context.EnvSpecificInfo.GetName()

	var platformContext interface{}
	kc := kubernetes.KubernetesContext{
		Namespace: lo.KClient.GetCurrentNamespace(),
	}
	platformContext = kc

	devfileHandler, err := adapters.NewComponentAdapter(componentName, lo.componentContext, lo.Application, devObj, platformContext)

	if err != nil {
		return err
	}

	var command devfilev1.Command
	if lo.debug {
		command, err = common.GetDebugCommand(devObj.Data, "")
		if err != nil {
			return err
		}
		if reflect.DeepEqual(devfilev1.Command{}, command) {
			return errors.Errorf("no debug command found in devfile, please run \"odo log\" for run command logs")
		}

	} else {
		command, err = common.GetRunCommand(devObj.Data, "")
		if err != nil {
			return err
		}
	}

	// Start or update the component
	rd, err := devfileHandler.Log(lo.logFollow, command)
	if err != nil {
		log.Errorf(
			"Failed to log component with name %s.\nError: %v",
			componentName,
			err,
		)
		return err
	}

	return util.DisplayLog(lo.logFollow, rd, os.Stdout, componentName, -1)
}

// DevfileComponentDelete deletes the devfile component
func (do *DeleteOptions) DevfileComponentDelete() error {
	devObj, err := devfile.ParseFromFile(do.devfilePath)
	if err != nil {
		return err
	}

	componentName := do.EnvSpecificInfo.GetName()

	kc := kubernetes.KubernetesContext{
		Namespace: do.namespace,
	}

	labels := componentlabels.GetLabels(componentName, do.EnvSpecificInfo.GetApplication(), false)
	devfileHandler, err := adapters.NewComponentAdapter(componentName, do.componentContext, do.Application, devObj, kc)
	if err != nil {
		return err
	}

	return devfileHandler.Delete(labels, do.show, do.componentDeleteWaitFlag)
}

// RunTestCommand runs the specific test command in devfile
func (to *TestOptions) RunTestCommand() error {
	componentName := to.Context.EnvSpecificInfo.GetName()

	var platformContext interface{}
	kc := kubernetes.KubernetesContext{
		Namespace: to.KClient.GetCurrentNamespace(),
	}
	platformContext = kc

	devfileHandler, err := adapters.NewComponentAdapter(componentName, to.componentContext, to.Application, to.devObj, platformContext)
	if err != nil {
		return err
	}
	return devfileHandler.Test(to.commandName, to.show)
}

// DevfileComponentExec executes the given user command inside the component
func (eo *ExecOptions) DevfileComponentExec(command []string) error {
	devObj, err := devfile.ParseFromFile(eo.devfilePath)
	if err != nil {
		return err
	}

	componentName := eo.componentOptions.EnvSpecificInfo.GetName()

	kc := kubernetes.KubernetesContext{
		Namespace: eo.namespace,
	}

	devfileHandler, err := adapters.NewComponentAdapter(componentName, eo.componentContext, eo.componentOptions.Application, devObj, kc)
	if err != nil {
		return err
	}

	return devfileHandler.Exec(command)
}
