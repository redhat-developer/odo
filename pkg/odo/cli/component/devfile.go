package component

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/openshift/odo/pkg/devfile"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/odo/util/pushtarget"
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

const componentNameMaxLen = 45

// DevfilePush has the logic to perform the required actions for a given devfile
func (po *PushOptions) DevfilePush() (err error) {
	// Parse devfile
	devObj, err := devfile.Parse(po.devfilePath)
	if err != nil {
		return err
	}

	componentName, err := getComponentName()
	if err != nil {
		return err
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
		Path:         po.sourcePath,
		IgnoredFiles: po.ignores,
		ForceBuild:   po.forceBuild,
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

/*
 * getComponentName generates a component name by using the current directory's name and manipulates it if needed so that it
 * can be used for kubernetes resource names as well. This will likely be moved/replaced once devfile create is
 * implemented because component name should be determined at that point.
 */
func getComponentName() (string, error) {
	retVal := ""
	currDir, err := os.Getwd()
	if err != nil {
		return "", errors.Wrapf(err, "unable to get component because getting current directory failed")
	}
	retVal = filepath.Base(currDir)
	// Kubernetes resources require a name that satisfies DNS-1123
	retVal = strings.TrimSpace(util.GetDNS1123Name(strings.ToLower(retVal)))
	// Kubernetes resources have a characters limit, truncate and give sufficient space for other resources
	retVal = util.TruncateString(retVal, componentNameMaxLen)
	return retVal, nil
}
