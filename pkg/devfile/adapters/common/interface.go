package common

import "io"

// ComponentAdapter defines the functions that platform-specific adapters must implement
type ComponentAdapter interface {
	Push(parameters PushParameters) error
	Build(parameters BuildParameters) error
	DoesComponentExist(cmpName string) (bool, error)
	Delete(labels map[string]string) error
	Deploy(parameters DeployParameters) error
	DeployDelete(manifest []byte) error
	Test(testCmd string, show bool) error
	Log(follow, debug bool) (io.ReadCloser, error)
	Exec(command []string) error
}

// StorageAdapter defines the storage functions that platform-specific adapters must implement
type StorageAdapter interface {
	Create([]Storage) error
}
