package alizer

import (
	"github.com/redhat-developer/alizer/go/pkg/apis/recognizer"
	"github.com/redhat-developer/odo/pkg/api"
)

type Client interface {
	DetectFramework(path string) (recognizer.DevFileType, api.Registry, error)
}
