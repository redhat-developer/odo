package url

import (
	"fmt"

	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/envinfo"
	"github.com/openshift/odo/pkg/log"
	clicomponent "github.com/openshift/odo/pkg/odo/cli/component"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/odo/util/completion"
	"github.com/openshift/odo/pkg/odo/util/experimental"
	"github.com/openshift/odo/pkg/url"
	"github.com/pkg/errors"

	"github.com/openshift/odo/pkg/util"
	"github.com/spf13/cobra"

	ktemplates "k8s.io/kubernetes/pkg/kubectl/util/templates"
)

const createRecommendedCommandName = "create"

var (
	urlCreateShortDesc = `Create a URL for a component`
	urlCreateLongDesc  = ktemplates.LongDesc(`Create a URL for a component.

	The created URL can be used to access the specified component from outside the OpenShift cluster.
	`)
	urlCreateExample = ktemplates.Examples(`  # Create a URL with a specific name by automatically detecting the port used by the component
	%[1]s example

	# Create a URL for the current component with a specific port
	%[1]s --port 8080
  
	# Create a URL with a specific name and port
	%[1]s example --port 8080
	  `)
)

// URLCreateOptions encapsulates the options for the odo url create command
type URLCreateOptions struct {
	*clicomponent.CommonPushOptions
	urlName       string
	urlPort       int
	secureURL     bool
	componentPort int
	now           bool
	clusterHost   string
	tlsSecret     string
}

// NewURLCreateOptions creates a new URLCreateOptions instance
func NewURLCreateOptions() *URLCreateOptions {
	return &URLCreateOptions{CommonPushOptions: clicomponent.NewCommonPushOptions()}
}

// Complete completes URLCreateOptions after they've been Created
func (o *URLCreateOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	if o.now {
		o.Context = genericclioptions.NewContextCreatingAppIfNeeded(cmd)
	} else {
		o.Context = genericclioptions.NewContext(cmd)
	}
	if experimental.IsExperimentalModeEnabled() {
		o.clusterHost = args[0]
		err = o.InitEnvInfoFromContext()
	} else {
		err = o.InitConfigFromContext()
		if err != nil {
			return err
		}
		o.componentPort, err = url.GetValidPortNumber(o.Component(), o.urlPort, o.LocalConfigInfo.GetPorts())
		if err != nil {
			return err
		}
		if len(args) == 0 {
			o.urlName = url.GetURLName(o.Component(), o.componentPort)
		} else {
			o.urlName = args[0]
		}
		if o.now {
			prjName := o.LocalConfigInfo.GetProject()
			o.ResolveSrcAndConfigFlags()
			err = o.ResolveProject(prjName)
			if err != nil {
				return err
			}
		}
	}

	return
}

// Validate validates the URLCreateOptions based on completed values
func (o *URLCreateOptions) Validate() (err error) {
	// Check if exist
	if experimental.IsExperimentalModeEnabled() {
		for _, localURL := range o.EnvSpecificInfo.GetURL() {
			if o.clusterHost == localURL.ClusterHost {
				return fmt.Errorf("the cluster host: %s already exists in the application: %s", o.clusterHost, o.Application)
			}
		}
	} else {
		for _, localURL := range o.LocalConfigInfo.GetURL() {
			if o.urlName == localURL.Name {
				return fmt.Errorf("the url %s already exists in the application: %s", o.urlName, o.Application)
			}
		}
	}
	// Check if url name is more than 63 characters long
	if len(o.urlName) > 63 {
		return fmt.Errorf("url name must be shorter than 63 characters")
	}

	if !util.CheckOutputFlag(o.OutputFlag) {
		return fmt.Errorf("given output format %s is not supported", o.OutputFlag)
	}

	if o.now {
		err = o.ValidateComponentCreate()
		if err != nil {
			return err
		}
	}
	return
}

// Run contains the logic for the odo url create command
func (o *URLCreateOptions) Run() (err error) {
	if experimental.IsExperimentalModeEnabled() {
		err = o.EnvSpecificInfo.SetConfiguration("url", envinfo.ConfigURL{ClusterHost: o.clusterHost, Secure: o.secureURL, TLSSecret: o.tlsSecret})
	} else {
		err = o.LocalConfigInfo.SetConfiguration("url", config.ConfigURL{Name: o.urlName, Port: o.componentPort, Secure: o.secureURL})
	}
	if err != nil {
		return errors.Wrapf(err, "failed to persist the component settings to config file")
	}
	if experimental.IsExperimentalModeEnabled() {
		log.Successf("URL created for component: %v, cluster host: %v", o.Component(), o.clusterHost)
	} else {
		log.Successf("URL %s created for component: %v", o.urlName, o.Component())
	}
	if o.now {
		err = o.Push()
		if err != nil {
			return errors.Wrap(err, "failed to push changes")
		}
	} else {
		log.Italic("\nTo create URL on the OpenShift Cluster, please use `odo push`")
	}

	return
}

// NewCmdURLCreate implements the odo url create command.
func NewCmdURLCreate(name, fullName string) *cobra.Command {
	o := NewURLCreateOptions()
	urlCreateCmd := &cobra.Command{
		Use:     name + " [url name]",
		Short:   urlCreateShortDesc,
		Long:    urlCreateLongDesc,
		Example: fmt.Sprintf(urlCreateExample, fullName),
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}
	urlCreateCmd.Flags().IntVarP(&o.urlPort, "port", "", -1, "port number for the url of the component, required in case of components which expose more than one service port")
	// create ingress, if experimental mode is enabled.
	// cluster host is a required argument
	urlCreateCmd.Flags().BoolVarP(&o.secureURL, "secure", "", false, "creates a secure https url")
	if experimental.IsExperimentalModeEnabled() {
		urlCreateCmd.Use = name + " [cluster host]"
		urlCreateCmd.Args = cobra.RangeArgs(1, 1)
		urlCreateCmd.Flags().StringVarP(&o.tlsSecret, "tlsSecret", "", "", "tls secret name for the url of the component if the user bring his own tls secret")
	}
	o.AddContextFlag(urlCreateCmd)
	genericclioptions.AddNowFlag(urlCreateCmd, &o.now)
	completion.RegisterCommandFlagHandler(urlCreateCmd, "context", completion.FileCompletionHandler)

	return urlCreateCmd
}
