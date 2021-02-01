package service

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/openshift/odo/pkg/odo/cli/component"
	"github.com/openshift/odo/pkg/util"

	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/odo/cli/service/ui"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/odo/util/completion"
	"github.com/openshift/odo/pkg/odo/util/validation"
	svc "github.com/openshift/odo/pkg/service"
	"github.com/spf13/cobra"

	ktemplates "k8s.io/kubectl/pkg/util/templates"
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
	%[1]s etcdoperator.v0.9.4/EtcdCluster`)

	createShortDesc = `Create a new service from Operator Hub or Service Catalog and deploy it on OpenShift.`

	createLongDesc = ktemplates.LongDesc(`
Create a new service from Operator Hub or Service Catalog and deploy it on OpenShift.

Service creation can be performed from a valid component directory (one containing a devfile.yaml) only.

To create the service from outside a component directory, specify path to a valid component directory using "--context" flag.

When creating a service using Operator Hub, provide a service name along with Operator name.

When creating a service using Service Catalog, a --plan must be passed along with the service type. Parameters to configure the service are passed as key=value pairs.

For a full list of service types, use: 'odo catalog list services'`)
)

// CreateOptions encapsulates the options for the odo service create command
type CreateOptions struct {
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
	// If set to true, DryRun prints the yaml that will create the service
	DryRun bool
	// Location of the file in which yaml specification of CR is stored.
	fromFile string
	// Backend is the service provider backend (Operator Hub or Service Catalog) providing the service requested by the user
	Backend ServiceProviderBackend
}

// NewCreateOptions creates a new CreateOptions instance
func NewCreateOptions() *CreateOptions {
	return &CreateOptions{}
}

// DynamicCRD holds the original CR obtained from the Operator (a CSV), or user
// (when they use --from-file flag), and few other attributes that are likely
// to be used to validate a CRD before creating a service from it
type DynamicCRD struct {
	// contains the CR as obtained from CSV or user
	OriginalCRD map[string]interface{}
}

func NewDynamicCRD() *DynamicCRD {
	return &DynamicCRD{}
}

// Complete completes CreateOptions after they've been created
func (o *CreateOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	o.Context, err = genericclioptions.New(genericclioptions.CreateParameters{
		Cmd:              cmd,
		DevfilePath:      component.DevfilePath,
		ComponentContext: o.componentContext,
	})
	if err != nil {
		return err
	}

	// check if service create is executed from a valid context because without that,
	// it's useless to execute further as we want to store service info in devfile
	if o.componentContext == "" {
		o.componentContext = component.LocalDirectoryDefaultLocation
	}
	devfilePath := filepath.Join(o.componentContext, component.DevfilePath)
	if !util.CheckPathExists(devfilePath) {
		return fmt.Errorf("service can be created from a valid component directory only\n"+
			"refer %q for more information", "odo servce create -h")
	}

	// decide which service backend to use
	if o.fromFile != "" {
		// fromFile is supported only for Operator backend
		o.Backend = NewOperatorBackend()
		// since interactive mode is not supported for Operators yet, set it to false
		o.interactive = false

		return o.Backend.CompleteServiceCreate(o, cmd, args)
	}
	_, _, err = svc.SplitServiceKindName(args[0])
	if err != nil {
		// failure to split provided name into two; hence ServiceCatalogBackend
		o.Backend = NewServiceCatalogBackend()
		err = nil
	} else {
		// provided name adheres to the format <operator-type>/<crd-name>; hence OperatorBackend
		o.Backend = NewOperatorBackend()
	}

	return o.Backend.CompleteServiceCreate(o, cmd, args)
}

// validateServiceName adopts the Validator interface and checks that the name of the service being created is valid
func (o *CreateOptions) validateServiceName(i interface{}) (err error) {
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
func (o *CreateOptions) outputNonInteractiveEquivalent() string {
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

// Validate validates the CreateOptions based on completed values
func (o *CreateOptions) Validate() (err error) {
	// if we are in interactive mode, all values are already valid
	if o.interactive {
		return nil
	}

	return o.Backend.ValidateServiceCreate(o)
}

// Run contains the logic for the odo service create command
func (o *CreateOptions) Run() (err error) {
	err = o.Backend.RunServiceCreate(o)
	if err != nil {
		return err
	}

	// Information on what to do next; don't do this if "--dry-run" was requested as it gets appended to the file
	if !o.DryRun {
		log.Infof("You can now link the service to a component using 'odo link'; check 'odo link -h'")
	}

	equivalent := o.outputNonInteractiveEquivalent()
	if len(equivalent) > 0 {
		log.Info("Equivalent command:\n" + ui.StyledOutput(equivalent, "cyan"))
	}
	return
}

// NewCmdServiceCreate implements the odo service create command.
func NewCmdServiceCreate(name, fullName string) *cobra.Command {
	o := NewCreateOptions()
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

	serviceCreateCmd.Use += fmt.Sprintf(" [flags]\n  %s <operator_type>/<crd_name> [service_name] [flags]", o.CmdFullName)
	serviceCreateCmd.Example += "\n\n" + fmt.Sprintf(createOperatorExample, fullName)
	serviceCreateCmd.Flags().BoolVar(&o.DryRun, "dry-run", false, "Print the yaml specificiation that will be used to create the operator backed service")
	// remove this feature after enabling service create interactive mode for operator backed services
	serviceCreateCmd.Flags().StringVar(&o.fromFile, "from-file", "", "Path to the file containing yaml specification to use to start operator backed service")

	serviceCreateCmd.Flags().StringVar(&o.Plan, "plan", "", "The name of the plan of the service to be created")
	serviceCreateCmd.Flags().StringArrayVarP(&o.parameters, "parameters", "p", []string{}, "Parameters of the plan where a parameter is expressed as <key>=<value")
	serviceCreateCmd.Flags().BoolVarP(&o.wait, "wait", "w", false, "Wait until the service is ready")
	genericclioptions.AddContextFlag(serviceCreateCmd, &o.componentContext)
	completion.RegisterCommandHandler(serviceCreateCmd, completion.ServiceClassCompletionHandler)
	completion.RegisterCommandFlagHandler(serviceCreateCmd, "plan", completion.ServicePlanCompletionHandler)
	completion.RegisterCommandFlagHandler(serviceCreateCmd, "parameters", completion.ServiceParameterCompletionHandler)
	return serviceCreateCmd
}

// validateMetadataInCRD validates if the CRD has metadata.name field and returns an error
func (d *DynamicCRD) validateMetadataInCRD() error {
	metadata, ok := d.OriginalCRD["metadata"].(map[string]interface{})
	if !ok {
		// this condition is satisfied if there's no metadata at all in the provided CRD
		return fmt.Errorf("couldn't find \"metadata\" in the yaml; need metadata start the service")
	}

	if _, ok := metadata["name"].(string); ok {
		// found the metadata.name; no error
		return nil
	}
	return fmt.Errorf("couldn't find metadata.name in the yaml; provide a name for the service")
}

// setServiceName modifies the CRD to contain user provided name on the CLI
// instead of using the default one in almExample
func (d *DynamicCRD) setServiceName(name string) {
	metaMap := d.OriginalCRD["metadata"].(map[string]interface{})

	for k := range metaMap {
		if k == "name" {
			metaMap[k] = name
			return
		}
		// if metadata doesn't have 'name' field, we set it up
		metaMap["name"] = name
	}
}

// getServiceNameFromCRD fetches the service name from metadata.name field of the CRD
func (d *DynamicCRD) getServiceNameFromCRD() (string, error) {
	metadata, ok := d.OriginalCRD["metadata"].(map[string]interface{})
	if !ok {
		// this condition is satisfied if there's no metadata at all in the provided CRD
		return "", fmt.Errorf("couldn't find \"metadata\" in the yaml; need metadata.name to start the service")
	}

	if name, ok := metadata["name"].(string); ok {
		// found the metadata.name; no error
		return name, nil
	}
	return "", fmt.Errorf("couldn't find metadata.name in the yaml; provide a name for the service")
}
