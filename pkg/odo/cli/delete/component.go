package delete

import (
	"fmt"
	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/pkg/errors"
	"github.com/redhat-developer/odo/pkg/libdevfile"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/cli/ui"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	"github.com/spf13/cobra"
	"path/filepath"
)

// ComponentRecommendedCommandName is the recommended component sub-command name
const ComponentRecommendedCommandName = "component"

type ComponentOptions struct {
	// name of the component to delete, optional
	name string

	// forceFlag forces deletion
	forceFlag bool

	// Context
	*genericclioptions.Context

	// Clients
	clientset *clientset.Clientset
}

// NewComponentOptions returns new instance of ComponentOptions
func NewComponentOptions() *ComponentOptions {
	return &ComponentOptions{}
}

func (o *ComponentOptions) SetClientset(clientset *clientset.Clientset) {
	o.clientset = clientset
}

func (o *ComponentOptions) Complete(cmdline cmdline.Cmdline, args []string) (err error) {
	o.Context, err = genericclioptions.New(genericclioptions.NewCreateParameters(cmdline).NeedDevfile(""))
	return err
}

func (o *ComponentOptions) Validate() (err error) {
	return nil
}

func (o *ComponentOptions) Run() error {
	if o.name != "" {
		return o.deleteNamedComponent()
	}
	return o.deleteDevfileComponent()
}

// deleteNamedComponent deletes a component given its name
func (o *ComponentOptions) deleteNamedComponent() error {
	return nil
}

// deleteDevfileComponent deletes a component defined by the devfile in the current directory
func (o *ComponentOptions) deleteDevfileComponent() error {
	// TODO: Print all the resources that will be deleted.

	componentName := o.EnvSpecificInfo.GetName()

	if o.forceFlag || ui.Proceed(fmt.Sprintf("Are you sure you want to delete the devfile component: %s?", componentName)) {
		// delete outerloop resources
		devfileObj := o.EnvSpecificInfo.GetDevfileObj()
		err := o.clientset.DeleteClient.UnDeploy(devfileObj, filepath.Dir(o.EnvSpecificInfo.GetDevfilePath()))
		if err != nil {
			// if there is no component in the devfile to undeploy or if the devfile is non-existent, then skip the error log
			if !errors.Is(err, libdevfile.NewNoCommandFoundError(v1alpha2.DeployCommandGroupKind)) {
				log.Errorf("error occurred while un-deploying, cause: %v", err)
			}
		}
		// delete innerloop resources
		err = o.clientset.DeleteClient.DeleteComponent(devfileObj, componentName)
		if err != nil {
			log.Errorf("error occurred while deleting component, cause: %v", err)
		}
	} else {
		log.Error("Aborting deletion of component")
	}

	return nil
}

// NewCmdComponent implements the component odo sub-command
func NewCmdComponent(name, fullName string) *cobra.Command {
	o := NewComponentOptions()

	var componentCmd = &cobra.Command{
		Use:   name,
		Short: "Delete component",
		Long:  "Delete component",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}
	componentCmd.Flags().StringVar(&o.name, "name", "", "Name of the component to delete, optional. By default, the component described in the local devfile is deleted")
	componentCmd.Flags().BoolVarP(&o.forceFlag, "force", "f", false, "Delete component without prompting")
	clientset.Add(componentCmd, clientset.DELETE)

	return componentCmd
}
