package deploy

import (
	"github.com/devfile/library/pkg/devfile/parser"

	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
)

type Client interface {
	// Deploy resources from a devfile located in path, for the specified appName
	Deploy(fs filesystem.Filesystem, devfileObj parser.DevfileObj, path string, appName string) error
}
