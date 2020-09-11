package genericclioptions

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/klog"

	"github.com/openshift/odo/pkg/component"
	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/envinfo"
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/occlient"
	"github.com/openshift/odo/pkg/odo/util"
	"github.com/openshift/odo/pkg/odo/util/pushtarget"
	pkgUtil "github.com/openshift/odo/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DefaultAppName is the default name of the application when an application name is not provided
const DefaultAppName = "app"

// NewContext creates a new Context struct populated with the current state based on flags specified for the provided command
func NewContext(command *cobra.Command, ignoreMissingConfiguration ...bool) *Context {
	ignoreMissingConfig := false
	if len(ignoreMissingConfiguration) == 1 {
		ignoreMissingConfig = ignoreMissingConfiguration[0]
	}
	return newContext(command, false, ignoreMissingConfig)
}

// NewDevfileContext creates a new Context struct populated with the current state based on flags specified for the provided command
func NewDevfileContext(command *cobra.Command, ignoreMissingConfiguration ...bool) *Context {
	return newDevfileContext(command)
}

// NewContextCreatingAppIfNeeded creates a new Context struct populated with the current state based on flags specified for the
// provided command, creating the application if none already exists
func NewContextCreatingAppIfNeeded(command *cobra.Command) *Context {
	return newContext(command, true, false)
}

// NewConfigContext is a special kind of context which only contains local configuration, other information is not retrived
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
// couldn't care less if there was a configuriation. We only need to check the parameters.
func NewContextCompletion(command *cobra.Command) *Context {
	return newContext(command, false, true)
}

// Client returns an oc client configured for this command's options
func Client(command *cobra.Command) *occlient.Client {
	return client(command)
}

// ClientWithConnectionCheck returns an oc client configured for this command's options but forcing the connection check status
// to the value of the provided bool, skipping it if true, checking the connection otherwise
func ClientWithConnectionCheck(command *cobra.Command, skipConnectionCheck bool) *occlient.Client {
	return client(command)
}

// client creates an oc client based on the command flags
func client(command *cobra.Command) *occlient.Client {
	client, err := occlient.New()
	util.LogErrorAndExit(err, "")

	return client
}

// kClient creates an kclient based on the command flags
func kClient(command *cobra.Command) *kclient.Client {
	kClient, err := kclient.New()
	util.LogErrorAndExit(err, "")

	return kClient
}

// checkProjectCreateOrDeleteOnlyOnInvalidNamespace errors out if user is trying to create or delete something other than project
// errFormatforCommand must contain one %s
func checkProjectCreateOrDeleteOnlyOnInvalidNamespace(command *cobra.Command, errFormatForCommand string) {
	// do not error out when its odo delete -a, so that we let users delete the local config on missing namespace
	if command.HasParent() && command.Parent().Name() != "project" && (command.Name() == "create" || (command.Name() == "delete" && !command.Flags().Changed("all"))) {
		err := fmt.Errorf(errFormatForCommand, command.Root().Name())
		util.LogErrorAndExit(err, "")
	}
}

// checkProjectCreateOrDeleteOnlyOnInvalidNamespaceNoFmt errors out if user is trying to create or delete something other than project
// compare to checkProjectCreateOrDeleteOnlyOnInvalidNamespace, no %s is needed
func checkProjectCreateOrDeleteOnlyOnInvalidNamespaceNoFmt(command *cobra.Command, errFormatForCommand string) {
	// do not error out when its odo delete -a, so that we let users delete the local config on missing namespace
	if command.HasParent() && command.Parent().Name() != "project" && (command.Name() == "create" || (command.Name() == "delete" && !command.Flags().Changed("all"))) {
		err := fmt.Errorf(errFormatForCommand)
		util.LogErrorAndExit(err, "")
	}
}

// getFirstChildOfCommand gets the first child command of the root command of command
func getFirstChildOfCommand(command *cobra.Command) *cobra.Command {
	// If command does not have a parent no point checking
	if command.HasParent() {
		// Get the root command and set current command and its parent
		rootCommand := command.Root()
		parentCommand := command.Parent()
		mainCommand := command
		for {
			// if parent is root, then we have our first child in c
			if parentCommand == rootCommand {
				return mainCommand
			}
			// Traverse backwards making current command as the parent and parent as the grandparent
			mainCommand = parentCommand
			parentCommand = mainCommand.Parent()
		}
	}
	return nil
}

