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
	odoutil "github.com/openshift/odo/pkg/odo/util"

	"github.com/openshift/odo/pkg/devfile/adapters"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/util"
	"github.com/pkg/errors"
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
		return err
	}

	spinner := log.SpinnerNoSpin(fmt.Sprintf("Push devfile component %s", componentName))
	defer spinner.End(false)

	devfileHandler, err := adapters.NewPlatformAdapter(componentName, devObj)
	if err != nil {
		return err
	}
	err = component.ApplyConfig(nil, po.Context.KClient, config.LocalConfigInfo{}, *po.EnvSpecificInfo, color.Output, po.doesComponentExist)
	if err != nil {
		odoutil.LogErrorAndExit(err, "Failed to update config to component deployed.")
	}

	err = devfileHandler.Start()
	if err != nil {
		log.Errorf(
			"Failed to start component with name %s.\nError: %v",
			componentName,
			err,
		)
		os.Exit(1)
	}

	spinner.End(true)
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
	return retVal, nil
}
