package common

// ComponentAdapter defines the functions that platform-specific adapters must implement
type ComponentAdapter interface {
	Push(parameters PushParameters) error
	DoesComponentExist(cmpName string) bool
	Delete(labels map[string]string) error
}

// StorageAdapter defines the storage functions that platform-specific adapters must implement
type StorageAdapter interface {
	Create([]Storage) error
}
