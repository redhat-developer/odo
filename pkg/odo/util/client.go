package util

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"github.com/redhat-developer/odo/pkg/occlient"
	"os"
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

// Global variables
var (
	GlobalSkipConnectionCheck bool
)
