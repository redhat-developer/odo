package adapters

import "io"

type PlatformAdapter interface {
	Start() error
	Push(path string, out io.Writer, files []string, delFiles []string, isForcePush bool, globExps []string, show bool) error
	DoesComponentExist(cmpName string) bool
}
