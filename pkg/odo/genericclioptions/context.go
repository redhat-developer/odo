package genericclioptions

import (
	"fmt"
	"os"

	"github.com/golang/glog"
	"github.com/spf13/cobra"

	"github.com/openshift/odo/pkg/application"
	"github.com/openshift/odo/pkg/component"
	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/occlient"
	"github.com/openshift/odo/pkg/odo/util"
	"github.com/openshift/odo/pkg/project"
	pkgUtil "github.com/openshift/odo/pkg/util"
)

// NewContext creates a new Context struct populated with the current state based on flags specified for the provided command
func NewContext(command *cobra.Command) *Context {
	return newContext(command, false)
}

// NewContextCreatingAppIfNeeded creates a new Context struct populated with the current state based on flags specified for the
// provided command, creating the application if none already exists
func NewContextCreatingAppIfNeeded(command *cobra.Command) *Context {
	return newContext(command, true)
}

// Client returns an oc client configured for this command's options
func Client(command *cobra.Command) *occlient.Client {
	return client(command)
}

// ClientWithConnectionCheck returns an oc client configured for this command's options but forcing the connection check status
// to the value of the provided bool, skipping it if true, checking the connection otherwise
func ClientWithConnectionCheck(command *cobra.Command, skipConnectionCheck bool) *occlient.Client {
	return client(command, skipConnectionCheck)
}

// client creates an oc client based on the command flags, overriding the skip connection check flag with the optionally
// specified shouldSkipConnectionCheck boolean.
// We use varargs to denote the optional status of that boolean.
func client(command *cobra.Command, shouldSkipConnectionCheck ...bool) *occlient.Client {
	var skipConnectionCheck bool
	switch len(shouldSkipConnectionCheck) {
	case 0:
		var err error
		skipConnectionCheck, err = command.Flags().GetBool(SkipConnectionCheckFlagName)
		util.LogErrorAndExit(err, "")
	case 1:
		skipConnectionCheck = shouldSkipConnectionCheck[0]
	default:
		// safeguard: fail if more than one optional bool is passed because it would be a programming error
		log.Errorf("client function only accepts one optional argument, was given: %v", shouldSkipConnectionCheck)
		os.Exit(1)
	}

	client, err := occlient.New(skipConnectionCheck)
	util.LogErrorAndExit(err, "")

	return client
}

// checkProjectCreateOrDeleteOnlyOnInvalidNamespace errors out if user is trying to create or delete something other than project
// errFormatforCommand must contain one %s
func checkProjectCreateOrDeleteOnlyOnInvalidNamespace(command *cobra.Command, errFormatForCommand string) {
	if command.HasParent() && command.Parent().Name() != "project" && (command.Name() == "create" || command.Name() == "delete") {
		err := fmt.Errorf(errFormatForCommand, command.Root().Name())
		util.LogErrorAndExit(err, "")
	}
}

// resolveProject resolves project
func resolveProject(command *cobra.Command, client *occlient.Client) string {
	var ns string
	projectFlag := FlagValueIfSet(command, ProjectFlagName)
	if len(projectFlag) > 0 {
		// if project flag was set, check that the specified project exists and use it
		_, err := project.Exists(client, projectFlag)
		util.LogErrorAndExit(err, "")
		ns = projectFlag
	} else {
		// Get details from config file
		fileName := FlagValueIfSet(command, ContextFlagName)
		if fileName != "" {
			fileAbs, err := pkgUtil.GetAbsPath(fileName)
			util.LogErrorAndExit(err, "")
			fileName = fileAbs
		}
		lci, err := config.NewLocalConfigInfo(fileName)
		util.LogErrorAndExit(err, "could not get component settings from config file")

		ns = lci.GetProject()
		if ns == "" {
			ns = project.GetCurrent(client)
			if len(ns) <= 0 {
				errFormat := "Could not get current project. Please create or set a project\n\t%s project create|set <project_name>"
				checkProjectCreateOrDeleteOnlyOnInvalidNamespace(command, errFormat)
			}
		}

		// check that the specified project exists
		_, err = project.Exists(client, ns)
		if err != nil {
			e1 := fmt.Sprintf("You dont have permission to project '%s' or it doesnt exist. Please create or set a different project\n\t", ns)
			errFormat := fmt.Sprint(e1, "%s project create|set <project_name>")
			checkProjectCreateOrDeleteOnlyOnInvalidNamespace(command, errFormat)
		}
	}
	client.Namespace = ns
	return ns
}

// newContext creates a new context based on the command flags, creating missing app when requested
func newContext(command *cobra.Command, createAppIfNeeded bool) *Context {
	client := client(command)

	// resolve project
	ns := resolveProject(command, client)

	// Get details from config file
	fileName := FlagValueIfSet(command, ContextFlagName)
	if fileName != "" {
		fAbs, err := pkgUtil.GetAbsPath(fileName)
		util.LogErrorAndExit(err, "")
		fileName = fAbs
	}
	lci, err := config.NewLocalConfigInfo(fileName)
	util.LogErrorAndExit(err, "could not get component settings from config file")

	// resolve application
	var app string
	appFlag := FlagValueIfSet(command, ApplicationFlagName)
	if len(appFlag) > 0 {
		app = appFlag
	} else {
		app = lci.GetApplication()
		if app == "" {
			if createAppIfNeeded {
				var err error
				app, err = application.GetDefaultAppName()
				if err != nil {
					util.LogErrorAndExit(err, "failed to generate a random app name")
				}
			}
		}
	}

	// resolve output flag
	outputFlag := FlagValueIfSet(command, OutputFlagName)

	// create the internal context representation based on calculated values
	internalCxt := internalCxt{
		Client:      client,
		Project:     ns,
		Application: app,
		OutputFlag:  outputFlag,
	}

	// create a context from the internal representation
	context := &Context{
		internalCxt: internalCxt,
	}

	// resolve component
	var cmp string
	cmpFlag := FlagValueIfSet(command, ComponentFlagName)
	if len(cmpFlag) == 0 {
		// retrieve the current component if it exists if we didn't set the component flag
		cmp = lci.GetName()
	} else {
		// if flag is set, check that the specified component exists
		context.checkComponentExistsOrFail(cmpFlag)
		cmp = cmpFlag
	}

	// once the component is resolved, add it to the context
	context.cmp = cmp

	return context
}

// FlagValueIfSet retrieves the value of the specified flag if it is set for the given command
func FlagValueIfSet(cmd *cobra.Command, flagName string) string {
	flag, err := cmd.Flags().GetString(flagName)

	// log the error for debugging purposes though an error should only occur if the flag hadn't been added to the command or
	// if the specified flag name doesn't match a string flag. This usually can be ignored.
	ignoreButLog(err)
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
	Client      *occlient.Client
	Project     string
	Application string
	cmp         string
	OutputFlag  string
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
		// only check the component if we passed a non-empty string, otherwise return the current component set in NewContext
		if len(cmp) > 0 {
			o.checkComponentExistsOrFail(cmp)
			o.cmp = cmp // update context
		}
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

// ignoreButLog logs a potential error when trying to resolve a flag value.
func ignoreButLog(err error) {
	if err != nil {
		glog.V(4).Infof("Ignoring error as it usually means flag wasn't set: %v", err)
	}
}
