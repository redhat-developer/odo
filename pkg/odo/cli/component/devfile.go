package component

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/openshift/odo/pkg/devfile"
	"github.com/openshift/odo/pkg/devfile/adapters"
<<<<<<< HEAD
=======
	//	"github.com/openshift/odo/pkg/devfile/parser"
	"github.com/openshift/odo/pkg/devfile/parser"

>>>>>>> e792124a71dde2630bd692553d694bdb7f169478
	"github.com/openshift/odo/pkg/envinfo"
	"github.com/openshift/odo/pkg/machineoutput"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/odo/util/pushtarget"
	"github.com/openshift/odo/pkg/util"
	"github.com/pkg/errors"

	"github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/devfile/adapters/kubernetes"
	"github.com/openshift/odo/pkg/log"
)

/*
Devfile support is an experimental feature which extends the support for the
use of Che devfiles in odo for performing various odo operations.

The devfile support progress can be tracked by:
https://github.com/openshift/odo/issues/2467

Please note that this feature is currently under development,
the feature will be available with experimental mode enabled.

The behaviour of this feature is subject to change as development for this
feature progresses.
*/

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

	// Parse devfile and validate
	devObj, err := devfile.ParseAndValidate(po.DevfilePath)
	if err != nil {
		return err
	}

	componentName, err := getComponentName(po.componentContext)
	if err != nil {
		return errors.Wrap(err, "unable to get component name")
	}

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
	if pushtarget.IsPushTargetDocker() {
		platformContext = nil
	} else {
		kc := kubernetes.KubernetesContext{
			Namespace: po.KClient.Namespace,
		}
		platformContext = kc
	}

	devfileHandler, err := adapters.NewComponentAdapter(componentName, po.componentContext, po.Application, devObj, platformContext)

	if err != nil {
		return err
	}
	pushParams := common.PushParameters{
		Path:            po.sourcePath,
		IgnoredFiles:    po.ignores,
		ForceBuild:      po.forceBuild,
		Show:            po.show,
		EnvSpecificInfo: *po.EnvSpecificInfo,
		DevfileInitCmd:  strings.ToLower(po.devfileInitCommand),
		DevfileBuildCmd: strings.ToLower(po.devfileBuildCommand),
		DevfileRunCmd:   strings.ToLower(po.devfileRunCommand),
		DevfileDebugCmd: strings.ToLower(po.devfileDebugCommand),
		Debug:           po.debugRun,
		DebugPort:       po.EnvSpecificInfo.GetDebugPort(),
	}

	warnIfURLSInvalid(po.EnvSpecificInfo.GetURL())

	// Start or update the component
	err = devfileHandler.Push(pushParams)
	if err != nil {
		err = errors.Errorf("Failed to start component with name %s. Error: %v",
			componentName,
			err,
		)
	} else {
		log.Infof("\nPushing devfile component %s", componentName)
		log.Success("Changes successfully pushed to component")
	}

	return
}

//DevfileDeploy
func (do *DeployOptions) DevfileDeploy() (err error) {
	componentName, err := getComponentName(do.componentContext)
	if err != nil {
		return errors.Wrap(err, "unable to get component name")
	}

	// Set the source path to either the context or current working directory (if context not set)
	do.sourcePath, err = util.GetAbsPath(do.componentContext)
	if err != nil {
		return errors.Wrap(err, "unable to get source path")
	}

	// Apply ignore information
	err = genericclioptions.ApplyIgnore(&do.ignores, do.sourcePath)
	if err != nil {
		return errors.Wrap(err, "unable to apply ignore information")
	}

	kubeContext := kubernetes.KubernetesContext{
		Namespace: do.namespace,
	}

	devfileHandler, err := adapters.NewComponentAdapter(componentName, do.componentContext, do.Application, do.devObj, kubeContext)
	if err != nil {
		return err
	}

	buildParams := common.BuildParameters{
		Path:                     do.sourcePath,
		Tag:                      do.tag,
		DockerConfigJSONFilename: do.dockerConfigJSONFilename,
		DockerfileBytes:          do.DockerfileBytes,
		EnvSpecificInfo:          *do.EnvSpecificInfo,
	}

	log.Infof("\nBuilding component %s", componentName)
	// Build image for the component
	err = devfileHandler.Build(buildParams)
	if err != nil {
		log.Errorf(
			"Failed to build component with name %s.\nError: %v",
			componentName,
			err,
		)
		os.Exit(1)
	}

	time.Sleep(5 * time.Second)

	deployParams := common.DeployParameters{
		EnvSpecificInfo: *do.EnvSpecificInfo,
		Tag:             do.tag,
		ManifestSource:  do.ManifestSource,
		DeploymentPort:  do.DeploymentPort,
	}

	warnIfURLSInvalid(do.EnvSpecificInfo.GetURL())

	log.Infof("\nDeploying component %s", componentName)
	// Deploy the application
	err = devfileHandler.Deploy(deployParams)
	if err != nil {
		log.Errorf(
			"Failed to deploy application with name %s.\nError: %v",
			componentName,
			err,
		)
		os.Exit(1)
	}

	return nil
}

