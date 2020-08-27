package parser

import "github.com/openshift/odo/pkg/testingutil/filesystem"

func FakeContext(fs filesystem.Filesystem, absPath string) DevfileCtx {
	return DevfileCtx{
		Fs:      fs,
		absPath: absPath,
	}
}
