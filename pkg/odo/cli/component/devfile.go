package component

import (
	"github.com/openshift/odo/pkg/devfile"
	// "fmt"
	// "os"
	// componentDevfile "github.com/openshift/odo/pkg/component/devfile"
	// "github.com/openshift/odo/pkg/devfile"
	// "github.com/openshift/odo/pkg/log"
	// cli "github.com/openshift/odo/pkg/odo/cli/devfile"
	// "github.com/openshift/odo/pkg/odo/cli/project"
	// "github.com/openshift/odo/pkg/odo/genericclioptions"
	// "github.com/spf13/cobra"
	// ktemplates "k8s.io/kubernetes/pkg/kubectl/util/templates"
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
// PushDevfileOptions encapsulates odo component push-devfile  options
// type PushDevfileOptions struct {
// 	devfilePath string
// 	*cli.Context
// }

// // NewPushDevfileOptions returns new instance of PushDevfileOptions
// func NewPushDevfileOptions() *PushDevfileOptions {
// 	return &PushDevfileOptions{}
// }

// // Complete completes  args
// func (pdo *PushDevfileOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
// 	pdo.Context, err = cli.NewDevfileContext(cmd)
// 	return err
// }

// // Validate validates the  parameters
// func (pdo *PushDevfileOptions) Validate() (err error) {
// 	return nil
// }

// Run has the logic to perform the required actions as part of command
func (po *PushOptions) DevfilePush() (err error) {

	// Parse devfile
	_, err = devfile.Parse(po.devfilePath)
	if err != nil {
		return err
	}

	// componentName := pdo.Context.DevfileComponent.Name
	// spinner := log.Spinnerf("Push devfile component %s")
	// defer spinner.End(false)

	// devfileHandler, err := componentDevfile.NewPlatformAdapter(devObj, pdo.Context.DevfileComponent)
	// if err != nil {
	// 	return err
	// }

	// err = devfileHandler.Start()
	// if err != nil {
	// 	log.Errorf(
	// 		"Failed to start component with name %s.\nError: %v",
	// 		componentName,
	// 		err,
	// 	)
	// 	os.Exit(1)
	// }

	// spinner.End(true)
	return
}
