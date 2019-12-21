package common

import (
	"fmt"
	"strings"
)

const (
	BuildCmdSubstr = "build"
	RunCmdSubstr   = "run"
)

// ValidateCommands validates all the devfile commands
func ValidateCommands(commands []DevfileCommand) error {

	// 1. Check if run and build commands are present
	var (
		isBuildCommandPresent = false
		isRunCommandPresent   = false
	)
	for _, command := range commands {
		if strings.Contains(strings.ToLower(command.Name), BuildCmdSubstr) {
			isBuildCommandPresent = true
		}
		if strings.Contains(strings.ToLower(command.Name), RunCmdSubstr) {
			isRunCommandPresent = true
		}
	}

	if !isRunCommandPresent || !isBuildCommandPresent {
		errMsg := fmt.Sprintf("odo requires '%s' and '%s' type of commands in devfile", DevfileCommandTypeBuild, DevfileCommandTypeRun)
		return fmt.Errorf(errMsg)
	}

	// Successful
	return nil
}
