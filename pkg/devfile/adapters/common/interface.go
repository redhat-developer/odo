package common

import (
	"io"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
)

// ComponentAdapter defines the functions that platform-specific adapters must implement
type ComponentAdapter interface {
	commandExecutor
	Push(parameters PushParameters) error
	DoesComponentExist(cmpName string, app string) (bool, error)
	Delete(labels map[string]string, show bool, wait bool) error
	Test(testCmd string, show bool) error
	CheckSupervisordCommandStatus(command devfilev1.Command) error
	StartContainerStatusWatch()
	StartSupervisordCtlStatusWatch()
	Log(follow bool, command devfilev1.Command) (io.ReadCloser, error)
	Exec(command []string) error
	Deploy() error
}
