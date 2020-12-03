package genericclioptions

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/envinfo"
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/occlient"
	"github.com/openshift/odo/pkg/odo/util"
	"github.com/openshift/odo/pkg/odo/util/pushtarget"
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
	Client          *occlient.Client
	command         *cobra.Command
	Project         string
	Application     string
	cmp             string
	OutputFlag      string
	LocalConfigInfo *config.LocalConfigInfo
	KClient         *kclient.Client
	EnvSpecificInfo *envinfo.EnvSpecificInfo
}

// NewContext creates a new Context struct populated with the current state based on flags specified for the provided command
func NewContext(command *cobra.Command, toggles ...bool) *Context {
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
func NewDevfileContext(command *cobra.Command) *Context {
	return newDevfileContext(command, false)
}

// NewContextCreatingAppIfNeeded creates a new Context struct populated with the current state based on flags specified for the
// provided command, creating the application if none already exists
func NewContextCreatingAppIfNeeded(command *cobra.Command) *Context {
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
	return newContext(command, false, true)
}

// UpdatedContext returns a new context updated from config file
func UpdatedContext(context *Context) (*Context, *config.LocalConfigInfo, error) {
	localConfiguration, err := getValidConfig(context.command, false)
	return newContext(context.command, true, false), localConfiguration, err
}

// newContext creates a new context based on the command flags, creating missing app when requested
func newContext(command *cobra.Command, createAppIfNeeded bool, ignoreMissingConfiguration bool) *Context {
	// Create a new occlient
	client := client(command)

	// Create a new kclient
	KClient, err := kclient.New()
	if err != nil {
		util.LogErrorAndExit(err, "")
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

	return context
}

// newDevfileContext creates a new context based on command flags for devfile components
func newDevfileContext(command *cobra.Command, createAppIfNeeded bool) *Context {

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

	// If the push target is NOT Docker we will set the client to Kubernetes.
	if !pushtarget.IsPushTargetDocker() {

		// Create a new kubernetes client
		internalCxt.KClient = kClient(command)
		internalCxt.Client = client(command)

		// Gather the environment information
		internalCxt.EnvSpecificInfo = envInfo

		internalCxt.resolveNamespace(envInfo)
	}

	// resolve the component
	internalCxt.resolveAndSetComponent(command, envInfo)

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
