package component

import (
	"fmt"
	"github.com/redhat-developer/odo/pkg/occlient"
	appCmd "github.com/redhat-developer/odo/pkg/odo/cli/application"
	projectCmd "github.com/redhat-developer/odo/pkg/odo/cli/project"
	"github.com/redhat-developer/odo/pkg/odo/util/completion"

	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/util"
	ktemplates "k8s.io/kubernetes/pkg/kubectl/cmd/templates"

	"github.com/spf13/cobra"
)

// RecommendedUnlinkCommandName is the recommended unlink command name
const RecommendedUnlinkCommandName = "unlink"

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
	port             string
	secretName       string
	isTargetAService bool
	*genericclioptions.Context
}

// "implement" the methods of CommonLinkOptions

func (o *UnlinkOptions) getSecretName() string {
	return o.secretName
}

func (o *UnlinkOptions) setSecretName(secretName string) {
	o.secretName = secretName
}

func (o *UnlinkOptions) getIsTargetAService() bool {
	return o.isTargetAService
}

func (o *UnlinkOptions) setIsTargetAService(isTargetAService bool) {
	o.isTargetAService = isTargetAService
}

func (o *UnlinkOptions) setContext(context *genericclioptions.Context) {
	o.Context = context
}

func (o *UnlinkOptions) getClient() *occlient.Client {
	return o.Client
}

func (o *UnlinkOptions) getApplication() string {
	return o.Application
}

func (o *UnlinkOptions) getProject() string {
	return o.Project
}

func (o *UnlinkOptions) getPort() string {
	return o.port
}

// NewUnlinkOptions creates a new LinkOptions instance
func NewUnlinkOptions() *UnlinkOptions {
	return &UnlinkOptions{}
}

// Run contains the logic for the odo link command
func (o *UnlinkOptions) Run(suppliedName string) (err error) {
	linkType := "Component"
	if o.isTargetAService {
		linkType = "Service"
	}

	err = o.Client.UnlinkSecret(o.secretName, o.Component(), o.Application)
	if err != nil {
		return err
	}

	log.Successf("%s %s has been successfully unlinked from the component %s", linkType, suppliedName, o.Component())
	return
}

// NewCmdUnlink implements the link odo command
func NewCmdUnlink(fullName string) *cobra.Command {
	o := NewUnlinkOptions()

	unlinkCmd := &cobra.Command{
		Use:     "unlink <service> --component [component] OR unlink <component> --component [component]",
		Short:   "Unlink component to a service or component",
		Long:    unlinkLongDesc,
		Example: fmt.Sprintf(unlinkExample, fullName),
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckError(Complete(o, cmd, args), "")
			util.CheckError(Validate(o, false), "")
			util.CheckError(o.Run(args[0]), "")
		},
	}

	unlinkCmd.PersistentFlags().StringVar(&o.port, "port", "", "Port of the backend to which to unlink")

	// Add a defined annotation in order to appear in the help menu
	unlinkCmd.Annotations = map[string]string{"command": "component"}
	unlinkCmd.SetUsageTemplate(util.CmdUsageTemplate)
	//Adding `--project` flag
	projectCmd.AddProjectFlag(unlinkCmd)
	//Adding `--application` flag
	appCmd.AddApplicationFlag(unlinkCmd)
	//Adding `--component` flag
	AddComponentFlag(unlinkCmd)

	completion.RegisterCommandHandler(unlinkCmd, completion.UnlinkCompletionHandler)

	return unlinkCmd
}
