package url

import (
	"fmt"
	"strconv"

	"github.com/openshift/odo/pkg/config"
	adapterutils "github.com/openshift/odo/pkg/devfile/adapters/kubernetes/utils"
	devfileParser "github.com/openshift/odo/pkg/devfile/parser"
	"github.com/openshift/odo/pkg/envinfo"
	"github.com/openshift/odo/pkg/log"
	clicomponent "github.com/openshift/odo/pkg/odo/cli/component"
	"github.com/openshift/odo/pkg/odo/cli/ui"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/odo/util/completion"
	"github.com/openshift/odo/pkg/odo/util/experimental"
	"github.com/openshift/odo/pkg/odo/util/pushtarget"
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
	  `)

	urlCreateExampleExperimental = ktemplates.Examples(`  # Create a URL with a specific host by automatically detecting the port used by the component (using CRC as an exampple)
	%[1]s example  --host apps-crc.testing
  
	# Create a URL with a specific name and host (using CRC as an example)
	%[1]s example --host apps-crc.testing

	# Create a URL for the current component with a specific port and host (using CRC as an example)
	%[1]s --port 8080 --host apps-crc.testing

	# Create a URL of ingress kind for the current component with a host (using CRC as an example)
	%[1]s --host apps-crc.testing --ingress

	# Create a secured URL for the current component with a specific host (using CRC as an example)
	%[1]s --host apps-crc.testing --secured
	  `)

	urlCreateExampleDocker = ktemplates.Examples(`  # Create a URL with a specific name by automatically detecting the port used by the component
	%[1]s example
	
	# Create a URL for the current component with a specific port
	%[1]s --port 8080
  
	# Create a URL with a specific port and exposed post
	%[1]s --port 8080 --exposed-port 55555
	  `)
)

// URLCreateOptions encapsulates the options for the odo url create command
type URLCreateOptions struct {
	*clicomponent.PushOptions
	urlName          string
	urlPort          int
	secureURL        bool
	componentPort    int
	now              bool
	host             string
	tlsSecret        string
	exposedPort      int
	forceFlag        bool
	isRouteSupported bool
	wantIngress      bool
	urlType          envinfo.URLKind
}

// NewURLCreateOptions creates a new URLCreateOptions instance
func NewURLCreateOptions() *URLCreateOptions {
	return &URLCreateOptions{PushOptions: clicomponent.NewPushOptions()}
}

// Complete completes URLCreateOptions after they've been Created
func (o *URLCreateOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	if experimental.IsExperimentalModeEnabled() && util.CheckPathExists(o.DevfilePath) {
		o.Context = genericclioptions.NewDevfileContext(cmd)
	} else if o.now {
		o.Context = genericclioptions.NewContextCreatingAppIfNeeded(cmd)
	} else {
		o.Context = genericclioptions.NewContext(cmd)
	}

	o.Client = genericclioptions.Client(cmd)

	routeSupported, err := o.Client.IsRouteSupported()
	if err != nil {
		return err
	}
	if routeSupported {
		o.isRouteSupported = true
	}

	if experimental.IsExperimentalModeEnabled() && util.CheckPathExists(o.DevfilePath) {
		if o.wantIngress || (!o.isRouteSupported) {
			o.urlType = envinfo.INGRESS
		} else {
			o.urlType = envinfo.ROUTE
		}

		if o.tlsSecret != "" && (!o.wantIngress || !o.secureURL) {
			return fmt.Errorf("tls secret is only available for secure URLs of ingress kind")
		}

		err = o.InitEnvInfoFromContext()
		if err != nil {
			return err
		}

		devObj, err := devfileParser.Parse(o.DevfilePath)
		if err != nil {
			return fmt.Errorf("fail to parse the devfile %s, with error: %s", o.DevfilePath, err)
		}
		containers, err := adapterutils.GetContainers(devObj)
		if err != nil {
			return err
		}
		if len(containers) == 0 {
			return fmt.Errorf("No valid components found in the devfile")
		}
		compWithEndpoint := 0
		var postList []string
		for _, c := range containers {
			if len(c.Ports) != 0 {
				compWithEndpoint++
				for _, port := range c.Ports {
					postList = append(postList, strconv.FormatInt(int64(port.ContainerPort), 10))
				}
			}
			if compWithEndpoint > 1 {
				return fmt.Errorf("Devfile should only have one component containing endpoint")
			}
		}
		if compWithEndpoint == 0 {
			return fmt.Errorf("No valid component with an endpoint found in the devfile")
		}
		componentName := o.EnvSpecificInfo.GetName()
		o.componentPort, err = url.GetValidPortNumber(componentName, o.urlPort, postList)
		if err != nil {
			return err
		}

		if pushtarget.IsPushTargetDocker() {
			o.exposedPort, err = url.GetValidExposedPortNumber(o.exposedPort)
			if err != nil {
				return err
			}
			o.urlType = envinfo.DOCKER
		}

		if len(args) != 0 {
			o.urlName = args[0]
		} else {
			o.urlName = url.GetURLName(componentName, o.componentPort)
		}

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
	if experimental.IsExperimentalModeEnabled() && util.CheckPathExists(o.DevfilePath) {
		// if experimental mode is enabled, and devfile is provided.
		// check if valid host is provided
		if !pushtarget.IsPushTargetDocker() && len(o.host) <= 0 && (!o.isRouteSupported || o.wantIngress) {
			return fmt.Errorf("host must be provided in order to create ingress")
		}
		for _, localURL := range o.EnvSpecificInfo.GetURL() {
			if o.urlName == localURL.Name {
				return fmt.Errorf("the url %s already exists", o.urlName)
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
	if !experimental.IsExperimentalModeEnabled() {
		if o.now {
			err = o.ValidateComponentCreate()
			if err != nil {
				return err
			}
		}
	}
	return
}

// Run contains the logic for the odo url create command
func (o *URLCreateOptions) Run() (err error) {
	if experimental.IsExperimentalModeEnabled() && util.CheckPathExists(o.DevfilePath) {
		if pushtarget.IsPushTargetDocker() {
			for _, localURL := range o.EnvSpecificInfo.GetURL() {
				if o.urlPort == localURL.Port && localURL.ExposedPort > 0 {
					if !o.forceFlag {
						if !ui.Proceed(fmt.Sprintf("Port %v already has an exposed port %v set for it. Do you want to override the exposed port", localURL.Port, localURL.ExposedPort)) {
							log.Info("Aborted by the user")
							return nil
						}
					}
					// delete existing port mapping
					err := o.EnvSpecificInfo.DeleteURL(localURL.Name)
					if err != nil {
						return err
					}
					break
				}
			}
		}
		err = o.EnvSpecificInfo.SetConfiguration("url", envinfo.EnvInfoURL{Name: o.urlName, Port: o.componentPort, Host: o.host, Secure: o.secureURL, TLSSecret: o.tlsSecret, ExposedPort: o.exposedPort, Kind: o.urlType})
	} else {
		err = o.LocalConfigInfo.SetConfiguration("url", config.ConfigURL{Name: o.urlName, Port: o.componentPort, Secure: o.secureURL})
	}
	if err != nil {
		return errors.Wrapf(err, "failed to persist the component settings to config file")
	}
	if experimental.IsExperimentalModeEnabled() && util.CheckPathExists(o.DevfilePath) {
		componentName := o.EnvSpecificInfo.GetName()
		if pushtarget.IsPushTargetDocker() {
			log.Successf("URL %s created for component: %v with exposed port: %v", o.urlName, componentName, o.exposedPort)
		} else {
			log.Successf("URL %s created for component: %v", o.urlName, componentName)
		}
	} else {
		log.Successf("URL %s created for component: %v", o.urlName, o.Component())
	}
	if o.now {
		if experimental.IsExperimentalModeEnabled() && util.CheckPathExists(o.DevfilePath) {
			err = o.DevfilePush()
		} else {
			err = o.Push()
		}
		if err != nil {
			return errors.Wrap(err, "failed to push changes")
		}
	} else {
		log.Italic("\nTo apply the URL configuration changes, please use `odo push`")
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
	// if experimental mode is enabled, add more flags to support ingress creation or docker application based on devfile
	if experimental.IsExperimentalModeEnabled() {
		if pushtarget.IsPushTargetDocker() {
			urlCreateCmd.Flags().IntVarP(&o.exposedPort, "exposed-port", "", -1, "an external port to the application container")
			urlCreateCmd.Flags().BoolVarP(&o.forceFlag, "force", "f", false, "Don't ask for confirmation, assign an exposed port directly")
			urlCreateCmd.Example = fmt.Sprintf(urlCreateExampleDocker, fullName)
		} else {
			urlCreateCmd.Flags().StringVar(&o.tlsSecret, "tls-secret", "", "tls secret name for the url of the component if the user bring his own tls secret")
			urlCreateCmd.Flags().StringVarP(&o.host, "host", "", "", "Cluster ip for this URL")
			urlCreateCmd.Flags().BoolVarP(&o.secureURL, "secure", "", false, "creates a secure https url")
			urlCreateCmd.Flags().BoolVar(&o.wantIngress, "ingress", false, "Creates an ingress instead of Route on OpenShift clusters")
			urlCreateCmd.Example = fmt.Sprintf(urlCreateExampleExperimental, fullName)
		}
		urlCreateCmd.Flags().StringVar(&o.DevfilePath, "devfile", "./devfile.yaml", "Path to a devfile.yaml")
	} else {
		urlCreateCmd.Flags().BoolVarP(&o.secureURL, "secure", "", false, "creates a secure https url")
		urlCreateCmd.Example = fmt.Sprintf(urlCreateExample, fullName)
	}
	genericclioptions.AddNowFlag(urlCreateCmd, &o.now)
	o.AddContextFlag(urlCreateCmd)
	completion.RegisterCommandFlagHandler(urlCreateCmd, "context", completion.FileCompletionHandler)

	return urlCreateCmd
}
