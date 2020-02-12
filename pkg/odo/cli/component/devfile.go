package component

import (
	"github.com/openshift/odo/pkg/devfile"
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
	devObj, err := devfile.Parse(po.devfilePath)
	if err != nil {
		return err
	}

	// Write back devfile yaml
	err = devObj.WriteYamlDevfile()
	if err != nil {
		return err
	}

	return nil
}
