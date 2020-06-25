package common

// ComponentAdapter defines the functions that platform-specific adapters must implement
type ComponentAdapter interface {
	Push(parameters PushParameters) error
	Build(parameters BuildParameters) error
	Deploy(parameters DeployParameters) error
	DoesComponentExist(cmpName string) bool
	Delete(labels map[string]string) error
	DeployDelete(manifest []byte) error
}

// StorageAdapter defines the storage functions that platform-specific adapters must implement
type StorageAdapter interface {
	Create([]Storage) error
}
