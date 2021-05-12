package util

import (
	// odo packages

	"os"
	"strings"

	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/preference"
	"github.com/pkg/errors"
)

const (
	RegistryUser = "default"
)

// IsSecure checks if the registry is secure
func IsSecure(registryName string) bool {
	cfg, err := preference.New()
	if err != nil {
		log.Error(errors.Cause(err))
		os.Exit(1)
	}

	isSecure := false
	if cfg.OdoSettings.RegistryList != nil {
		for _, registry := range *cfg.OdoSettings.RegistryList {
			if registry.Name == registryName && registry.Secure {
				isSecure = true
				break
			}
		}
	}

	return isSecure
}

func IsGitBasedRegistry(url string) bool {
	return strings.Contains(url, "github.com")
}

func PrintGitRegistryDeprecationWarning() {
	log.Deprecate("Git based registries", "Please see https://github.com/openshift/odo/tree/main/docs/public/git-registry-deprecation.adoc")
}
