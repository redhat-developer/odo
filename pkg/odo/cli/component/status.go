package component

import (
	"fmt"
	"time"

	"github.com/openshift/odo/pkg/devfile/adapters"
	"github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/devfile/adapters/kubernetes"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/machineoutput"
	"github.com/openshift/odo/pkg/occlient"
	appCmd "github.com/openshift/odo/pkg/odo/cli/application"
	projectCmd "github.com/openshift/odo/pkg/odo/cli/project"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/url"
	"github.com/pkg/errors"

	"github.com/openshift/odo/pkg/odo/util/completion"
	ktemplates "k8s.io/kubectl/pkg/util/templates"

	odoutil "github.com/openshift/odo/pkg/odo/util"

	"github.com/spf13/cobra"
)

// StatusRecommendedCommandName is the recommended watch command name
const StatusRecommendedCommandName = "status"

var statusExample = ktemplates.Examples(`  # Get the status for the nodejs component
%[1]s nodejs -o json --follow
`)

// StatusOptions contains status options
type StatusOptions struct {
	componentContext string

	componentName  string
	devfileHandler common.ComponentAdapter

	logFollow bool
	*genericclioptions.Context
}

// NewStatusOptions returns new instance of StatusOptions
func NewStatusOptions() *StatusOptions {
	return &StatusOptions{}
}

// Complete completes status args
func (so *StatusOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	so.Context, err = genericclioptions.New(genericclioptions.NewCreateParameters(cmd).NeedDevfile().SetComponentContext(so.componentContext))
	if err != nil {
		return err
	}

	// Get the component name
	so.componentName = so.EnvSpecificInfo.GetName()

	platformContext := kubernetes.KubernetesContext{
		Namespace: so.KClient.GetCurrentNamespace(),
	}

	so.devfileHandler, err = adapters.NewComponentAdapter(so.componentName, so.componentContext, so.GetApplication(), so.EnvSpecificInfo.GetDevfileObj(), platformContext)
	return err
}

// Validate validates the status parameters
func (so *StatusOptions) Validate() (err error) {

	if !so.logFollow {
		return fmt.Errorf("this command must be called with --follow")
	}

	return
}

// Run has the logic to perform the required actions as part of command
func (so *StatusOptions) Run(cmd *cobra.Command) (err error) {
	if !log.IsJSON() {
		return errors.New("this command only supports the '-o json' output format")
	}
	so.devfileHandler.StartSupervisordCtlStatusWatch()
	so.devfileHandler.StartContainerStatusWatch()

	loggingClient := machineoutput.NewConsoleMachineEventLoggingClient()

	// occlient is required so that we can report the status for route URLs (eg in addition to our already testing ingress URLs for k8s)
	oclient, err := occlient.New()
	if err != nil {
		// Fallback to k8s if occlient throws an error
		oclient = nil
	} else {
		oclient.Namespace = so.KClient.GetCurrentNamespace()
	}

	url.StartURLHttpRequestStatusWatchForK8S(oclient, so.KClient, &so.LocalConfigProvider, loggingClient)

	// You can call Run() any time you like, but you can never leave.
	for {
		time.Sleep(60 * time.Second)
	}

}

// NewCmdStatus implements the status odo command
func NewCmdStatus(name, fullName string) *cobra.Command {
	o := NewStatusOptions()

	annotations := map[string]string{"command": "component", "machineoutput": "json"}

	var statusCmd = &cobra.Command{
		Use:         fmt.Sprintf("%s [component_name]", name),
		Short:       "Watches the given component and outputs machine-readable JSON events representing component status changes",
		Long:        `Watches the given component and outputs machine-readable JSON events representing component status changes`,
		Example:     fmt.Sprintf(statusExample, fullName),
		Args:        cobra.MaximumNArgs(1),
		Annotations: annotations,
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	statusCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)

	// Adding context flag
	genericclioptions.AddContextFlag(statusCmd, &o.componentContext)

	statusCmd.Flags().BoolVarP(&o.logFollow, "follow", "f", false, "Follow the component and report all changes")

	//Adding `--application` flag
	appCmd.AddApplicationFlag(statusCmd)

	//Adding `--project` flag
	projectCmd.AddProjectFlag(statusCmd)

	completion.RegisterCommandHandler(statusCmd, completion.ComponentNameCompletionHandler)

	return statusCmd
}
