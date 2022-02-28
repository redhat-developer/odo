package delete

import (
	"github.com/devfile/library/pkg/devfile/parser"
)

type Client interface {
	UnDeploy(devfileObj parser.DevfileObj, path string) error
	DeleteComponent(devfileObj parser.DevfileObj, componentName string) error
}
