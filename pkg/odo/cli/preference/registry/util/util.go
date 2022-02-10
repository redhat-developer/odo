package util

import (
	// odo packages

	"strings"

	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/preference"
)

const (
	RegistryUser = "default"
)

// IsSecure checks if the registry is secure
func IsSecure(prefClient preference.Client, registryName string) bool {
	isSecure := false
	if prefClient.RegistryList() != nil {
		for _, registry := range *prefClient.RegistryList() {
			if registry.Name == registryName && registry.Secure {
				isSecure = true
				break
			}
		}
	}

	return isSecure
}

func IsGitBasedRegistry(url string) bool {
	return strings.Contains(url, "github.com") || strings.Contains(url, "raw.githubusercontent.com")
}

func PrintGitRegistryDeprecationWarning() {
	log.Deprecate("Git based registries", "Please see https://github.com/redhat-developer/odo/tree/main/docs/public/git-registry-deprecation.adoc")
}
