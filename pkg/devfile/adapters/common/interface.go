package common

import "io"

// ComponentAdapter defines the functions that platform-specific adapters must implement
type ComponentAdapter interface {
	Start() error
	Push(path string, out io.Writer, files []string, delFiles []string, isForcePush bool, globExps []string, show bool) error
	DoesComponentExist(cmpName string) bool
}