// getValidEnvInfo accesses the environment file
func getValidEnvInfo(command *cobra.Command) (*envinfo.EnvSpecificInfo, error) {

	// Get details from the env file
	componentContext := FlagValueIfSet(command, ContextFlagName)

	// Grab the absolute path of the eenv file
	if componentContext != "" {
		fAbs, err := pkgUtil.GetAbsPath(componentContext)
		util.LogErrorAndExit(err, "")
		componentContext = fAbs
	}

	// Access the env file
	envInfo, err := envinfo.NewEnvSpecificInfo(componentContext)
	if err != nil {
		return nil, err
	}

	// Now we check to see if we can skip gathering the information.
	// Return if we can skip gathering configuration information
	canWeSkip, err := checkIfConfigurationNeeded(command)
	if err != nil {
		return nil, err
	}
	if canWeSkip {
		return envInfo, nil
	}

	// Check to see if the environment file exists
	if !envInfo.EnvInfoFileExists() {
		return nil, fmt.Errorf("The current directory does not represent an odo component. Use 'odo create' to create component here or switch to directory with a component")
	}

	return envInfo, nil
}

func GetContextFlagValue(command *cobra.Command) string {
	contextDir := FlagValueIfSet(command, ContextFlagName)

	// Grab the absolute path of the configuration
	if contextDir != "" {
		fAbs, err := pkgUtil.GetAbsPath(contextDir)
		util.LogErrorAndExit(err, "")
		contextDir = fAbs
	} else {
		fAbs, err := pkgUtil.GetAbsPath(".")
		util.LogErrorAndExit(err, "")
		contextDir = fAbs
	}
	return contextDir
}

func getValidConfig(command *cobra.Command, ignoreMissingConfiguration bool) (*config.LocalConfigInfo, error) {

	// Get details from the local config file
	contextDir := FlagValueIfSet(command, ContextFlagName)

	// Grab the absolute path of the configuration
	if contextDir != "" {
		fAbs, err := pkgUtil.GetAbsPath(contextDir)
		util.LogErrorAndExit(err, "")
		contextDir = fAbs
	}
	// Access the local configuration
	localConfiguration, err := config.NewLocalConfigInfo(contextDir)
	if err != nil {
		return nil, err
	}

	// Now we check to see if we can skip gathering the information.
	// If true, we just return.
	canWeSkip, err := checkIfConfigurationNeeded(command)
	if err != nil {
		return nil, err
	}
	if canWeSkip {
		return localConfiguration, nil
	}

	// If file does not exist at this point, raise an error
	// HOWEVER..
	// When using auto-completion, we should NOT error out, just ignore the fact that there is no configuration
	if !localConfiguration.ConfigFileExists() && ignoreMissingConfiguration {
		klog.V(4).Info("There is NO config file that exists, we are however ignoring this as the ignoreMissingConfiguration flag has been passed in as true")
	} else if !localConfiguration.ConfigFileExists() {
		return nil, fmt.Errorf("The current directory does not represent an odo component. Use 'odo create' to create component here or switch to directory with a component")
	}

	// else simply return the local config info
	return localConfiguration, nil
}

// resolveProject resolves project
func resolveProject(command *cobra.Command, client *occlient.Client, localConfiguration *config.LocalConfigInfo) string {
	var namespace string
	projectFlag := FlagValueIfSet(command, ProjectFlagName)
	if len(projectFlag) > 0 {
		// if project flag was set, check that the specified project exists and use it
		project, err := client.GetProject(projectFlag)
		if err != nil || project == nil {
			util.LogErrorAndExit(err, "")
		}
		namespace = projectFlag
	} else {
		namespace = localConfiguration.GetProject()
		if namespace == "" {
			namespace = client.Namespace
			if len(namespace) <= 0 {
				errFormat := "Could not get current project. Please create or set a project\n\t%s project create|set <project_name>"
				checkProjectCreateOrDeleteOnlyOnInvalidNamespace(command, errFormat)
			}
		}

		// check that the specified project exists
		_, err := client.GetProject(namespace)
		if err != nil {
			e1 := fmt.Sprintf("You don't have permission to create or set project '%s' or the project doesn't exist. Please create or set a different project\n\t", namespace)
			errFormat := fmt.Sprint(e1, "%s project create|set <project_name>")
			checkProjectCreateOrDeleteOnlyOnInvalidNamespace(command, errFormat)
		}
	}
	client.Namespace = namespace
	return namespace
}

