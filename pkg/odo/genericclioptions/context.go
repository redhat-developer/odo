package genericclioptions

import (
	"fmt"

	"github.com/devfile/library/pkg/devfile/parser"
	dfutil "github.com/devfile/library/pkg/util"

	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/devfile"
	"github.com/redhat-developer/odo/pkg/devfile/location"
	"github.com/redhat-developer/odo/pkg/devfile/validate"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/util"
	odoutil "github.com/redhat-developer/odo/pkg/util"

	"github.com/spf13/cobra"
)

const (

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
	// componentName is the name of the component (computed either from the Devfile metadata, or detected by Alizer, or built from the current directory)
	componentName string
	// The path of the detected devfile
	DevfilePath string
	DevfileObj  parser.DevfileObj
}

// CreateParameters defines the options which can be provided while creating the context
type CreateParameters struct {
	cmdline          cmdline.Cmdline
	componentContext string
	devfile          bool
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

func (o CreateParameters) WithVariables(variables map[string]string) CreateParameters {
	o.variables = variables
	return o
}

// New creates a context based on the given parameters
// If NeedDevfile is passed and a Devfile is not found, a NoDevfileError is returned with a valid context without Devfile
func New(parameters CreateParameters) (*Context, error) {
	ctx := internalCxt{}
	var err error

	if parameters.devfile {
		devfilePath := location.DevfileLocation(parameters.componentContext)
		isDevfile := odoutil.CheckPathExists(devfilePath)
		if isDevfile {
			ctx.DevfilePath, err = dfutil.GetAbsPath(devfilePath)
			if err != nil {
				return nil, err
			}
			// Parse devfile and validate
			var devObj parser.DevfileObj
			devObj, err = devfile.ParseAndValidateFromFileWithVariables(ctx.DevfilePath, parameters.variables)
			if err != nil {
				return nil, fmt.Errorf("failed to parse the devfile %s: %w", ctx.DevfilePath, err)
			}
			err = validate.ValidateDevfileData(devObj.Data)
			if err != nil {
				return nil, err
			}
			ctx.DevfileObj = devObj

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

func (o *Context) GetComponentName() string {
	return o.componentName
}
