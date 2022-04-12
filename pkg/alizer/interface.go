package alizer

import (
	"github.com/redhat-developer/alizer/go/pkg/apis/recognizer"
	"github.com/redhat-developer/odo/pkg/registry"
)

type Client interface {
	DetectFramework(path string) (recognizer.DevFileType, registry.Registry, error)
}
