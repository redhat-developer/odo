package url

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	dfutil "github.com/devfile/library/pkg/util"

	"github.com/redhat-developer/odo/pkg/localConfigProvider"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/machineoutput"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/odo/util/completion"
	"github.com/redhat-developer/odo/pkg/url"

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

	# Create a URL of ingress kind 
	%[1]s --port 8080 --ingress

	# Create a URL with a specific name and port
	%[1]s example --port 8080

	# Create a URL with a subdomain and hostname for example.domain.com
	%[1]s example --port 8080 --host domain.com

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
	*genericclioptions.Context

	contextFlag string

	// Parameters
	urlName string

	// Flags
	portFlag      int
	secureFlag    bool
	hostFlag      string // host of the URL
	tlsSecretFlag string // tlsSecret is the secret to te used by the URL
	pathFlag      string // path of the URL
	protocolFlag  string // protocol of the URL
	containerFlag string // container to which the URL belongs
	ingressFlag   bool

	url localConfigProvider.LocalURL
}

// NewURLCreateOptions creates a new CreateOptions instance
func NewURLCreateOptions() *CreateOptions {
	return &CreateOptions{}
}

func (o *CreateOptions) SetClientset(clientset *clientset.Clientset) {}

// Complete completes CreateOptions after they've been Created
func (o *CreateOptions) Complete(cmdline cmdline.Cmdline, args []string) (err error) {
	params := genericclioptions.NewCreateParameters(cmdline).NeedDevfile(o.contextFlag).RequireRouteAvailability()
	o.Context, err = genericclioptions.New(params)
	if err != nil {
		return err
	}

	var urlType localConfigProvider.URLKind
	if o.ingressFlag {
		urlType = localConfigProvider.INGRESS
	}

	// get the name
	if len(args) != 0 {
		o.urlName = args[0]
	}

	// create the localURL
	o.url = localConfigProvider.LocalURL{
		Name:      o.urlName,
		Port:      o.portFlag,
		Secure:    o.secureFlag,
		Host:      o.hostFlag,
		TLSSecret: o.tlsSecretFlag,
		Kind:      urlType,
		Container: o.containerFlag,
		Protocol:  o.protocolFlag,
		Path:      o.pathFlag,
	}

	// complete the URL
	err = o.Context.LocalConfigProvider.CompleteURL(&o.url)
	if err != nil {
		return err
	}

	return nil
}

// Validate validates the CreateOptions based on completed values
func (o *CreateOptions) Validate() (err error) {
	if !dfutil.CheckOutputFlag(o.GetOutputFlag()) {
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
	return nil
}

// Run contains the logic for the odo url create command
func (o *CreateOptions) Run(cmdline cmdline.Cmdline) (err error) {

	// create the URL and write it to the local config
	err = o.Context.LocalConfigProvider.CreateURL(o.url)
	if err != nil {
		return err
	}
	log.Successf("URL %s created for component: %v", o.url.Name, o.LocalConfigProvider.GetName())

	log.Italic("\nTo apply the URL configuration changes, please use `odo dev`")

	if log.IsJSON() {
		u := url.NewURLFromLocalURL(o.url)
		u.Status.State = url.StateTypeNotPushed
		machineoutput.OutputSuccess(u)
	}
	return nil
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
	clientset.Add(urlCreateCmd, clientset.PROJECT, clientset.PREFERENCE)

	urlCreateCmd.Flags().IntVarP(&o.portFlag, "port", "", -1, "Port number for the url of the component, required in case of components which expose more than one service port")
	urlCreateCmd.Flags().StringVar(&o.tlsSecretFlag, "tls-secret", "", "TLS secret name for the url of the component if the user bring their own TLS secret")
	urlCreateCmd.Flags().StringVarP(&o.hostFlag, "host", "", "", "Cluster IP for this URL (alias \"--hostname\")")

	// Alias for `--host` in case someone where to enter --hostname by mistake, and then mark it as hidden.
	urlCreateCmd.Flags().StringVarP(&o.hostFlag, "hostname", "", "", "Alias for --host")
	_ = urlCreateCmd.Flags().MarkHidden("hostname")

	urlCreateCmd.Flags().BoolVar(&o.ingressFlag, "ingress", false, "Create an Ingress instead of Route on OpenShift clusters")
	urlCreateCmd.Flags().BoolVarP(&o.secureFlag, "secure", "", false, "Create a secure HTTPS URL")
	urlCreateCmd.Flags().StringVarP(&o.pathFlag, "path", "", "", "path for this URL")
	urlCreateCmd.Flags().StringVarP(&o.protocolFlag, "protocol", "", string(devfilev1.HTTPEndpointProtocol), "protocol for this URL")
	urlCreateCmd.Flags().StringVarP(&o.containerFlag, "container", "", "", "container of the endpoint in devfile")

	odoutil.AddContextFlag(urlCreateCmd, &o.contextFlag)
	completion.RegisterCommandFlagHandler(urlCreateCmd, "context", completion.FileCompletionHandler)

	return urlCreateCmd
}
