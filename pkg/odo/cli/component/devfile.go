package component

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/openshift/odo/pkg/component"
	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/devfile"
	"github.com/openshift/odo/pkg/envinfo"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	odoutil "github.com/openshift/odo/pkg/odo/util"
	"github.com/openshift/odo/pkg/util"
	"github.com/pkg/errors"

	"github.com/openshift/odo/pkg/devfile/adapters"
	"github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/devfile/adapters/kubernetes"
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
	devObj, err := devfile.Parse(po.DevfilePath)
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

	spinner := log.SpinnerNoSpin(fmt.Sprintf("Push devfile component %s", componentName))
	defer spinner.End(false)
	if len(po.namespace) <= 0 {
		po.namespace, err = getNamespace()
		if err != nil {
			return err
		}
	}
	kc := kubernetes.KubernetesContext{
		Namespace: po.namespace,
	}
	devfileHandler, err := adapters.NewPlatformAdapter(componentName, devObj, kc)

	if err != nil {
		return err
	}

	po.Context.KClient.Namespace = po.namespace
	err = component.ApplyConfig(nil, po.Context.KClient, config.LocalConfigInfo{}, *po.EnvSpecificInfo, color.Output, po.doesComponentExist)
	if err != nil {
		odoutil.LogErrorAndExit(err, "Failed to update config to component deployed.")
	}

	pushParams := common.PushParameters{
		Path:            po.sourcePath,
		IgnoredFiles:    po.ignores,
		ForceBuild:      po.forceBuild,
		Show:            po.show,
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

	spinner.End(true)

	log.Success("Changes successfully pushed to component")
	return
}

// Get component name from env.yaml file
func getComponentName() (string, error) {
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

// Get namespace name from env.yaml file
func getNamespace() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	envInfo, err := envinfo.NewEnvSpecificInfo(dir)
	if err != nil {
		return "", err
	}
	Namespace := envInfo.GetNamespace()
	return Namespace, nil
}
