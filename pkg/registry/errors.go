package registry

type ErrGithubRegistryNotSupported struct {
}

func (s *ErrGithubRegistryNotSupported) Error() string {
	return "github based registries are no longer supported, use OCI based registries instead, see https://github.com/devfile/registry-support"
}
