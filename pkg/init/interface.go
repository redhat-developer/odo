package init

import (
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/redhat-developer/odo/pkg/init/params"
)

type Client interface {
	SelectDevfile(args map[string]string) (*params.DevfileLocation, error)
	DownloadDevfile(devfileLocation *params.DevfileLocation, destDir string) (string, error)
	DownloadStarterProject(devfile parser.DevfileObj, project string, dest string) error
}
