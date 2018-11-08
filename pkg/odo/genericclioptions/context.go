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

func newContext(command *cobra.Command, createAppIfNeeded bool) *Context {
	flags := command.Flags()
	skipConnectionCheck, err := flags.GetBool(util.SkipConnectionCheckFlagName)
	util.CheckError(err, "")

	client, err := occlient.New(skipConnectionCheck)
	util.CheckError(err, "")

	// project
	var ns string
	projectFlag, err := flags.GetString(util.ProjectFlagName)
	ignoreButLog(err)
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
	appFlag, err := flags.GetString(util.ApplicationFlagName)
	ignoreButLog(err)

	if len(appFlag) > 0 {
		_, err := application.Exists(client, appFlag)
		util.CheckError(err, "")
		app = appFlag
	} else {
		if !createAppIfNeeded {
			app, err = application.GetCurrent(ns)
		} else {
			app, err = application.GetCurrentOrGetCreateSetDefault(client)
		}
		util.CheckError(err, "unable to get current application")
	}

	context := &Context{
		Client:      client,
		Project:     ns,
		Application: app,
	}

	// component
	var cmp string
	cmpFlag, err := flags.GetString(util.ComponentFlagName)
	ignoreButLog(err)
	if len(cmpFlag) == 0 {
		cmp, err = component.GetCurrent(app, ns)
		util.CheckError(err, "could not get current component")
	} else {
		context.existsOrExit(cmpFlag)
		cmp = cmpFlag
	}

	context.cmp = cmp

	return context
}

// Context holds contextual information useful to commands such as correctly configured client, target project and application
// (based on specified flag values) and provides for a way to retrieve a given component given this context
type Context struct {
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

// GetOrCreateAppName retrieves the current application name from the context or creates a new default application
func (context *Context) GetOrCreateAppName() (applicationName string) {
	if len(ApplicationFlag) > 0 && len(ProjectFlag) > 0 {
		applicationName = context.Application
	} else {
		var err error
		applicationName, err = application.GetCurrentOrGetCreateSetDefault(context.Client)
		util.CheckError(err, "")
	}
	return
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
