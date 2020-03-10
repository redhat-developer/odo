package adapters

import "io"

type PlatformAdapter interface {
	Start(path string, out io.Writer, ignoredFiles []string, forceBuild bool, globExps []string, show bool) error
}
