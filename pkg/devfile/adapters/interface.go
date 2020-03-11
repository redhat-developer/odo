package adapters

type PlatformAdapter interface {
	Push(path string, ignoredFiles []string, forceBuild bool, globExps []string) error
}
