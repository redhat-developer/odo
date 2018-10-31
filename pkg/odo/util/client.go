package util

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"github.com/redhat-developer/odo/pkg/application"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/occlient"
	"github.com/redhat-developer/odo/pkg/project"
	"os"
)

// RootCommandName is the name of the root command
const RootCommandName = "odo"

// Global variables
var (
	GlobalSkipConnectionCheck bool
	ProjectFlag               string
	ApplicationFlag           string
	ComponentFlag             string
)

// GetOcClient creates a client to connect to OpenShift cluster
func GetOcClient() *occlient.Client {
	client, err := occlient.New(GlobalSkipConnectionCheck)
	CheckError(err, "")
	return client
}

// CheckError prints the cause of the given error and exits the code with an
// exit code of 1.
// If the context is provided, then that is printed, if not, then the cause is
// detected using errors.Cause(err)
func CheckError(err error, context string, a ...interface{}) {
	if err != nil {
		glog.V(4).Infof("Error:\n%v", err)
		if context == "" {
			fmt.Println(errors.Cause(err))
		} else {
			fmt.Printf(fmt.Sprintf("%s\n", context), a...)
		}

		os.Exit(1)
	}
}

// GetAppName returns application name from the provided flag or if flag is not provided, it will return current application name
func GetAppName(client *occlient.Client) string {
	// applicationFlag is `--application` flag
	if ApplicationFlag != "" {
		_, err := application.Exists(client, ApplicationFlag)
		CheckError(err, "")
		return ApplicationFlag
	}
	applicationName, err := application.GetCurrent(client.Namespace)
	CheckError(err, "unable to get current application")

	return applicationName
}

// GetAndSetNamespace checks whether project flag is provided,
// if provided, it validates the name and sets it as namespace for further operations
// if not provided, it fetches current namespace and sets it as namespace for further operations
// getAndSetNamespace also return the project name
func GetAndSetNamespace(client *occlient.Client) string {
	// projectFlag is `--project` flag
	if ProjectFlag != "" {
		_, err := project.Exists(client, ProjectFlag)
		CheckError(err, "")
		client.Namespace = ProjectFlag
		return ProjectFlag
	}
	client.Namespace = project.GetCurrent(client)
	return client.Namespace
}

// GetComponent returns the component to be used for the operation. If an input
// component is specified, then it is returned if it exists, if not,
// the current component is fetched and returned. If no component set, throws error
func GetComponent(client *occlient.Client, inputComponent string, applicationName string) string {
	if len(inputComponent) == 0 {
		c, err := component.GetCurrent(applicationName, client.Namespace)
		CheckError(err, "Could not get current component")
		if c == "" {
			fmt.Println("There is no component set")
			os.Exit(1)
		}
		return c
	}
	exists, err := component.Exists(client, inputComponent, applicationName)
	CheckError(err, "")
	if !exists {
		fmt.Printf("Component %v does not exist\n", inputComponent)
		os.Exit(1)
	}
	return inputComponent
}
