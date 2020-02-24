package service

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	commonui "github.com/openshift/odo/pkg/odo/cli/ui"
	"github.com/openshift/odo/pkg/odo/util/validation"
	"github.com/pkg/errors"

	"github.com/golang/glog"
	scv1beta1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/odo/cli/service/ui"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/odo/util/completion"
	"github.com/openshift/odo/pkg/odo/util/experimental"
	svc "github.com/openshift/odo/pkg/service"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubernetes/pkg/kubectl/util/templates"
)

const (
	createRecommendedCommandName = "create"
	equivalentTemplate           = "{{.CmdFullName}} {{.ServiceType}}" +
		"{{if .ServiceName}} {{.ServiceName}}{{end}}" +
		" --app {{.Application}}" +
		" --project {{.Project}}" +
		"{{if .Plan}} --plan {{.Plan}}{{end}}" +
		"{{range $key, $value := .ParametersMap}} -p {{$key}}={{$value}}{{end}}"
)

var (
	createExample = ktemplates.Examples(`
    # Create new postgresql service from service catalog using dev plan and name my-postgresql-db.
    %[1]s dh-postgresql-apb my-postgresql-db --plan dev -p postgresql_user=luke -p postgresql_password=secret`)

	createOperatorExample = ktemplates.Examples(`
	# Create new EtcdCluster service from etcdoperator.v0.9.4 operator.
	%[1]s etcdoperator.v0.9.4 --crd EtcdCluster`)

	createShortDesc = `Create a new service from service catalog using the plan defined and deploy it on OpenShift.`

	createLongDesc = ktemplates.LongDesc(`
Create a new service from service catalog using the plan defined and deploy it on OpenShift.

A --plan must be passed along with the service type. Parameters to configure the service are passed as key=value pairs.

For a full list of service types, use: 'odo catalog list services'`)
)

// ServiceCreateOptions encapsulates the options for the odo service create command
type ServiceCreateOptions struct {
	// parameters hold the user-provided values for service class parameters via flags (populated by cobra)
	parameters []string
	// Plan is the selected service plan
	Plan string
	// ServiceType corresponds to the service class name
	ServiceType string
	// ServiceName is how the service will be named and known by odo
	ServiceName string
	// ParametersMap is populated from the flag-provided values (parameters) and/or the interactive mode and is the expected format by the business logic
	ParametersMap map[string]string
	// interactive specifies whether the command operates in interactive mode or not
	interactive bool
	// outputCLI specifies whether to output the non-interactive version of the command or not
	outputCLI bool
	// CmdFullName records the command's full name
	CmdFullName string
	// whether or not to wait for the service to be ready
	wait bool
	// generic context options common to all commands
	*genericclioptions.Context
	// Context to use when creating service. This will use app and project values from the context
	componentContext string
	// CRD to use when creating an operator backed service
	Crd string
}

// NewServiceCreateOptions creates a new ServiceCreateOptions instance
func NewServiceCreateOptions() *ServiceCreateOptions {
	return &ServiceCreateOptions{}
}

// Complete completes ServiceCreateOptions after they've been created
func (o *ServiceCreateOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	if len(args) == 0 || !cmd.HasFlags() {
		o.interactive = true
	}

	if o.componentContext != "" {
		o.Context = genericclioptions.NewContext(cmd)
	} else {
		o.Context = genericclioptions.NewContextCreatingAppIfNeeded(cmd)
	}

	client := o.Client

	var class scv1beta1.ClusterServiceClass
	if o.interactive {
		classesByCategory, err := client.GetServiceClassesByCategory()
		if err != nil {
			return fmt.Errorf("unable to retrieve service classes: %v", err)
		}

		if len(classesByCategory) == 0 {
			return fmt.Errorf("no available service classes")
		}

		class, o.ServiceType = ui.SelectClassInteractively(classesByCategory)

		plans, err := client.GetMatchingPlans(class)
		if err != nil {
			return fmt.Errorf("couldn't retrieve plans for class %s: %v", class.GetExternalName(), err)
		}

		var svcPlan scv1beta1.ClusterServicePlan
		// if there is only one available plan, we select it
		if len(plans) == 1 {
			for k, v := range plans {
				o.Plan = k
				svcPlan = v
			}
			glog.V(4).Infof("Plan %s was automatically selected since it's the only one available for service %s", o.Plan, o.ServiceType)
		} else {
			// otherwise select the plan interactively
			o.Plan = ui.SelectPlanNameInteractively(plans, "Which service plan should we use ")
			svcPlan = plans[o.Plan]
		}

		o.ParametersMap = ui.EnterServicePropertiesInteractively(svcPlan)
		o.ServiceName = ui.EnterServiceNameInteractively(o.ServiceType, "How should we name your service ", o.validateServiceName)
		o.outputCLI = commonui.Proceed("Output the non-interactive version of the selected options")
		o.wait = commonui.Proceed("Wait for the service to be ready")
	} else {
		o.ServiceType = args[0]
		// if only one arg is given, then it is considered as service name and service type both
		o.ServiceName = o.ServiceType
		// if two args are given, first is service type and second one is service name
		if len(args) == 2 {
			o.ServiceName = args[1]
		}

		// we convert the param list provided in the format of key=value list
		// to a map
		o.ParametersMap = make(map[string]string)
		for _, kv := range o.parameters {
			kvSlice := strings.Split(kv, "=")
			// key value not provided in format of key=value
			if len(kvSlice) != 2 {
				return errors.New("parameters not provided in key=value format")
			}
			o.ParametersMap[kvSlice[0]] = kvSlice[1]
		}
	}

	return
}

