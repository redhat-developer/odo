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
	// project used for the command, either passed with the `--project` flag, or the current one by default
	project string
	// application used for the command, either passed with the `--app` flag, or the current one by default
	application string
	// component used for the command, either passed with the `--component` flag, or the current one by default
	component string
	// componentContext is the value passed with the `--context` flag
	componentContext string
	// outputFlag is the value passed with the `--output` flag
	outputFlag string
	// Kclient can be used to access Kubernetes resources
	KClient kclient.ClientInterface
	// Client can be used to access OpenShift resources
	Client              *occlient.Client
	EnvSpecificInfo     *envinfo.EnvSpecificInfo
	LocalConfigProvider localConfigProvider.LocalConfigProvider
}

// CreateParameters defines the options which can be provided while creating the context
type CreateParameters struct {
	Cmd                    *cobra.Command
	DevfilePath            string
	ComponentContext       string
	CheckRouteAvailability bool
	Devfile                bool
	Offline                bool
	CreateAppIfNeeded      bool
}

// New creates a context based on the given parameters
func New(parameters CreateParameters) (*Context, error) {

	context, err := newContext(parameters)
	if err != nil {
		return nil, err
	}

	parameters.DevfilePath = completeDevfilePath(parameters.ComponentContext, parameters.DevfilePath)
	isDevfile := odoutil.CheckPathExists(parameters.DevfilePath)
	if parameters.Devfile && isDevfile {
		context.EnvSpecificInfo, err = envinfo.NewEnvSpecificInfo(context.componentContext)
		if err != nil {
			return nil, err
		}

		// Parse devfile and validate
		devObj, err := devfile.ParseAndValidateFromFile(parameters.DevfilePath)
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

		err = context.resolveNamespace(parameters.Cmd, context.EnvSpecificInfo)
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
	}
	return context, nil
}

// newContext creates a new context based on command flags for devfile components
func newContext(parameters CreateParameters) (*Context, error) {

	// Resolve output flag
	outputFlag := FlagValueIfSet(parameters.Cmd, OutputFlagName)

	// Get valid env information
	envInfo, err := getValidEnvInfo(parameters.Cmd)
	if err != nil {
		return nil, err
	}

	// Create the internal context representation based on calculated values
	ctx := internalCxt{
		outputFlag:       outputFlag,
		EnvSpecificInfo:  envInfo,
		componentContext: parameters.ComponentContext,
	}

	if parameters.Offline {
		ctx.LocalConfigProvider = envInfo
		projectFlag := FlagValueIfSet(parameters.Cmd, ProjectFlagName)
		if projectFlag != "" {
			ctx.project = projectFlag
		} else {
			ctx.project = envInfo.GetNamespace()
		}
	} else {
		ctx.KClient, err = kclient.New()
		if err != nil {
			return nil, err
		}
		ctx.Client, err = occlient.New()
		if err != nil {
			return nil, err
		}
		if e := ctx.resolveNamespace(parameters.Cmd, envInfo); e != nil {
			return nil, e
		}
	}

	ctx.resolveApp(parameters.Cmd, parameters.CreateAppIfNeeded, envInfo)

	if err = ctx.resolveAndSetComponent(parameters.Cmd, envInfo); err != nil {
		return nil, err
	}

	// Create a context from the internal representation
	context := &Context{
		internalCxt: ctx,
	}
	return context, nil
}

// completeDevfilePath completes the devfile path from context
func completeDevfilePath(componentContext, devfilePath string) string {
	if len(devfilePath) > 0 {
		return filepath.Join(componentContext, devfilePath)
	} else {
		return location.DevfileLocation(componentContext)
	}
}

// NewContextCompletion disables checking for a local configuration since when we use autocompletion on the command line, we
// couldn't care less if there was a configuration. We only need to check the parameters.
func NewContextCompletion(command *cobra.Command) *Context {
	ctx, err := newContext(CreateParameters{Cmd: command})
	if err != nil {
		util.LogErrorAndExit(err, "")
	}
	return ctx
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
		if !allowEmpty && len(o.component) == 0 {
			return "", fmt.Errorf("No component is set")
		}
	case 1:
		cmp := optionalComponent[0]
		o.component = cmp
	default:
		// safeguard: fail if more than one optional string is passed because it would be a programming error
		return "", fmt.Errorf("ComponentAllowingEmpty function only accepts one optional argument, was given: %v", optionalComponent)
	}

	return o.component, nil
}

func (o *Context) GetProject() string {
	return o.project
}

func (o *Context) GetApplication() string {
	return o.application
}

func (o *Context) GetOutputFlag() string {
	return o.outputFlag
}

func (o *Context) GetComponentContext() string {
	return o.componentContext
}
