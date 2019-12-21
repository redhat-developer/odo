package parser

import "github.com/openshift/odo/pkg/testingutil/filesystem"

// GetFs returns the filesystem object
func (d *DevfileCtx) GetFs() filesystem.Filesystem {
	return d.Fs
}
