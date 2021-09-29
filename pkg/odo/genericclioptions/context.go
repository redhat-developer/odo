package genericclioptions

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/openshift/odo/pkg/devfile"
	"github.com/openshift/odo/pkg/devfile/validate"
	"github.com/openshift/odo/pkg/localConfigProvider"
	odoutil "github.com/openshift/odo/pkg/util"

	"github.com/spf13/cobra"

	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/envinfo"
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/occlient"
	"github.com/openshift/odo/pkg/odo/util"
)

const (
	// DefaultAppName is the default name of the application when an application name is not provided
	DefaultAppName = "app"

	// gitDirName is the git dir name in a project
	gitDirName = ".git"
)

// Context holds contextual information useful to commands such as correctly configured client, target project and application
// (based on specified flag values) and provides for a way to retrieve a given component given this context
type Context struct {
	internalCxt
}

// internalCxt holds the actual context values and is not exported so that it cannot be instantiated outside of this package.
// This ensures that Context objects are always created properly via NewContext factory functions.
type internalCxt struct {
	ComponentContext    string
	Client              *occlient.Client
	command             *cobra.Command
	Project             string
	Application         string
	cmp                 string
	OutputFlag          string
	LocalConfigInfo     *config.LocalConfigInfo
	KClient             kclient.ClientInterface
	EnvSpecificInfo     *envinfo.EnvSpecificInfo
	LocalConfigProvider localConfigProvider.LocalConfigProvider
}

// CreateParameters defines the options which can be provided while creating the context
type CreateParameters struct {
	Cmd                    *cobra.Command
	DevfilePath            string
	ComponentContext       string
	IsNow                  bool
	CheckRouteAvailability bool
}

// New creates a context based on the given parameters
func New(parameters CreateParameters, toggles ...bool) (context *Context, err error) {
	parameters.DevfilePath = completeDevfilePath(parameters.ComponentContext, parameters.DevfilePath)
	isDevfile := odoutil.CheckPathExists(parameters.DevfilePath)
	if isDevfile {
		context, err = NewDevfileContext(parameters.Cmd)
		if err != nil {
			return context, err
		}
		context.ComponentContext = parameters.ComponentContext

		err = context.InitEnvInfoFromContext()
		if err != nil {
			return nil, err
		}

		// Parse devfile and validate
		devObj, err := devfile.ParseFromFile(parameters.DevfilePath)
		if err != nil {
			return context, fmt.Errorf("failed to parse the devfile %s, with error: %s", parameters.DevfilePath, err)
		}

		err = validate.ValidateDevfileData(devObj.Data)
		if err != nil {
			return context, err
		}

		context.EnvSpecificInfo.SetDevfileObj(devObj)

		context.Client, err = Client()
		if err != nil {
			return nil, err
		}
		context.resolveNamespace(context.EnvSpecificInfo)

		if parameters.CheckRouteAvailability {
			isRouteSupported, err := context.Client.IsRouteSupported()
			if err != nil {
				return nil, err
			}
			context.EnvSpecificInfo.SetIsRouteSupported(isRouteSupported)
		}
		context.LocalConfigProvider = context.EnvSpecificInfo
	} else {
		if parameters.IsNow {
			context, err = NewContextCreatingAppIfNeeded(parameters.Cmd)
			if err != nil {
				return nil, err
			}
			context.ComponentContext = parameters.ComponentContext
		} else {
			context, err = NewContext(parameters.Cmd)
			if err != nil {
				return nil, err
			}
			context.ComponentContext = parameters.ComponentContext
		}

		err = context.InitConfigFromContext()
		if err != nil {
			return nil, err
		}
		context.LocalConfigProvider = context.LocalConfigInfo
	}
	return context, nil
}

//InitConfigFromContext initializes localconfiginfo from the context
func (o *Context) InitConfigFromContext() error {
	var err error
	o.LocalConfigInfo, err = config.NewLocalConfigInfo(o.ComponentContext)
	if err != nil {
		return err
	}
	return nil
}

