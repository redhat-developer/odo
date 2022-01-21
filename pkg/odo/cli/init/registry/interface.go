package registry

import (
	"github.com/devfile/registry-support/registry-library/library"
)

type Client interface {
	PullStackFromRegistry(registry string, stack string, destDir string, options library.RegistryOptions) error
}