// DevfileComponentDelete deletes the devfile component
func (ddo *DeployDeleteOptions) DevfileDeployDelete() error {
	// Parse devfile
	devObj, err := devfile.ParseAndValidate(ddo.DevfilePath)
	if err != nil {
		return err
	}

	componentName, err := getComponentName(ddo.componentContext)
	if err != nil {
		return err
	}
	componentName = componentName + "-deploy"

	kc := kubernetes.KubernetesContext{
		Namespace: ddo.namespace,
	}

	devfileHandler, err := adapters.NewComponentAdapter(componentName, ddo.componentContext, ddo.Application, devObj, kc)
	if err != nil {
		return err
	}

	spinner := log.Spinner(fmt.Sprintf("Deleting deployed devfile component %s", componentName))
	defer spinner.End(false)

	manifestErr := devfileHandler.DeployDelete(ddo.ManifestSource)
	if manifestErr != nil && strings.Contains(manifestErr.Error(), "as component was not found") {
		log.Warning(manifestErr.Error())
		err = os.Remove(ddo.ManifestPath)
		if err != nil {
			return err
		}
		spinner.End(false)
		log.Success(ddo.ManifestPath + " deleted. Exiting gracefully :)")
		return nil
	} else if manifestErr != nil {
		err = os.Remove(ddo.ManifestPath)
		return err
	}

	err = os.Remove(ddo.ManifestPath)
	if err != nil {
		return err
	}

	spinner.End(true)
	log.Successf("Successfully deleted component")
	return nil
}

// DevfileComponentLog fetch and display log from devfile components
func (lo LogOptions) DevfileComponentLog() error {
	// Parse devfile
	devObj, err := devfile.ParseAndValidate(lo.devfilePath)
	if err != nil {
		return err
	}

	componentName, err := getComponentName(lo.componentContext)
	if err != nil {
		return errors.Wrap(err, "unable to get component name")
	}

	var platformContext interface{}
	if pushtarget.IsPushTargetDocker() {
		platformContext = nil
	} else {
		kc := kubernetes.KubernetesContext{
			Namespace: lo.KClient.Namespace,
		}
		platformContext = kc
	}

	devfileHandler, err := adapters.NewComponentAdapter(componentName, lo.componentContext, lo.Application, devObj, platformContext)

	if err != nil {
		return err
	}

	// Start or update the component
	rd, err := devfileHandler.Log(lo.logFollow, lo.debug)
	if err != nil {
		log.Errorf(
			"Failed to log component with name %s.\nError: %v",
			componentName,
			err,
		)
		os.Exit(1)
	}

	return util.DisplayLog(lo.logFollow, rd, componentName)
}

// Get component name from env.yaml file
func getComponentName(context string) (string, error) {
	var dir string
	var err error
	if context == "" {
		dir, err = os.Getwd()
		if err != nil {
			return "", err
		}
	} else {
		dir = context
	}

	envInfo, err := envinfo.NewEnvSpecificInfo(dir)
	if err != nil {
		return "", err
	}
	componentName := envInfo.GetName()
	return componentName, nil
}

// DevfileComponentDelete deletes the devfile component
func (do *DeleteOptions) DevfileComponentDelete() error {
	// Parse devfile and validate
	devObj, err := devfile.ParseAndValidate(do.devfilePath)
	if err != nil {
		return err
	}

	componentName, err := getComponentName(do.componentContext)
	if err != nil {
		return err
	}

	kc := kubernetes.KubernetesContext{
		Namespace: do.namespace,
	}

	labels := map[string]string{
		"component": componentName,
	}
	devfileHandler, err := adapters.NewComponentAdapter(componentName, do.componentContext, do.Application, devObj, kc)
	if err != nil {
		return err
	}

	return devfileHandler.Delete(labels)
}

// RunTestCommand runs the specific test command in devfile
func (to *TestOptions) RunTestCommand() error {
	componentName, err := getComponentName(to.componentContext)
	if err != nil {
		return err
	}

	var platformContext interface{}
	if pushtarget.IsPushTargetDocker() {
		platformContext = nil
	} else {
		kc := kubernetes.KubernetesContext{
			Namespace: to.KClient.Namespace,
		}
		platformContext = kc
	}

	devfileHandler, err := adapters.NewComponentAdapter(componentName, to.componentContext, to.Application, to.devObj, platformContext)
	if err != nil {
		return err
	}
	return devfileHandler.Test(to.commandName, to.show)
}

func warnIfURLSInvalid(url []envinfo.EnvInfoURL) {
	// warnIfURLSInvalid checks if env.yaml contains a valide URL for the current pushtarget
	// display a warning if no url(s) found for the current push target, but found url(s) for another push target
	dockerURLExists := false
	kubeURLExists := false
	for _, element := range url {
		if element.Kind == envinfo.DOCKER {
			dockerURLExists = true
		} else {
			kubeURLExists = true
		}
	}
	var urlOutput string
	if len(url) > 1 {
		urlOutput = "URLs"
	} else {
		urlOutput = "a URL"
	}
	if pushtarget.IsPushTargetDocker() && !dockerURLExists && kubeURLExists {
		log.Warningf("Found %v defined for Kubernetes, but no valid URLs for Docker.", urlOutput)
	} else if !pushtarget.IsPushTargetDocker() && !kubeURLExists && dockerURLExists {
		log.Warningf("Found %v defined for Docker, but no valid URLs for Kubernetes.", urlOutput)
	}
}

// DevfileComponentExec executes the given user command inside the component
func (eo *ExecOptions) DevfileComponentExec(command []string) error {
	// Parse devfile
	devObj, err := devfile.ParseAndValidate(eo.devfilePath)
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
