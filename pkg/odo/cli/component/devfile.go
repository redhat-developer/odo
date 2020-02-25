package component

import (
	"github.com/fatih/color"
	"github.com/openshift/odo/pkg/component"
	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/devfile"
	odoutil "github.com/openshift/odo/pkg/odo/util"
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

// Run has the logic to perform the required actions as part of command
func (po *PushOptions) DevfilePush() (err error) {

	// Parse devfile
	devObj, err := devfile.Parse(po.DevfilePath)
	if err != nil {
		return err
	}

	// Write back devfile yaml
	err = devObj.WriteYamlDevfile()
	if err != nil {
		return err
	}
	err = component.ApplyConfig(nil, po.Context.KClient, config.LocalConfigInfo{}, *po.EnvSpecificInfo, color.Output, po.doesComponentExist)
	if err != nil {
		odoutil.LogErrorAndExit(err, "Failed to update config to component deployed.")
	}

	return nil
}
