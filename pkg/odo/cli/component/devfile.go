package component

import (
	"fmt"
	"os"
	"strings"

	"github.com/openshift/odo/pkg/devfile/adapters"
	"github.com/openshift/odo/pkg/devfile/parser"
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

	return err
}

func (po *PushOptions) devfilePushInner() (err error) {
	// Parse devfile and validate
	devObj, err := parser.ParseAndValidate(po.DevfilePath)
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

	devfileHandler, err := adapters.NewComponentAdapter(componentName, po.componentContext, devObj, platformContext)

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

	devfileHandler, err := adapters.NewComponentAdapter(componentName, do.componentContext, do.devObj, kubeContext)
	if err != nil {
		return err
	}

	buildParams := common.BuildParameters{
		Path:            do.sourcePath,
		Tag:             do.tag,
		DockerfileBytes: do.DockerfileBytes,
		EnvSpecificInfo: *do.EnvSpecificInfo,
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
	log.Successf("Successfully built container image: %s", do.tag)

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
	devObj, err := parser.ParseAndValidate(ddo.DevfilePath)
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

	devfileHandler, err := adapters.NewComponentAdapter(componentName, ddo.componentContext, devObj, kc)
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
	devObj, err := parser.ParseAndValidate(do.devfilePath)
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
	devfileHandler, err := adapters.NewComponentAdapter(componentName, do.componentContext, devObj, kc)
	if err != nil {
		return err
	}

	spinner := log.Spinner(fmt.Sprintf("Deleting devfile component %s", componentName))
	defer spinner.End(false)

	err = devfileHandler.Delete(labels)
	if err != nil {
		return err
	}
	spinner.End(true)
	log.Successf("Successfully deleted component")
	return nil
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