//InitEnvInfoFromContext initializes envinfo from the context
func (o *Context) InitEnvInfoFromContext() (err error) {
	o.EnvSpecificInfo, err = envinfo.NewEnvSpecificInfo(o.ComponentContext)
	if err != nil {
		return err
	}
	return nil
}

// completeDevfilePath completes the devfile path from context
func completeDevfilePath(componentContext, devfilePath string) string {
	if len(devfilePath) > 0 {
		return filepath.Join(componentContext, devfilePath)
	} else {
		return filepath.Join(componentContext, "devfile.yaml")
	}
}

// NewContext creates a new Context struct populated with the current state based on flags specified for the provided command
func NewContext(command *cobra.Command, toggles ...bool) (*Context, error) {
	ignoreMissingConfig := false
	createApp := false
	if len(toggles) == 1 {
		ignoreMissingConfig = toggles[0]
	}
	if len(toggles) == 2 {
		createApp = toggles[1]
	}
	return newContext(command, createApp, ignoreMissingConfig)
}

// NewDevfileContext creates a new Context struct populated with the current state based on flags specified for the provided command
func NewDevfileContext(command *cobra.Command) (*Context, error) {
	return newDevfileContext(command, false)
}

// NewContextCreatingAppIfNeeded creates a new Context struct populated with the current state based on flags specified for the
// provided command, creating the application if none already exists
func NewContextCreatingAppIfNeeded(command *cobra.Command) (*Context, error) {
	return newContext(command, true, false)
}

// NewConfigContext is a special kind of context which only contains local configuration, other information is not retrieved
//  from the cluster. This is useful for commands which don't want to connect to cluster.
func NewConfigContext(command *cobra.Command) *Context {

	// Check for valid config
	localConfiguration, err := getValidConfig(command, false)
	if err != nil {
		util.LogErrorAndExit(err, "")
	}
	outputFlag := FlagValueIfSet(command, OutputFlagName)

	ctx := &Context{
		internalCxt{
			LocalConfigInfo: localConfiguration,
			OutputFlag:      outputFlag,
		},
	}
	return ctx
}

// NewContextCompletion disables checking for a local configuration since when we use autocompletion on the command line, we
// couldn't care less if there was a configuration. We only need to check the parameters.
func NewContextCompletion(command *cobra.Command) *Context {
	ctx, err := newContext(command, false, true)
	if err != nil {
		util.LogErrorAndExit(err, "")
	}
	return ctx
}

// UpdatedContext returns a new context updated from config file
func UpdatedContext(context *Context) (*Context, *config.LocalConfigInfo, error) {
	localConfiguration, err := getValidConfig(context.command, false)
	if err != nil {
		return nil, nil, err
	}
	ctx, err := newContext(context.command, true, false)
	if err != nil {
		return nil, localConfiguration, err
	}
	return ctx, localConfiguration, err
}

// newContext creates a new context based on the command flags, creating missing app when requested
func newContext(command *cobra.Command, createAppIfNeeded bool, ignoreMissingConfiguration bool) (*Context, error) {
	// Create a new occlient
	client, err := ocClient()
	if err != nil {
		return nil, err
	}

	// Create a new kclient
	KClient, err := kclient.New()
	if err != nil {
		return nil, err
	}

	// Check for valid config
	localConfiguration, err := getValidConfig(command, ignoreMissingConfiguration)
	if err != nil {
		util.LogErrorAndExit(err, "")
	}

	// Resolve output flag
	outputFlag := FlagValueIfSet(command, OutputFlagName)

	// Create the internal context representation based on calculated values
	internalCxt := internalCxt{
		Client:          client,
		OutputFlag:      outputFlag,
		command:         command,
		LocalConfigInfo: localConfiguration,
		KClient:         KClient,
	}

	internalCxt.resolveProject(localConfiguration)
	internalCxt.resolveApp(createAppIfNeeded, localConfiguration)

	// Once the component is resolved, add it to the context
	internalCxt.resolveAndSetComponent(command, localConfiguration)

	// Create a context from the internal representation
	context := &Context{
		internalCxt: internalCxt,
	}

	return context, nil
}

