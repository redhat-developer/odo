package main

import (
	"fmt"
	"runtime"
	"strings"

	version "github.com/hashicorp/go-version"
	"github.com/openshift/odo/pkg/odo/util"
)

func main() {
	minimumVersionString := ">= 1.13"
	minimumVersion, err := version.NewConstraint(minimumVersionString)
	if err != nil {
		util.LogErrorAndExit(err, "")
	}
	v := runtime.Version()
	v = strings.TrimPrefix(v, "go")
	sv, err := version.NewVersion(v)
	if err != nil {
		util.LogErrorAndExit(err, "")
	}
	if !minimumVersion.Check(sv) {
		err := fmt.Errorf("Golang version %s does not match the constraint %s, please install correct version", v, minimumVersionString)
		util.LogErrorAndExit(err, "")
	}
	fmt.Printf("Golang version %s checked successfully to be match constraint %s\n", v, minimumVersionString)
}
