package genericclioptions

import (
	"errors"
	"fmt"

	"github.com/redhat-developer/odo/pkg/devfile"
	"github.com/redhat-developer/odo/pkg/devfile/location"
	"github.com/redhat-developer/odo/pkg/devfile/validate"
	"github.com/redhat-developer/odo/pkg/localConfigProvider"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/util"
	odoutil "github.com/redhat-developer/odo/pkg/util"

	"github.com/spf13/cobra"

	"github.com/redhat-developer/odo/pkg/envinfo"
	"github.com/redhat-developer/odo/pkg/kclient"
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
	// outputFlag is the value passed with the `-o` flag
	outputFlag string
	// The path of the detected devfile
	devfilePath string
	// Kclient can be used to access Kubernetes resources
	KClient             kclient.ClientInterface
	EnvSpecificInfo     *envinfo.EnvSpecificInfo
	LocalConfigProvider localConfigProvider.LocalConfigProvider
}

// CreateParameters defines the options which can be provided while creating the context
type CreateParameters struct {
	cmdline          cmdline.Cmdline
	componentContext string
	devfile          bool
	offline          bool
	appIfNeeded      bool
}

func NewCreateParameters(cmdline cmdline.Cmdline) CreateParameters {
	return CreateParameters{cmdline: cmdline}
}

func (o CreateParameters) NeedDevfile(ctx string) CreateParameters {
	o.devfile = true
	o.componentContext = ctx
	return o
}

func (o CreateParameters) IsOffline() CreateParameters {
	o.offline = true
	return o
}

func (o CreateParameters) CreateAppIfNeeded() CreateParameters {
	o.appIfNeeded = true
	return o
}

// New creates a context based on the given parameters
func New(parameters CreateParameters) (*Context, error) {
	ctx := internalCxt{}
	var err error

	ctx.EnvSpecificInfo, err = GetValidEnvInfo(parameters.cmdline)
	if err != nil {
		return nil, err
	}
	ctx.LocalConfigProvider = ctx.EnvSpecificInfo

	ctx.project = resolveProject(parameters.cmdline, ctx.EnvSpecificInfo)

	ctx.application = resolveApp(parameters.cmdline, ctx.EnvSpecificInfo, parameters.appIfNeeded)

	ctx.component = resolveComponent(parameters.cmdline, ctx.EnvSpecificInfo)

	ctx.componentContext = parameters.componentContext

	ctx.outputFlag = parameters.cmdline.FlagValueIfSet(util.OutputFlagName)

	if !parameters.offline {
		ctx.KClient, err = parameters.cmdline.GetKubeClient()
		if err != nil {
			return nil, err
		}
		if e := ctx.resolveProjectAndNamespace(parameters.cmdline, ctx.EnvSpecificInfo); e != nil {
			return nil, e
		}

		if parameters.cmdline.FlagValueIfSet(util.ComponentFlagName) != "" {
			if err = ctx.checkComponentExistsOrFail(); err != nil {
				return nil, err
			}
		}
	}

	ctx.devfilePath = location.DevfileLocation(parameters.componentContext)
	if parameters.devfile {
		isDevfile := odoutil.CheckPathExists(ctx.devfilePath)
		if isDevfile {
			// Parse devfile and validate
			devObj, err := devfile.ParseAndValidateFromFile(ctx.devfilePath)
			if err != nil {
				return nil, fmt.Errorf("failed to parse the devfile %s, with error: %s", ctx.devfilePath, err)
			}
			err = validate.ValidateDevfileData(devObj.Data)
			if err != nil {
				return nil, err
			}
			ctx.EnvSpecificInfo.SetDevfileObj(devObj)
		} else {
			return nil, errors.New("no devfile found")
		}
	}

	return &Context{
		internalCxt: ctx,
	}, nil
}

// NewContextCompletion disables checking for a local configuration since when we use autocompletion on the command line, we
// couldn't care less if there was a configuration. We only need to check the parameters.
func NewContextCompletion(command *cobra.Command) *Context {
	cmdline := cmdline.NewCobra(command)
	ctx, err := New(CreateParameters{cmdline: cmdline})
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
			return "", errors.New("no component is set")
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

func (o *Context) GetDevfilePath() string {
	return o.devfilePath
}
