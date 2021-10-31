package url

import (
	"fmt"
	"strings"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/openshift/odo/pkg/localConfigProvider"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/machineoutput"
	clicomponent "github.com/openshift/odo/pkg/odo/cli/component"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/odo/util/completion"
	"github.com/openshift/odo/pkg/url"
	"github.com/pkg/errors"

	"github.com/openshift/odo/pkg/util"
	"github.com/spf13/cobra"

	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

const createRecommendedCommandName = "create"

var (
	urlCreateShortDesc = `Create a URL for a component`
	urlCreateLongDesc  = ktemplates.LongDesc(`Create a URL for a component.
	The created URL can be used to access the specified component from outside the cluster.
	`)
	urlCreateExample = ktemplates.Examples(`  # Create a URL with a specific name by automatically detecting the port used by the component
	%[1]s example

	# Create a URL for the current component with a specific port
	%[1]s --port 8080

	# Create a URL with a specific name and port
	%[1]s example --port 8080

	# Create a URL of ingress kind for the current component with a host
	%[1]s --port 8080 --host example.com --ingress

	# Create a secure URL for the current component
	%[1]s --port 8080 --secure

	# Create a URL with a specific path and protocol type
	%[1]s --port 8080 --path /hello --protocol http

	# Create a URL under a specific container
	%[1]s --port 8080 --container runtime
	  `)
)

// CreateOptions encapsulates the options for the odo url create command
type CreateOptions struct {
	*clicomponent.PushOptions
	urlName     string
	urlPort     int
	secureURL   bool
	now         bool
	host        string // host of the URL
	tlsSecret   string // tlsSecret is the secret to te used by the URL
	path        string // path of the URL
	protocol    string // protocol of the URL
	container   string // container to which the URL belongs
	wantIngress bool
	url         localConfigProvider.LocalURL
}

// NewURLCreateOptions creates a new CreateOptions instance
func NewURLCreateOptions() *CreateOptions {
	return &CreateOptions{PushOptions: clicomponent.NewPushOptions()}
}

// Complete completes CreateOptions after they've been Created
func (o *CreateOptions) Complete(_ string, cmd *cobra.Command, args []string) (err error) {
	params := genericclioptions.NewCreateParameters(cmd).NeedDevfile().SetComponentContext(o.GetComponentContext()).CheckRouteAvailability()
	if o.now {
		params.CreateAppIfNeeded()
	}
	o.Context, err = genericclioptions.New(params)
	if err != nil {
		return err
	}

	var urlType localConfigProvider.URLKind
	if o.wantIngress {
		urlType = localConfigProvider.INGRESS
	}

	// get the name
	if len(args) != 0 {
		o.urlName = args[0]
	}

	// create the localURL
	o.url = localConfigProvider.LocalURL{
		Name:      o.urlName,
		Port:      o.urlPort,
		Secure:    o.secureURL,
		Host:      o.host,
		TLSSecret: o.tlsSecret,
		Kind:      urlType,
		Container: o.container,
		Protocol:  o.protocol,
		Path:      o.path,
	}

	// complete the URL
	err = o.Context.LocalConfigProvider.CompleteURL(&o.url)
	if err != nil {
		return err
	}

	if o.now {
		prjName := o.Context.LocalConfigProvider.GetNamespace()
		o.ResolveSrcAndConfigFlags()
		err = o.ResolveProject(prjName)
		if err != nil {
			return err
		}
	}

	return
}

// Validate validates the CreateOptions based on completed values
func (o *CreateOptions) Validate() (err error) {
	if !util.CheckOutputFlag(o.GetOutputFlag()) {
		return fmt.Errorf("given output format %s is not supported", o.GetOutputFlag())
	}

	errorList := make([]string, 0)
	// Check if url name is more than 63 characters long
	if len(o.urlName) > 63 {
		errorList = append(errorList, "URL name must be shorter than 63 characters")
	}

	// validate the URL
	err = o.LocalConfigProvider.ValidateURL(o.url)
	if err != nil {
		errorList = append(errorList, err.Error())
	}

	if len(errorList) > 0 {
		for i := range errorList {
			errorList[i] = fmt.Sprintf("\t- %s", errorList[i])
		}
		return fmt.Errorf("URL creation failed:\n%s", strings.Join(errorList, "\n"))
	}
	return
}

// Run contains the logic for the odo url create command
func (o *CreateOptions) Run(cmd *cobra.Command) (err error) {

	// create the URL and write it to the local config
	err = o.Context.LocalConfigProvider.CreateURL(o.url)
	if err != nil {
		return err
	}
	log.Successf("URL %s created for component: %v", o.url.Name, o.LocalConfigProvider.GetName())

	if o.now {
		// if the now flag is specified, push the changes
		o.CompleteDevfilePath()
		o.EnvSpecificInfo = o.Context.EnvSpecificInfo
		err = o.DevfilePush()
		if err != nil {
			return errors.Wrap(err, "failed to push changes")
		}
	} else {
		log.Italic("\nTo apply the URL configuration changes, please use `odo push`")
	}

	if log.IsJSON() {
		u := url.NewURLFromLocalURL(o.url)
		u.Status.State = url.StateTypeNotPushed
		if o.now {
			u.Status.State = url.StateTypePushed
		}
		machineoutput.OutputSuccess(u)
	}
	return
}

// NewCmdURLCreate implements the odo url create command.
func NewCmdURLCreate(name, fullName string) *cobra.Command {
	o := NewURLCreateOptions()
	urlCreateCmd := &cobra.Command{
		Use:         name + " [url name]",
		Short:       urlCreateShortDesc,
		Long:        urlCreateLongDesc,
		Example:     fmt.Sprintf(urlCreateExample, fullName),
		Args:        cobra.MaximumNArgs(1),
		Annotations: map[string]string{"machineoutput": "json"},
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}
	urlCreateCmd.Flags().IntVarP(&o.urlPort, "port", "", -1, "Port number for the url of the component, required in case of components which expose more than one service port")
	urlCreateCmd.Flags().StringVar(&o.tlsSecret, "tls-secret", "", "TLS secret name for the url of the component if the user bring their own TLS secret")
	urlCreateCmd.Flags().StringVarP(&o.host, "host", "", "", "Cluster IP for this URL")
	urlCreateCmd.Flags().BoolVar(&o.wantIngress, "ingress", false, "Create an Ingress instead of Route on OpenShift clusters")
	urlCreateCmd.Flags().BoolVarP(&o.secureURL, "secure", "", false, "Create a secure HTTPS URL")
	urlCreateCmd.Flags().StringVarP(&o.path, "path", "", "", "path for this URL")
	urlCreateCmd.Flags().StringVarP(&o.protocol, "protocol", "", string(devfilev1.HTTPEndpointProtocol), "protocol for this URL")
	urlCreateCmd.Flags().StringVarP(&o.container, "container", "", "", "container of the endpoint in devfile")

	genericclioptions.AddNowFlag(urlCreateCmd, &o.now)
	o.AddContextFlag(urlCreateCmd)
	completion.RegisterCommandFlagHandler(urlCreateCmd, "context", completion.FileCompletionHandler)

	return urlCreateCmd
}
