package genericclioptions

import (
	"fmt"
	"path/filepath"

	"github.com/openshift/odo/pkg/devfile"
	"github.com/openshift/odo/pkg/devfile/location"
	"github.com/openshift/odo/pkg/devfile/validate"
	"github.com/openshift/odo/pkg/localConfigProvider"
	"github.com/openshift/odo/pkg/odo/util"
	odoutil "github.com/openshift/odo/pkg/util"

	"github.com/spf13/cobra"

	"github.com/openshift/odo/pkg/envinfo"
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/occlient"
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
func New(parameters CreateParameters) (context *Context, err error) {
	parameters.DevfilePath = completeDevfilePath(parameters.ComponentContext, parameters.DevfilePath)
	isDevfile := odoutil.CheckPathExists(parameters.DevfilePath)
	if isDevfile {
		context, err = NewContext(parameters.Cmd)
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

		err = context.resolveNamespace(context.EnvSpecificInfo)
		if err != nil {
			return nil, err
		}

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
	}
	return context, nil
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
		return location.DevfileLocation(componentContext)
	}
}

// NewContext creates a new Context struct populated with the current state based on flags specified for the provided command
func NewContext(command *cobra.Command) (*Context, error) {
	return newContext(command, false)
}

// NewContextCreatingAppIfNeeded creates a new Context struct populated with the current state based on flags specified for the
// provided command, creating the application if none already exists
func NewContextCreatingAppIfNeeded(command *cobra.Command) (*Context, error) {
	return newContext(command, true)
}

// NewContextCompletion disables checking for a local configuration since when we use autocompletion on the command line, we
// couldn't care less if there was a configuration. We only need to check the parameters.
func NewContextCompletion(command *cobra.Command) *Context {
	ctx, err := newContext(command, false)
	if err != nil {
		util.LogErrorAndExit(err, "")
	}
	return ctx
}

// newContext creates a new context based on command flags for devfile components
func newContext(command *cobra.Command, createAppIfNeeded bool) (*Context, error) {

	// Resolve output flag
	outputFlag := FlagValueIfSet(command, OutputFlagName)

	// Create the internal context representation based on calculated values
	internalCxt := internalCxt{
		OutputFlag: outputFlag,
		command:    command,
	}

	// Get valid env information
	envInfo, err := getValidEnvInfo(command)
	if err != nil {
		return nil, err
	}

	// Create a new kubernetes client
	internalCxt.KClient, err = kClient()
	if err != nil {
		return nil, err
	}
	internalCxt.Client, err = ocClient()
	if err != nil {
		return nil, err
	}

	// Gather env specific info
	internalCxt.EnvSpecificInfo = envInfo
	internalCxt.resolveApp(createAppIfNeeded, envInfo)

	if err := internalCxt.resolveNamespace(envInfo); err != nil {
		return nil, err
	}

	// resolve the component
	if _, err = internalCxt.resolveAndSetComponent(command, envInfo); err != nil {
		return nil, err
	}
	// Create a context from the internal representation
	context := &Context{
		internalCxt: internalCxt,
	}
	return context, nil
}

// NewOfflineContext initializes a context for devfile components without any cluster calls
func NewOfflineContext(command *cobra.Command) (*Context, error) {
	// Resolve output flag
	outputFlag := FlagValueIfSet(command, OutputFlagName)

	// Create the internal context representation based on calculated values
	internalCxt := internalCxt{
		OutputFlag: outputFlag,
		command:    command,
	}

	// Get valid env information
	envInfo, err := getValidEnvInfo(command)
	if err != nil {
		return nil, err
	}

	internalCxt.EnvSpecificInfo = envInfo
	internalCxt.LocalConfigProvider = envInfo
	internalCxt.resolveApp(false, envInfo)

	// resolve the component
	_, err = internalCxt.resolveAndSetComponent(command, envInfo)
	if err != nil {
		return nil, err
	}
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
	return context, nil
}

// Component retrieves the optionally specified component or the current one if it is set. If no component is set, returns
// an error
func (o *Context) Component(optionalComponent ...string) (string, error) {
	return o.ComponentAllowingEmpty(false, optionalComponent...)
}

// ComponentAllowingEmpty retrieves the optionally specified component or the current one if it is set, allowing empty
// components (instead of exiting with an error) if so specified
func (o *Context) ComponentAllowingEmpty(allowEmpty bool, optionalComponent ...string) (string, error) {
	switch len(optionalComponent) {
	case 0:
		// if we're not specifying a component to resolve, get the current one (resolved in NewContext as cmp)
		// so nothing to do here unless the calling context doesn't allow no component to be set in which case we return an error
		if !allowEmpty && len(o.cmp) == 0 {
			return "", fmt.Errorf("No component is set")
		}
	case 1:
		cmp := optionalComponent[0]
		o.cmp = cmp
	default:
		// safeguard: fail if more than one optional string is passed because it would be a programming error
		return "", fmt.Errorf("ComponentAllowingEmpty function only accepts one optional argument, was given: %v", optionalComponent)
	}

	return o.cmp, nil
}
