package component

import (
	"path/filepath"
	"time"

	"github.com/pkg/errors"

	"fmt"

	"github.com/openshift/odo/pkg/devfile"
	"github.com/openshift/odo/pkg/devfile/adapters"
	"github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/devfile/adapters/kubernetes"
	"github.com/openshift/odo/pkg/devfile/parser"
	"github.com/openshift/odo/pkg/envinfo"
	"github.com/openshift/odo/pkg/kclient/generator"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/machineoutput"
	"github.com/openshift/odo/pkg/occlient"
	appCmd "github.com/openshift/odo/pkg/odo/cli/application"
	projectCmd "github.com/openshift/odo/pkg/odo/cli/project"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/url"
	"github.com/openshift/odo/pkg/util"

	"github.com/openshift/odo/pkg/odo/util/completion"
	"github.com/openshift/odo/pkg/odo/util/pushtarget"
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
	devfilePath    string
	namespace      string
	devfileHandler common.ComponentAdapter

	devObj parser.DevfileObj

	logFollow       bool
	EnvSpecificInfo *envinfo.EnvSpecificInfo
	*genericclioptions.Context
	isDevfile bool
}

// NewStatusOptions returns new instance of StatusOptions
func NewStatusOptions() *StatusOptions {
	return &StatusOptions{}
}

// Complete completes status args
func (so *StatusOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	so.devfilePath = filepath.Join(so.componentContext, DevfilePath)

	so.isDevfile = util.CheckPathExists(so.devfilePath)

	// If devfile is present
	if so.isDevfile {
		envinfo, err := envinfo.NewEnvSpecificInfo(so.componentContext)
		if err != nil {
			return errors.Wrap(err, "unable to retrieve configuration information")
		}
		so.EnvSpecificInfo = envinfo
		so.Context = genericclioptions.NewDevfileContext(cmd)

		// Get the component name
		so.componentName = so.EnvSpecificInfo.GetName()
		if err != nil {
			return err
		}

		// Parse devfile
		devObj, err := devfile.ParseAndValidate(so.devfilePath)
		if err != nil {
			return err
		}
		so.devObj = devObj

		var platformContext interface{}
		if !pushtarget.IsPushTargetDocker() {
			// The namespace was retrieved from the --project flag (or from the kube client if not set) and stored in kclient when initializing the context
			so.namespace = so.KClient.Namespace
			platformContext = kubernetes.KubernetesContext{
				Namespace: so.namespace,
			}
		} else {
			platformContext = nil
		}
		so.devfileHandler, err = adapters.NewComponentAdapter(so.componentName, so.componentContext, so.Application, devObj, platformContext)

		return err
	}

	// Set the correct context
	so.Context = genericclioptions.NewContextCreatingAppIfNeeded(cmd)

	return
}

// Validate validates the status parameters
func (so *StatusOptions) Validate() (err error) {

	if !so.logFollow {
		return fmt.Errorf("this command must be called with --follow")
	}

	return
}

// Run has the logic to perform the required actions as part of command
func (so *StatusOptions) Run() (err error) {

	if !so.isDevfile {
		return errors.New("the status command is only supported for devfiles")
	}

	if !log.IsJSON() {
		return errors.New("this command only supports the '-o json' output format")
	}
	so.devfileHandler.StartSupervisordCtlStatusWatch()
	so.devfileHandler.StartContainerStatusWatch()

	loggingClient := machineoutput.NewConsoleMachineEventLoggingClient()

	if pushtarget.IsPushTargetDocker() {
		url.StartURLHttpRequestStatusWatchForDocker(so.EnvSpecificInfo, loggingClient)
	} else {

		// occlient is required so that we can report the status for route URLs (eg in addition to our already testing ingress URLs for k8s)
		oclient, err := occlient.New()
		if err != nil {
			// Fallback to k8s if occlient throws an error
			oclient = nil
		} else {
			oclient.Namespace = so.KClient.Namespace
		}

		containerComponents := generator.GetDevfileContainerComponents(so.devObj.Data)
		url.StartURLHttpRequestStatusWatchForK8S(oclient, so.KClient, so.EnvSpecificInfo, loggingClient, containerComponents)
	}

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
