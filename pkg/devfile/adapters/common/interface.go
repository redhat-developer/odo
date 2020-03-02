package common

// ComponentAdapter defines the functions that platform-specific adapters must implement
type ComponentAdapter interface {
	Start() error
}
