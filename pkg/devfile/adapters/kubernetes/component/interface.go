package component

import "github.com/redhat-developer/odo/pkg/devfile/adapters/common"

// ComponentAdapter defines the functions that platform-specific adapters must implement
type ComponentAdapter interface {
	Push(parameters common.PushParameters) error
}
