package component

import "github.com/redhat-developer/odo/pkg/devfile/adapters"

// ComponentAdapter defines the functions that platform-specific adapters must implement
type ComponentAdapter interface {
	Push(parameters adapters.PushParameters) error
}
