package url

import (
	"fmt"

	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/url"

	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/util"
	"github.com/spf13/cobra"

	ktemplates "k8s.io/kubernetes/pkg/kubectl/cmd/templates"
)

const createRecommendedCommandName = "create"

var (
	urlCreateShortDesc = `Create a URL for a component`
	urlCreateLongDesc  = ktemplates.LongDesc(`Create a URL for a component.

	The created URL can be used to access the specified component from outside the OpenShift cluster.
	`)
	urlCreateExample = ktemplates.Examples(`  # Create a URL for the current component with a specific port
	%[1]s --port 8080
  
	# Create a URL with a specific name and port
	%[1]s example --port 8080
  
	# Create a URL with a specific name by automatic detection of port (only for components which expose only one service port) 
	%[1]s example
  
	# Create a URL with a specific name and port for component frontend
	%[1]s example --port 8080 --component frontend
	  `)
)

// URLCreateOptions encapsulates the options for the odo url create command
type URLCreateOptions struct {
	urlName       string
	urlPort       int
	urlOpenFlag   bool
	componentPort int
	*genericclioptions.Context
}

// NewURLCreateOptions creates a new UrlCreateOptions instance
func NewURLCreateOptions() *URLCreateOptions {
	return &URLCreateOptions{}
}

// Complete completes UrlCreateOptions after they've been Created
func (o *URLCreateOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	o.Context = genericclioptions.NewContext(cmd)
	o.componentPort, err = url.GetValidPortNumber(o.Client, o.urlPort, o.Component(), o.Application)
	if err != nil {
		return err
	}
	if len(args) == 0 {
		o.urlName = url.GetURLName(o.Component(), o.componentPort)
	} else {
		o.urlName = args[0]
	}
	return
}

// Validate validates the UrlCreateOptions based on completed values
func (o *URLCreateOptions) Validate() (err error) {
	exists, err := url.Exists(o.Client, o.urlName, o.Component(), o.Application)
	if exists {
		return fmt.Errorf("The url %s already exists in the application: %s\n%v", o.urlName, o.Application, err)
	}

	return
}

// Run contains the logic for the odo url create command
func (o *URLCreateOptions) Run() (err error) {

	log.Infof("Adding URL to component: %v", o.Component())
	urlRoute, err := url.Create(o.Client, o.urlName, o.componentPort, o.Component(), o.Application)
	if err != nil {
		return err
	}
	urlCreated := url.GetURLString(*urlRoute)
	log.Successf("URL created for component: %v\n\n"+
		"%v - %v\n", o.Component(), urlRoute.Name, urlCreated)

	if o.urlOpenFlag {
		err := util.OpenBrowser(urlCreated)
		if err != nil {
			return fmt.Errorf("unable to open URL within default browser:\n%v", err)
		}
	}
	return
}

// NewCmdURLCreate implements the odo url create command.
func NewCmdURLCreate(name, fullName string) *cobra.Command {
	o := NewURLCreateOptions()
	urlCreateCmd := &cobra.Command{
		Use:     name + " [component name]",
		Short:   urlCreateShortDesc,
		Long:    urlCreateLongDesc,
		Example: fmt.Sprintf(urlCreateExample, fullName),
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			odoutil.LogErrorAndExit(o.Complete(name, cmd, args), "")
			odoutil.LogErrorAndExit(o.Validate(), "")
			odoutil.LogErrorAndExit(o.Run(), "")
		},
	}
	urlCreateCmd.Flags().IntVarP(&o.urlPort, "port", "", -1, "port number for the url of the component, required in case of components which expose more than one service port")
	urlCreateCmd.Flags().BoolVar(&o.urlOpenFlag, "open", false, "open the created link with your default browser")

	return urlCreateCmd
}
