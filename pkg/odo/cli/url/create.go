package url

import (
	"fmt"

	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/odo/util/completion"
	"github.com/openshift/odo/pkg/url"
	"github.com/pkg/errors"

	"github.com/openshift/odo/pkg/util"
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
	localConfigInfo  *config.LocalConfigInfo
	componentContext string
	urlName          string
	urlPort          int
	componentPort    int
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
	o.localConfigInfo, err = config.NewLocalConfigInfo(o.componentContext)

	return
}

// Validate validates the UrlCreateOptions based on completed values
func (o *URLCreateOptions) Validate() (err error) {

	// Check if exist
	for _, localUrl := range o.localConfigInfo.GetUrl() {
		if o.urlName == localUrl.Name {
			return fmt.Errorf("the url %s already exists in the application: %s", o.urlName, o.Application)
		}
	}

	if !util.CheckOutputFlag(o.OutputFlag) {
		return fmt.Errorf("given output format %s is not supported", o.OutputFlag)
	}
	return
}

// Run contains the logic for the odo url create command
func (o *URLCreateOptions) Run() (err error) {
	err = o.localConfigInfo.SetConfiguration("url", config.ConfigUrl{Name: o.urlName, Port: o.urlPort})
	if err != nil {
		return errors.Wrapf(err, "failed to persist the component settings to config file")
	}
	log.Successf("URL created for component: %v\n", o.Component())
	fmt.Println("To create URL on the OpenShift cluster, please run `odo push`")
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
			genericclioptions.GenericRun(o, cmd, args)
		},
	}
	urlCreateCmd.Flags().IntVarP(&o.urlPort, "port", "", -1, "port number for the url of the component, required in case of components which expose more than one service port")
	_ = urlCreateCmd.MarkFlagRequired("port")
	genericclioptions.AddOutputFlag(urlCreateCmd)
	genericclioptions.AddContextFlag(urlCreateCmd, &o.componentContext)
	completion.RegisterCommandFlagHandler(urlCreateCmd, "context", completion.FileCompletionHandler)
	return urlCreateCmd
}
