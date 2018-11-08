package util

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"github.com/redhat-developer/odo/pkg/occlient"
	"os"
)

const (
	// RootCommandName is the name of the root command
	RootCommandName = "odo"
	// SkipConnectionCheckFlagName is the name of the global flag used to skip connection check in the client
	SkipConnectionCheckFlagName = "skip-connection-check"
	// ProjectFlagName is the name of the flag allowing a user to specify which project to operate on
	ProjectFlagName = "project"
	// ApplicationFlagName is the name of the flag allowing a user to specify which application to operate on
	ApplicationFlagName = "app"
	// ComponentFlagName is the name of the flag allowing a user to specify which component to operate on
	ComponentFlagName = "component"
)

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
