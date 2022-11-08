package component

import (
	"context"

	"github.com/redhat-developer/odo/pkg/devfile/adapters"
	"github.com/redhat-developer/odo/pkg/watch"
)

// ComponentAdapter defines the functions that platform-specific adapters must implement
type ComponentAdapter interface {
	Push(ctx context.Context, parameters adapters.PushParameters, componentStatus *watch.ComponentStatus) error
}
