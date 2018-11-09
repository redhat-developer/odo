package genericclioptions

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/redhat-developer/odo/pkg/application"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/occlient"
	"github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/project"
	"github.com/spf13/cobra"
	"os"
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

//
func client(command *cobra.Command, shouldSkipConnectionCheck ...bool) *occlient.Client {
	var skipConnectionCheck bool
	if len(shouldSkipConnectionCheck) > 0 {
		skipConnectionCheck = shouldSkipConnectionCheck[0]
	} else {
		var err error
		skipConnectionCheck, err = command.Flags().GetBool(util.SkipConnectionCheckFlagName)
		util.CheckError(err, "")
	}

	client, err := occlient.New(skipConnectionCheck)
	util.CheckError(err, "")

	return client
}

func newContext(command *cobra.Command, createAppIfNeeded bool) *Context {
	client := client(command)

	// project
	var ns string
	projectFlag := FlagValueIfSet(command, util.ProjectFlagName)
	if len(projectFlag) > 0 {
		_, err := project.Exists(client, projectFlag)
		util.CheckError(err, "")
		ns = projectFlag
	} else {
		ns = project.GetCurrent(client)
	}
	client.Namespace = ns

	// application
	var app string
	appFlag := FlagValueIfSet(command, util.ApplicationFlagName)
	if len(appFlag) > 0 {
		_, err := application.Exists(client, appFlag)
		util.CheckError(err, "")
		app = appFlag
	} else {
		var err error
		if !createAppIfNeeded {
			app, err = application.GetCurrent(ns)
		} else {
			app, err = application.GetCurrentOrGetCreateSetDefault(client)
		}
		util.CheckError(err, "unable to get current application")
	}

	internalCxt := internalCxt{
		Client:      client,
		Project:     ns,
		Application: app,
	}

	context := &Context{
		internalCxt: internalCxt,
	}

	// component
	var cmp string
	cmpFlag := FlagValueIfSet(command, util.ComponentFlagName)
	if len(cmpFlag) == 0 {
		var err error
		cmp, err = component.GetCurrent(app, ns)
		util.CheckError(err, "could not get current component")
	} else {
		context.existsOrExit(cmpFlag)
		cmp = cmpFlag
	}

	context.cmp = cmp

	return context
}

func FlagValueIfSet(cmd *cobra.Command, flagName string) string {
	flag, err := cmd.Flags().GetString(flagName)
	ignoreButLog(err)
	return flag
}

// Context holds contextual information useful to commands such as correctly configured client, target project and application
// (based on specified flag values) and provides for a way to retrieve a given component given this context
type Context struct {
	internalCxt
}

type internalCxt struct {
	Client      *occlient.Client
	Project     string
	Application string
	cmp         string
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
		if !allowEmpty && len(o.cmp) == 0 {
			fmt.Println("No component is set")
			os.Exit(-1)
		}
	case 1:
		cmp := optionalComponent[0]
		// only check the component if we passed a non-empty string, otherwise return the current component set in NewContext
		if len(cmp) > 0 {
			o.existsOrExit(cmp)
			o.cmp = cmp // update context
		}
	default:
		fmt.Printf("Component function only accepts one optional argument, was given: %v", optionalComponent)
		os.Exit(-1)
	}

	return o.cmp
}

func (o *Context) existsOrExit(cmp string) {
	exists, err := component.Exists(o.Client, cmp, o.Application)
	util.CheckError(err, "")
	if !exists {
		fmt.Printf("Component %v does not exist in application %s\n", cmp, o.Application)
		os.Exit(-1)
	}
}

func ignoreButLog(err error) {
	glog.V(4).Infof("Ignoring error as it usually means flag wasn't set: %v", err)
}
