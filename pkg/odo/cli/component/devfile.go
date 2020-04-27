package component

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/openshift/odo/pkg/envinfo"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/odo/util/pushtarget"
	"github.com/openshift/odo/pkg/util"
	"github.com/pkg/errors"

	"github.com/openshift/odo/pkg/devfile/adapters"
	"github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/devfile/adapters/kubernetes"
	devfileParser "github.com/openshift/odo/pkg/devfile/parser"
	"github.com/openshift/odo/pkg/log"
)

/*
Devfile support is an experimental feature which extends the support for the
use of Che devfiles in odo for performing various odo operations.

The devfile support progress can be tracked by:
https://github.com/openshift/odo/issues/2467

Please note that this feature is currently under development and the "--devfile"
flag is exposed only if the experimental mode in odo is enabled.

The behaviour of this feature is subject to change as development for this
feature progresses.
*/

// DevfilePush has the logic to perform the required actions for a given devfile
func (po *PushOptions) DevfilePush() (err error) {
	// Parse devfile
	devObj, err := devfileParser.Parse(po.DevfilePath)
	if err != nil {
		return err
	}

	componentName, err := getComponentName()
	if err != nil {
		return errors.Wrap(err, "unable to get component name")
	}

	// Set the source path to either the context or current working directory (if context not set)
	po.sourcePath, err = util.GetAbsPath(filepath.Dir(po.componentContext))
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
			Namespace: po.namespace,
		}
		platformContext = kc
	}

	devfileHandler, err := adapters.NewPlatformAdapter(componentName, devObj, platformContext)

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
	}

	// Start or update the component
	err = devfileHandler.Push(pushParams)
	if err != nil {
		log.Errorf(
			"Failed to start component with name %s.\nError: %v",
			componentName,
			err,
		)
		os.Exit(1)
	}

	log.Infof("\nPushing devfile component %s", componentName)
	log.Success("Changes successfully pushed to component")

	return
}

// Get component name from env.yaml file
func getComponentName() (string, error) {
	// Todo: Use context to get the approraite envinfo after context is supported in experimental mode
	dir, err := os.Getwd()
	if err != nil {
		return "", err
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
	// Parse devfile
	devObj, err := devfileParser.Parse(do.devfilePath)
	if err != nil {
		return err
	}

	componentName, err := getComponentName()
	if err != nil {
		return err
	}

	kc := kubernetes.KubernetesContext{
		Namespace: do.namespace,
	}

	labels := map[string]string{
		"component": componentName,
	}
	devfileHandler, err := adapters.NewPlatformAdapter(componentName, devObj, kc)
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
