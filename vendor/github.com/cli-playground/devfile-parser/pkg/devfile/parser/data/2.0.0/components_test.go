package version200

import (
	"testing"

	common "github.com/cli-playground/devfile-parser/pkg/devfile/parser/data/common"
)

func TestGetCommands(t *testing.T) {

	testDevfile, execCommands := getTestDevfileData()

	got := testDevfile.GetCommands()
	want := execCommands

	for i, command := range got {
		if command.Exec != want[i].Exec {
			t.Error("Commands returned don't match expected commands")
		}
	}

}

func getTestDevfileData() (testDevfile Devfile200, commands []common.DevfileCommand) {

	command := "ls -la"
	component := "alias1"
	debugCommand := "nodemon --inspect={DEBUG_PORT}"
	debugComponent := "alias2"
	workDir := "/root"

	execCommands := []common.DevfileCommand{
		{
			Exec: &common.Exec{
				CommandLine: command,
				Component:   component,
				WorkingDir:  workDir,
			},
		},
		{
			Exec: &common.Exec{
				CommandLine: debugCommand,
				Component:   debugComponent,
				WorkingDir:  workDir,
			},
		},
	}

	testDevfileobj := Devfile200{
		Commands: execCommands,
	}

	return testDevfileobj, execCommands
}