// newDevfileContext creates a new context based on command flags for devfile components
func newDevfileContext(command *cobra.Command, createAppIfNeeded bool) (*Context, error) {

	// Resolve output flag
	outputFlag := FlagValueIfSet(command, OutputFlagName)

	// Create the internal context representation based on calculated values
	internalCxt := internalCxt{
		OutputFlag: outputFlag,
		command:    command,
		// this is only so we can make devfile and s2i work together for certain cases
		LocalConfigInfo: &config.LocalConfigInfo{},
	}

	// Get valid env information
	envInfo, err := getValidEnvInfo(command)
	if err != nil {
		util.LogErrorAndExit(err, "")
	}

	internalCxt.EnvSpecificInfo = envInfo
	internalCxt.resolveApp(createAppIfNeeded, envInfo)

	// Create a new kubernetes client
	internalCxt.KClient, err = kClient()
	if err != nil {
		return nil, err
	}
	internalCxt.Client, err = ocClient()
	if err != nil {
		return nil, err
	}

	// Gather the environment information
	internalCxt.EnvSpecificInfo = envInfo

	internalCxt.resolveNamespace(envInfo)

	// resolve the component
	internalCxt.resolveAndSetComponent(command, envInfo)

	// Create a context from the internal representation
	context := &Context{
		internalCxt: internalCxt,
	}
	return context, nil
}

// NewOfflineDevfileContext initializes a context for devfile components without any cluster calls
func NewOfflineDevfileContext(command *cobra.Command) *Context {
	// Resolve output flag
	outputFlag := FlagValueIfSet(command, OutputFlagName)

	// Create the internal context representation based on calculated values
	internalCxt := internalCxt{
		OutputFlag: outputFlag,
		command:    command,
		// this is only so we can make devfile and s2i work together for certain cases
		LocalConfigInfo: &config.LocalConfigInfo{},
	}

	// Get valid env information
	envInfo, err := getValidEnvInfo(command)
	if err != nil {
		util.LogErrorAndExit(err, "")
	}

	internalCxt.EnvSpecificInfo = envInfo
	internalCxt.LocalConfigProvider = envInfo
	internalCxt.resolveApp(false, envInfo)

	// resolve the component
	internalCxt.resolveAndSetComponent(command, envInfo)

	projectFlag := FlagValueIfSet(command, ProjectFlagName)
	if projectFlag != "" {
		internalCxt.Project = projectFlag
	} else {
		internalCxt.Project = envInfo.GetNamespace()
	}

	// Create a context from the internal representation
	context := &Context{
		internalCxt: internalCxt,
	}
	return context
}

// Component retrieves the optionally specified component or the current one if it is set. If no component is set, exit with
// an error
func (o *Context) Component(optionalComponent ...string) string {
	return o.ComponentAllowingEmpty(false, optionalComponent...)
}

// ComponentAllowingEmpty retrieves the optionally specified component or the current one if it is set, allowing empty
// components (instead of exiting with an error) if so specified
func (o *Context) ComponentAllowingEmpty(allowEmpty bool, optionalComponent ...string) string {
	switch len(optionalComponent) {
	case 0:
		// if we're not specifying a component to resolve, get the current one (resolved in NewContext as cmp)
		// so nothing to do here unless the calling context doesn't allow no component to be set in which case we exit with error
		if !allowEmpty && len(o.cmp) == 0 {
			log.Errorf("No component is set")
			os.Exit(1)
		}
	case 1:
		cmp := optionalComponent[0]
		o.cmp = cmp
	default:
		// safeguard: fail if more than one optional string is passed because it would be a programming error
		log.Errorf("ComponentAllowingEmpty function only accepts one optional argument, was given: %v", optionalComponent)
		os.Exit(1)
	}

	return o.cmp
}
