package component

import (
	"fmt"

	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"

	"github.com/redhat-developer/odo/pkg/odo/util"
	ktemplates "k8s.io/kubectl/pkg/util/templates"

	"github.com/spf13/cobra"
)

// UnlinkRecommendedCommandName is the recommended unlink command name
const UnlinkRecommendedCommandName = "unlink"

var (
	unlinkExample = ktemplates.Examples(`# Unlink the 'my-postgresql' service from the current component 
%[1]s my-postgresql

# Unlink the 'my-postgresql' service  from the 'nodejs' component
%[1]s my-postgresql --component nodejs

# Unlink the 'backend' component from the current component (backend must have a single exposed port)
%[1]s backend

# Unlink the 'backend' service  from the 'nodejs' component
%[1]s backend --component nodejs

# Unlink the backend's 8080 port from the current component 
%[1]s backend --port 8080`)

	unlinkLongDesc = `Unlink component or service from a component. 
For this command to be successful, the service or component needs to have been linked prior to the invocation using 'odo link'`
)

// UnlinkOptions encapsulates the options for the odo link command
type UnlinkOptions struct {
	// Common link/unlink context
	*commonLinkOptions

	// Flags
	contextFlag string
}

// NewUnlinkOptions creates a new UnlinkOptions instance
func NewUnlinkOptions() *UnlinkOptions {
	options := UnlinkOptions{}
	options.commonLinkOptions = newCommonLinkOptions()
	return &options
}

// Complete completes UnlinkOptions after they've been created
func (o *UnlinkOptions) Complete(cmdline cmdline.Cmdline, args []string) (err error) {
	err = o.complete(cmdline, args, o.contextFlag)
	if err != nil {
		return err
	}

	if o.csvSupport {
		o.operation = o.KClient.UnlinkSecret
	}
	return err
}

// Validate validates the UnlinkOptions based on completed values
func (o *UnlinkOptions) Validate() (err error) {
	return o.validate()
}

// Run contains the logic for the odo link command
func (o *UnlinkOptions) Run() (err error) {
	return o.run()
}

// NewCmdUnlink implements the link odo command
func NewCmdUnlink(name, fullName string) *cobra.Command {
	o := NewUnlinkOptions()

	unlinkCmd := &cobra.Command{
		Use:         fmt.Sprintf("%s <service> --component [component] OR %s <component> --component [component]", name, name),
		Short:       "Unlink component to a service or component",
		Long:        unlinkLongDesc,
		Example:     fmt.Sprintf(unlinkExample, fullName),
		Args:        cobra.ExactArgs(1),
		Annotations: map[string]string{"command": "component"},
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	unlinkCmd.SetUsageTemplate(util.CmdUsageTemplate)
	//Adding `--component` flag
	AddComponentFlag(unlinkCmd)
	// Adding context flag
	odoutil.AddContextFlag(unlinkCmd, &o.contextFlag)

	return unlinkCmd
}
