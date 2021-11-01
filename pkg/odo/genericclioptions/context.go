package genericclioptions

import (
	"fmt"

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
	// defaultAppName is the default name of the application when an application name is not provided
	defaultAppName = "app"

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
	// The path of the detected devfile
	devfilePath string
	// Kclient can be used to access Kubernetes resources
	KClient kclient.ClientInterface
	// Client can be used to access OpenShift resources
	Client              *occlient.Client
	EnvSpecificInfo     *envinfo.EnvSpecificInfo
	LocalConfigProvider localConfigProvider.LocalConfigProvider
}

// createParameters defines the options which can be provided while creating the context
type createParameters struct {
	cmd               *cobra.Command
	componentContext  string
	routeAvailability bool
	devfile           bool
	offline           bool
	appIfNeeded       bool
}

func NewCreateParameters(cmd *cobra.Command) createParameters {
	return createParameters{cmd: cmd}
}

func (o createParameters) SetComponentContext(ctx string) createParameters {
	o.componentContext = ctx
	return o
}

func (o createParameters) CheckRouteAvailability() createParameters {
	o.routeAvailability = true
	return o
}

func (o createParameters) NeedDevfile() createParameters {
	o.devfile = true
	return o
}

func (o createParameters) IsOffline() createParameters {
	o.offline = true
	return o
}

func (o createParameters) CreateAppIfNeeded() createParameters {
	o.appIfNeeded = true
	return o
}

// New creates a context based on the given parameters
func New(parameters createParameters) (*Context, error) {

	// TODO check parameters is OK:
	// - setcomponentContext called when needDevfile is called

	ctx := internalCxt{}
	var err error

	ctx.EnvSpecificInfo, err = getValidEnvInfo(parameters.cmd)
	if err != nil {
		return nil, err
	}
	ctx.LocalConfigProvider = ctx.EnvSpecificInfo

	ctx.project = resolveProject(parameters.cmd, ctx.EnvSpecificInfo)

	ctx.application = resolveApp(parameters.cmd, ctx.EnvSpecificInfo, parameters.appIfNeeded)

	ctx.component = resolveComponent(parameters.cmd, ctx.EnvSpecificInfo)

	ctx.componentContext = parameters.componentContext

	ctx.outputFlag = FlagValueIfSet(parameters.cmd, OutputFlagName)

	if !parameters.offline {
		ctx.KClient, err = kclient.New()
		if err != nil {
			return nil, err
		}
		ctx.Client, err = Client()
		if err != nil {
			return nil, err
		}
		if e := ctx.resolveNamespace(parameters.cmd, ctx.EnvSpecificInfo); e != nil {
			return nil, e
		}

		if FlagValueIfSet(parameters.cmd, ComponentFlagName) != "" {
			if err = ctx.checkComponentExistsOrFail(); err != nil {
				return nil, err
			}
		}

		if parameters.routeAvailability {
			isRouteSupported, err := ctx.Client.IsRouteSupported()
			if err != nil {
				return nil, err
			}
			ctx.EnvSpecificInfo.SetIsRouteSupported(isRouteSupported)
		}
	}

	ctx.devfilePath = location.DevfileLocation(parameters.componentContext)
	isDevfile := odoutil.CheckPathExists(ctx.devfilePath)
	if parameters.devfile && isDevfile {
		// Parse devfile and validate
		devObj, err := devfile.ParseFromFile(ctx.devfilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse the devfile %s, with error: %s", ctx.devfilePath, err)
		}
		err = validate.ValidateDevfileData(devObj.Data)
		if err != nil {
			return nil, err
		}
		ctx.EnvSpecificInfo.SetDevfileObj(devObj)
	}

	return &Context{
		internalCxt: ctx,
	}, nil
}

// NewContextCompletion disables checking for a local configuration since when we use autocompletion on the command line, we
// couldn't care less if there was a configuration. We only need to check the parameters.
func NewContextCompletion(command *cobra.Command) *Context {
	ctx, err := New(createParameters{cmd: command})
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

func (o *Context) GetDevfilePath() string {
	return o.devfilePath
}
