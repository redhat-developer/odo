package common

// ComponentAdapter defines the component functions that platform-specific adapters must implement
type ComponentAdapter interface {
	Create() error
}

// StorageAdapter defines the storage functions that platform-specific adapters must implement
type StorageAdapter interface {
	Create([]Storage) error
}
