package alizer

import (
	"context"

	"github.com/redhat-developer/alizer/go/pkg/apis/model"
	"github.com/redhat-developer/odo/pkg/api"
)

type Client interface {
	DetectFramework(ctx context.Context, path string) (model.DevFileType, api.Registry, error)
	DetectName(path string) (string, error)
	DetectPorts(path string) ([]int, error)
}
