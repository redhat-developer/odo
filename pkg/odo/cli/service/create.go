package service

import (
	"fmt"

	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

const (
	createRecommendedCommandName = "create"
)

var (
	createOperatorExample = ktemplates.Examples(`
	# Create new EtcdCluster service from etcdoperator.v0.9.4 operator.
	%[1]s etcdoperator.v0.9.4/EtcdCluster
	
	# Create new EtcdCluster service from etcdoperator.v0.9.4 operator and puts the service definition in the devfile instead of a separate file.
	%[1]s etcdoperator.v0.9.4/EtcdCluster --inlined`)

	createShortDesc = `Create a new service from Operator Hub and deploy it on Kubernetes or OpenShift.`

	createLongDesc = ktemplates.LongDesc(`
Create a new service from Operator Hub and deploy it on Kubernetes or OpenShift.

Service creation can be performed from a valid component directory (one containing a devfile.yaml) only.

To create the service from outside a component directory, specify path to a valid component directory using "--context" flag.

When creating a service using Operator Hub, provide a service name along with Operator name.

For a full list of service types, use: 'odo catalog list services'`)
)

// CreateOptions encapsulates the options for the odo service create command
type CreateOptions struct {
	// Context
	*genericclioptions.Context

	// Flags
	parametersFlag []string
	waitFlag       bool
	contextFlag    string
	DryRunFlag     bool
	fromFileFlag   string
	inlinedFlag    bool

	// ServiceType corresponds to the service class name
	ServiceType string
	// ServiceName is how the service will be named and known by odo
	ServiceName string
	// ParametersMap is populated from the flag-provided values (parameters) and/or the interactive mode and is the expected format by the business logic
	ParametersMap map[string]string
	// interactive specifies whether the command operates in interactive mode or not
	interactive bool
	// CmdFullName records the command's full name
	CmdFullName string
	// Backend is the service provider backend providing the service requested by the user
	Backend ServiceProviderBackend
}

// NewCreateOptions creates a new CreateOptions instance
func NewCreateOptions() *CreateOptions {
	return &CreateOptions{}
}

// Complete completes CreateOptions after they've been created
func (o *CreateOptions) Complete(name string, cmdline cmdline.Cmdline, args []string) (err error) {
	cmd := cmdline.GetCmd()
	o.Context, err = genericclioptions.New(genericclioptions.NewCreateParameters(cmdline).NeedDevfile(o.contextFlag))
	if err != nil {
		return err
	}
	// we convert the param list provided in the format of key=value list
	// to a map
	o.ParametersMap, err = util.MapFromParameters(o.parametersFlag)
	if err != nil {
		return err
	}

	err = validDevfileDirectory(o.contextFlag)
	if err != nil {
		return err
	}
	//if no args are provided and if request is not from file, user wants interactive mode
	if o.fromFileFlag == "" && len(args) == 0 {
		return fmt.Errorf("odo doesn't support interactive mode for creating Operator backed service")
	}
	o.Backend = NewOperatorBackend()
	o.interactive = false
	return o.Backend.CompleteServiceCreate(o, cmd, args)
}

// Validate validates the CreateOptions based on completed values
func (o *CreateOptions) Validate() (err error) {
	return o.Backend.ValidateServiceCreate(o)
}

// Run contains the logic for the odo service create command
func (o *CreateOptions) Run() (err error) {
	err = o.Backend.RunServiceCreate(o)
	if err != nil {
		return fmt.Errorf("service %q already exists in configuration", o.ServiceName)
	}

	// Information on what to do next; don't do this if "--dry-run" was requested as it gets appended to the file
	if !o.DryRunFlag {
		log.Info("Successfully added service to the configuration; do 'odo push' to create service on the cluster")
	}

	return nil
}

// NewCmdServiceCreate implements the odo service create command.
func NewCmdServiceCreate(name, fullName string) *cobra.Command {
	o := NewCreateOptions()
	o.CmdFullName = fullName
	serviceCreateCmd := &cobra.Command{
		Use:         name + " <operator_type>/<crd_name> [service_name] [flags]",
		Short:       createShortDesc,
		Long:        createLongDesc,
		Example:     fmt.Sprintf(createOperatorExample, fullName),
		Args:        cobra.RangeArgs(0, 2),
		Annotations: map[string]string{"machineoutput": "json"},
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	serviceCreateCmd.Flags().BoolVar(&o.inlinedFlag, "inlined", false, "Puts the service definition in the devfile instead of a separate file")
	serviceCreateCmd.Flags().BoolVar(&o.DryRunFlag, "dry-run", false, "Print the yaml specificiation that will be used to create the operator backed service")
	// remove this feature after enabling service create interactive mode for operator backed services
	serviceCreateCmd.Flags().StringVar(&o.fromFileFlag, "from-file", "", "Path to the file containing yaml specification to use to start operator backed service")

	serviceCreateCmd.Flags().StringArrayVarP(&o.parametersFlag, "parameters", "p", []string{}, "Parameters to be used to create Operator backed service where a parameter is expressed as <key>=<value")
	serviceCreateCmd.Flags().BoolVarP(&o.waitFlag, "wait", "w", false, "Wait until the service is ready")
	genericclioptions.AddContextFlag(serviceCreateCmd, &o.contextFlag)
	return serviceCreateCmd
}
