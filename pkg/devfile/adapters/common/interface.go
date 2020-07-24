package common

import (
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/machineoutput"
	"io"
)

// ComponentAdapter defines the functions that platform-specific adapters must implement
type ComponentAdapter interface {
	ExecClient
	Push(parameters PushParameters) error
	DoesComponentExist(cmpName string) (bool, error)
	Delete(labels map[string]string, show bool) error
	Test(testCmd string, show bool) error
	Log(follow, debug bool) (io.ReadCloser, error)
	Exec(command []string) error
	Logger() machineoutput.MachineEventLoggingClient
	ComponentInfo(command common.DevfileCommand) (ComponentInfo, error)
	SupervisorComponentInfo(command common.DevfileCommand) (ComponentInfo, error)
}

// StorageAdapter defines the storage functions that platform-specific adapters must implement
type StorageAdapter interface {
	Create([]Storage) error
}
