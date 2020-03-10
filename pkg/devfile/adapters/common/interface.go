package common

import "io"

// ComponentAdapter defines the functions that platform-specific adapters must implement
type ComponentAdapter interface {
	Start(path string, out io.Writer, ignoredFiles []string, forceBuild bool, globExps []string, show bool) error
}