// validateServiceName adopts the Validator interface and checks that the name of the service being created is valid
func (o *ServiceCreateOptions) validateServiceName(i interface{}) (err error) {
	s := i.(string)
	err = validation.ValidateName(s)
	if err != nil {
		return err
	}
	exists, err := svc.SvcExists(o.Client, s, o.Application)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("%s service already exists in the current application", o.ServiceName)
	}
	return
}

// outputNonInteractiveEquivalent outputs the populated options as the equivalent command that would be used in non-interactive mode
func (o *ServiceCreateOptions) outputNonInteractiveEquivalent() string {
	if o.outputCLI {
		var tpl bytes.Buffer
		t := template.Must(template.New("service-create-cli").Parse(equivalentTemplate))
		e := t.Execute(&tpl, o)
		if e != nil {
			panic(e) // shouldn't happen
		}
		return strings.TrimSpace(tpl.String())
	}
	return ""
}

// Validate validates the ServiceCreateOptions based on completed values
func (o *ServiceCreateOptions) Validate() (err error) {
	// if we are in interactive mode, all values are already valid
	if o.interactive {
		return nil
	}

	// make sure the service type exists
	classPtr, err := o.Client.GetClusterServiceClass(o.ServiceType)
	if err != nil {
		return errors.Wrap(err, "unable to create service because Service Catalog is not enabled in your cluster")
	}
	if classPtr == nil {
		return fmt.Errorf("service %v doesn't exist\nRun 'odo catalog list services' to see a list of supported services.\n", o.ServiceType)
	}

	// check plan
	plans, err := o.Client.GetMatchingPlans(*classPtr)
	if err != nil {
		return err
	}
	if len(o.Plan) == 0 {
		// when the plan has not been supplied, if there is only one available plan, we select it
		if len(plans) == 1 {
			for k := range plans {
				o.Plan = k
			}
			glog.V(4).Infof("Plan %s was automatically selected since it's the only one available for service %s", o.Plan, o.ServiceType)
		} else {
			return fmt.Errorf("no plan was supplied for service %v.\nPlease select one of: %v\n", o.ServiceType, strings.Join(ui.GetServicePlanNames(plans), ","))
		}
	} else {
		// when the plan has been supplied, we need to make sure it exists
		if _, ok := plans[o.Plan]; !ok {
			return fmt.Errorf("plan %s is invalid for service %v.\nPlease select one of: %v\n", o.Plan, o.ServiceType, strings.Join(ui.GetServicePlanNames(plans), ","))
		}
	}
	//validate service name
	return o.validateServiceName(o.ServiceName)
}

// Run contains the logic for the odo service create command
func (o *ServiceCreateOptions) Run() (err error) {
	log.Infof("Deploying service %s of type: %s", o.ServiceName, o.ServiceType)

	s := log.Spinner("Deploying service")
	defer s.End(false)
	err = svc.CreateService(o.Client, o.ServiceName, o.ServiceType, o.Plan, o.ParametersMap, o.Application)
	if err != nil {
		return err
	}
	s.End(true)

	if o.wait {
		s = log.Spinner("Waiting for service to come up")
		_, err = o.Client.WaitAndGetSecret(o.ServiceName, o.Project)
		if err == nil {
			s.End(true)
			log.Successf(`Service '%s' is ready for use`, o.ServiceName)
		}
	} else {
		log.Successf(`Service '%s' was created`, o.ServiceName)
		log.Italic("\nProgress of the provisioning will not be reported and might take a long time\nYou can see the current status by executing 'odo service list'")
	}

	// Information on what to do next
	log.Infof("Optionally, link %s to your component by running: 'odo link <component-name>'", o.ServiceType)

	equivalent := o.outputNonInteractiveEquivalent()
	if len(equivalent) > 0 {
		log.Info("Equivalent command:\n" + ui.StyledOutput(equivalent, "cyan"))
	}
	return
}

// NewCmdServiceCreate implements the odo service create command.
func NewCmdServiceCreate(name, fullName string) *cobra.Command {
	o := NewServiceCreateOptions()
	o.CmdFullName = fullName
	serviceCreateCmd := &cobra.Command{
		Use:     name + " <service_type> --plan <plan_name> [service_name]",
		Short:   createShortDesc,
		Long:    createLongDesc,
		Example: fmt.Sprintf(createExample, fullName),
		Args:    cobra.RangeArgs(0, 2),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	if experimental.IsExperimentalModeEnabled() {
		serviceCreateCmd.Use += fmt.Sprintf(" [flags]\n  %s <operator_type> --crd <crd_name> [service_name] [flags]", o.CmdFullName)
		serviceCreateCmd.Example += fmt.Sprintf("\n\n") + fmt.Sprintf(createOperatorExample, fullName)
	}

	serviceCreateCmd.Flags().StringVar(&o.Plan, "plan", "", "The name of the plan of the service to be created")
	if experimental.IsExperimentalModeEnabled() {
		serviceCreateCmd.Flags().StringVar(&o.Crd, "crd", "", "The name of the CRD of the operator to be used to create the service")
	}
	serviceCreateCmd.Flags().StringArrayVarP(&o.parameters, "parameters", "p", []string{}, "Parameters of the plan where a parameter is expressed as <key>=<value")
	serviceCreateCmd.Flags().BoolVarP(&o.wait, "wait", "w", false, "Wait until the service is ready")
	genericclioptions.AddContextFlag(serviceCreateCmd, &o.componentContext)
	completion.RegisterCommandHandler(serviceCreateCmd, completion.ServiceClassCompletionHandler)
	completion.RegisterCommandFlagHandler(serviceCreateCmd, "plan", completion.ServicePlanCompletionHandler)
	completion.RegisterCommandFlagHandler(serviceCreateCmd, "parameters", completion.ServiceParameterCompletionHandler)
	return serviceCreateCmd
}
