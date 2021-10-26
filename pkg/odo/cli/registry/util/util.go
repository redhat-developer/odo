package util

import (
	// odo packages

	"strings"

	"github.com/openshift/odo/v2/pkg/log"
	"github.com/openshift/odo/v2/pkg/preference"
)

const (
	RegistryUser = "default"
)

// IsSecure checks if the registry is secure
func IsSecure(registryName string) (bool, error) {
	cfg, err := preference.New()
	if err != nil {
		return false, err
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

	return isSecure, nil
}

func IsGitBasedRegistry(url string) bool {
	return strings.Contains(url, "github.com") || strings.Contains(url, "raw.githubusercontent.com")
}

func PrintGitRegistryDeprecationWarning() {
	log.Deprecate("Git based registries", "Please see https://github.com/openshift/odo/v2/tree/main/docs/public/git-registry-deprecation.adoc")
}
