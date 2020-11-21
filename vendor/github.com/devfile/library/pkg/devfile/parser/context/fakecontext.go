package parser

import "github.com/devfile/library/pkg/testingutil/filesystem"

func FakeContext(fs filesystem.Filesystem, absPath string) DevfileCtx {
	return DevfileCtx{
		fs:      fs,
		absPath: absPath,
	}
}
