package alizer

import (
	"context"

	"github.com/redhat-developer/alizer/go/pkg/apis/recognizer"
	"github.com/redhat-developer/odo/pkg/api"
)

type Client interface {
	DetectFramework(ctx context.Context, path string) (recognizer.DevFileType, api.Registry, error)
	DetectName(path string) (string, error)
}