// resolveNamespace resolves namespace for devfile component
func resolveNamespace(command *cobra.Command, client *kclient.Client, envSpecificInfo *envinfo.EnvSpecificInfo) string {
	var namespace string
	projectFlag := FlagValueIfSet(command, "project")
	if len(projectFlag) > 0 {
		// if namespace flag was set, check that the specified namespace exists and use it
		_, err := client.KubeClient.CoreV1().Namespaces().Get(projectFlag, metav1.GetOptions{})
		// do not error out when its odo delete -a, so that we let users delete the local config on missing namespace
		if command.HasParent() && command.Parent().Name() != "project" && !(command.Name() == "delete" && command.Flags().Changed("all")) {
			util.LogErrorAndExit(err, "")
		}
		namespace = projectFlag
	} else {
		namespace = envSpecificInfo.GetNamespace()
		if namespace == "" {
			namespace = client.Namespace
			if len(namespace) <= 0 {
				errFormat := "Could not get current namespace. Please create or set a namespace\n"
				checkProjectCreateOrDeleteOnlyOnInvalidNamespace(command, errFormat)
			}
		}

		// check that the specified namespace exists
		_, err := client.KubeClient.CoreV1().Namespaces().Get(namespace, metav1.GetOptions{})
		if err != nil {
			errFormat := fmt.Sprintf("You don't have permission to create or set namespace '%s' or the namespace doesn't exist. Please create or set a different namespace\n\t", namespace)
			// errFormat := fmt.Sprint(e1, "%s project create|set <project_name>")
			checkProjectCreateOrDeleteOnlyOnInvalidNamespaceNoFmt(command, errFormat)
		}
	}
	client.Namespace = namespace
	return namespace
}

// resolveApp resolves the app
func resolveApp(command *cobra.Command, createAppIfNeeded bool, localConfiguration envinfo.LocalConfigProvider) string {
	var app string
	appFlag := FlagValueIfSet(command, ApplicationFlagName)
	if len(appFlag) > 0 {
		app = appFlag
	} else {
		app = localConfiguration.GetApplication()
		if app == "" {
			if createAppIfNeeded {
				return DefaultAppName
			}
		}
	}
	return app
}

// ResolveAppFlag resolves the app from the flag
func ResolveAppFlag(command *cobra.Command) string {
	appFlag := FlagValueIfSet(command, ApplicationFlagName)
	if len(appFlag) > 0 {
		return appFlag
	}
	return DefaultAppName
}

// resolveComponent resolves component
func resolveComponent(command *cobra.Command, localConfiguration *config.LocalConfigInfo, context *Context) string {
	var cmp string
	cmpFlag := FlagValueIfSet(command, ComponentFlagName)
	if len(cmpFlag) == 0 {
		// retrieve the current component if it exists if we didn't set the component flag
		cmp = localConfiguration.GetName()
	} else {
		// if flag is set, check that the specified component exists
		context.checkComponentExistsOrFail(cmpFlag)
		cmp = cmpFlag
	}
	return cmp
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

	// Resolve project
	namespace := resolveProject(command, client, localConfiguration)

	// resolve application
	app := resolveApp(command, createAppIfNeeded, localConfiguration)

	// Resolve output flag
	outputFlag := FlagValueIfSet(command, OutputFlagName)

	// Create the internal context representation based on calculated values
	internalCxt := internalCxt{
		Client:          client,
		Project:         namespace,
		Application:     app,
		OutputFlag:      outputFlag,
		command:         command,
		LocalConfigInfo: localConfiguration,
		KClient:         KClient,
	}

	// Create a context from the internal representation
	context := &Context{
		internalCxt: internalCxt,
	}
	// Once the component is resolved, add it to the context
	context.cmp = resolveComponent(command, localConfiguration, context)

	return context
}

