package component

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/openshift/odo/pkg/devfile"
	"github.com/openshift/odo/pkg/envinfo"
	"github.com/openshift/odo/pkg/util"
	"github.com/pkg/errors"

	"github.com/openshift/odo/pkg/devfile/adapters"
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

const componentNameMaxLen = 45

// DevfilePush has the logic to perform the required actions for a given devfile
func (po *PushOptions) DevfilePush() (err error) {
	// Parse devfile
	devObj, err := devfile.Parse(po.devfilePath)
	if err != nil {
		return err
	}

	// Get component name from env.yaml file
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	envInfo, err := envinfo.NewEnvSpecificInfo(dir)
	if err != nil {
		return err
	}
	componentName := envInfo.GetName()

	// Set the source path to either the context or current working directory (if context not set)
	po.sourcePath, err = util.GetAbsPath(filepath.Dir(po.componentContext))
	if err != nil {
		return errors.Wrap(err, "unable to get source path")
	}

	spinner := log.SpinnerNoSpin(fmt.Sprintf("Push devfile component %s", componentName))
	defer spinner.End(false)

	kc := kubernetes.KubernetesContext{
		Namespace: po.namespace,
	}
	devfileHandler, err := adapters.NewPlatformAdapter(componentName, devObj, kc)

	if err != nil {
		return err
	}

	// Start or update the component
	err = devfileHandler.Push(po.sourcePath, po.ignores, po.forceBuild, util.GetAbsGlobExps(po.sourcePath, po.ignores))
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
