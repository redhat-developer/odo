package alizer

import (
	"context"

	"github.com/devfile/alizer/pkg/apis/model"

	"github.com/redhat-developer/odo/pkg/api"
)

type Client interface {
	DetectFramework(ctx context.Context, path string) (_ model.DevFileType, defaultVersion string, _ api.Registry, _ error)
	DetectName(path string) (string, error)
	DetectPorts(path string) ([]int, error)
}
