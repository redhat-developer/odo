package genericclioptions

import (
	"fmt"

	"github.com/devfile/library/pkg/devfile/parser"

	"github.com/redhat-developer/odo/pkg/component"
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
	// componentContext is the value passed with the `--context` flag
	componentContext string
	// componentName is the name of the component (computed either from the Devfile metadata, or detected by Alizer, or built from the current directory)
	componentName string
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
	// variables override devfile variables
	variables map[string]string
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

func (o CreateParameters) WithVariables(variables map[string]string) CreateParameters {
	o.variables = variables
	return o
}

// New creates a context based on the given parameters
// If NeedDevfile is passed and a Devfile is not found, a NoDevfileError is returned with a valid context without Devfile
func New(parameters CreateParameters) (*Context, error) {
	ctx := internalCxt{}
	var err error

	ctx.EnvSpecificInfo, err = GetValidEnvInfo(parameters.cmdline)
	if err != nil {
		return nil, err
	}
	ctx.LocalConfigProvider = ctx.EnvSpecificInfo

	ctx.application = defaultAppName

	ctx.componentContext = parameters.componentContext

	if !parameters.offline {
		ctx.KClient, err = parameters.cmdline.GetKubeClient()
		if err != nil {
			return nil, err
		}
		if e := ctx.resolveProjectAndNamespace(parameters.cmdline); e != nil {
			return nil, e
		}

	}

	if parameters.devfile {
		devfilePath := location.DevfileLocation(parameters.componentContext)
		isDevfile := odoutil.CheckPathExists(devfilePath)
		if isDevfile {
			ctx.devfilePath = devfilePath
			// Parse devfile and validate
			var devObj parser.DevfileObj
			devObj, err = devfile.ParseAndValidateFromFileWithVariables(ctx.devfilePath, parameters.variables)
			if err != nil {
				return nil, fmt.Errorf("failed to parse the devfile %s: %w", ctx.devfilePath, err)
			}
			err = validate.ValidateDevfileData(devObj.Data)
			if err != nil {
				return nil, err
			}
			ctx.EnvSpecificInfo.SetDevfileObj(devObj)

			ctx.componentName, err = component.GatherName(parameters.componentContext, &devObj)
			if err != nil {
				return nil, err
			}
		} else {
			return &Context{
				internalCxt: ctx,
			}, NewNoDevfileError(".")
		}
	} else {
		ctx.componentName, err = component.GatherName(".", nil)
		if err != nil {
			return nil, err
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

func (o *Context) GetProject() string {
	return o.project
}

func (o *Context) GetApplication() string {
	return o.application
}

func (o *Context) GetComponentName() string {
	return o.componentName
}

func (o *Context) GetDevfilePath() string {
	return o.devfilePath
}
