package common

import (
	"github.com/openshift/odo/pkg/machineoutput"
	"io"
)

// ComponentAdapter defines the functions that platform-specific adapters must implement
type ComponentAdapter interface {
	Push(parameters PushParameters) error
	DoesComponentExist(cmpName string) (bool, error)
	Delete(labels map[string]string, show bool) error
	Test(testCmd string, show bool) error
	Log(follow, debug bool) (io.ReadCloser, error)
	Exec(command []string) error
	LoggingClient() machineoutput.MachineEventLoggingClient
}

// StorageAdapter defines the storage functions that platform-specific adapters must implement
type StorageAdapter interface {
	Create([]Storage) error
}
