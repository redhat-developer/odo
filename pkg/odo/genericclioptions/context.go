package genericclioptions

import (
	"fmt"
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
	flags := command.Flags()
	skipConnectionCheck, err := flags.GetBool(util.SkipConnectionCheckFlagName)
	util.CheckError(err, "")

	client, err := occlient.New(skipConnectionCheck)
	util.CheckError(err, "")

	// project
	var ns string
	projectFlag, err := flags.GetString(util.ProjectFlagName)
	util.CheckError(err, "")
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
	util.CheckError(err, "")
	if len(appFlag) > 0 {
		_, err := application.Exists(client, appFlag)
		util.CheckError(err, "")
		app = appFlag
	} else {
		appName, err := application.GetCurrent(ns)
		util.CheckError(err, "unable to get current application")
		app = appName
	}

	context := &Context{
		Client:      client,
		Project:     ns,
		Application: app,
	}

	// component
	var cmp string
	cmpFlag, err := flags.GetString(util.ComponentFlagName)
	util.CheckError(err, "")
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

// Component retrieves the optionally specified component or the current one if it is set
func (o *Context) Component(optionalComponent ...string) string {
	if len(o.cmp) > 0 {
		return o.cmp
	}

	switch len(optionalComponent) {
	case 0:
		if len(o.cmp) == 0 {
			fmt.Println("No component is set")
			os.Exit(-1)
		}
		break
	case 1:
		cmp := optionalComponent[0]
		// only check the component if we passed a non-empty string, otherwise return the current component set in NewContext
		if len(cmp) > 0 {
			o.existsOrExit(cmp)
			o.cmp = cmp // update context
		}
		break
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