// newDevfileContext creates a new context based on command flags for devfile components
func newDevfileContext(command *cobra.Command) *Context {

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
	internalCxt.Application = resolveApp(command, true, envInfo)

	// If the push target is NOT Docker we will set the client to Kubernetes.
	if !pushtarget.IsPushTargetDocker() {

		// Create a new kubernetes client
		kClient := kClient(command)
		internalCxt.KClient = kClient

		// Gather the environment information
		internalCxt.EnvSpecificInfo = envInfo
		resolveNamespace(command, kClient, envInfo)
	}

	// Create a context from the internal representation
	context := &Context{
		internalCxt: internalCxt,
	}
	return context
}

// FlagValueIfSet retrieves the value of the specified flag if it is set for the given command
func FlagValueIfSet(cmd *cobra.Command, flagName string) string {
	flag, _ := cmd.Flags().GetString(flagName)
	return flag
}

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

// existsOrExit checks if the specified component exists with the given context and exits the app if not.
func (o *Context) checkComponentExistsOrFail(cmp string) {
	exists, err := component.Exists(o.Client, cmp, o.Application)
	util.LogErrorAndExit(err, "")
	if !exists {
		log.Errorf("Component %v does not exist in application %s", cmp, o.Application)
		os.Exit(1)
	}
}

// ApplyIgnore will take the current ignores []string and either ignore it (if .odoignore is used)
// or find the .gitignore file in the directory and use that instead.
func ApplyIgnore(ignores *[]string, sourcePath string) (err error) {
	if len(*ignores) == 0 {
		rules, err := pkgUtil.GetIgnoreRulesFromDirectory(sourcePath)
		if err != nil {
			util.LogErrorAndExit(err, "")
		}
		*ignores = append(*ignores, rules...)
	}
	return nil
}

// checkIfConfigurationNeeded checks against a set of commands that do *NOT* need configuration.
func checkIfConfigurationNeeded(command *cobra.Command) (bool, error) {

	// Here we will check for parent commands, if the match a certain criteria, we will skip
	// using the configuration.
	//
	// For example, `odo create` should NOT check to see if there is actually a configuration yet.
	if command.HasParent() {

		// Gather necessary preliminary information
		parentCommand := command.Parent()
		rootCommand := command.Root()
		flagValue := FlagValueIfSet(command, ApplicationFlagName)

		// Find the first child of the command, as some groups are allowed even with non existent configuration
		firstChildCommand := getFirstChildOfCommand(command)

		// This should *never* happen, but added just to be safe
		if firstChildCommand == nil {
			return false, fmt.Errorf("Unable to get first child of command")
		}
		// Case 1 : if command is create operation just allow it
		if command.Name() == "create" && (parentCommand.Name() == "component" || parentCommand.Name() == rootCommand.Name()) {
			return true, nil
		}
		// Case 2 : if command is describe or delete and app flag is used just allow it
		if (firstChildCommand.Name() == "describe" || firstChildCommand.Name() == "delete") && len(flagValue) > 0 {
			return true, nil
		}
		// Case 3 : if command is list, just allow it
		if firstChildCommand.Name() == "list" {
			return true, nil
		}
		// Case 4 : Check if firstChildCommand is project. If so, skip validation of context
		if firstChildCommand.Name() == "project" {
			return true, nil
		}
		// Case 5 : Check if specific flags are set for specific first child commands
		if firstChildCommand.Name() == "app" {
			return true, nil
		}
		// Case 6 : Check if firstChildCommand is catalog and request is to list or search
		if firstChildCommand.Name() == "catalog" && (parentCommand.Name() == "list" || parentCommand.Name() == "search") {
			return true, nil
		}
		// Case 7: Check if firstChildCommand is component and  request is list
		if (firstChildCommand.Name() == "component" || firstChildCommand.Name() == "service") && command.Name() == "list" {
			return true, nil
		}
		// Case 8 : Check if firstChildCommand is component and app flag is used
		if firstChildCommand.Name() == "component" && len(flagValue) > 0 {
			return true, nil
		}
		// Case 9 : Check if firstChildCommand is logout and app flag is used
		if firstChildCommand.Name() == "logout" {
			return true, nil
		}
		// Case 10: Check if firstChildCommand is service and command is create or delete. Allow it if that's the case
		if firstChildCommand.Name() == "service" && (command.Name() == "create" || command.Name() == "delete") {
			return true, nil
		}

	} else {
		return true, nil
	}

	return false, nil
}
