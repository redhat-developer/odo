package deploy

import "github.com/devfile/library/pkg/devfile/parser"

type Client interface {
	// Deploy resources from a devfile located in path, for the specified appName
	Deploy(devfileObj parser.DevfileObj, path string, appName string) error
}
