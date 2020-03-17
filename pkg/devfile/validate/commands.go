package validate

import (
	"fmt"
	"strings"

	"github.com/openshift/odo/pkg/devfile/parser/data/common"
)

const (
	BuildCmdSubstr = "build"
	RunCmdSubstr   = "run"
)

// ValidateCommands validates all the devfile commands
func ValidateCommands(commands []common.DevfileCommand) error {

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
		errMsg := fmt.Sprintf("odo requires '%s' and '%s' type of commands in devfile", common.DevfileCommandTypeBuild, common.DevfileCommandTypeRun)
		return fmt.Errorf(errMsg)
	}

	// Successful
	return nil
}
