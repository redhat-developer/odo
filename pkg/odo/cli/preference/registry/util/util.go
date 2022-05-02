package util

import (
	// odo packages

	"errors"
	"strings"

	"github.com/redhat-developer/odo/pkg/preference"
)

const (
	RegistryUser = "default"
)

var ErrGithubRegistryNotSupported = errors.New("github based registries are no longer supported")

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

func IsGithubBasedRegistry(url string) bool {
	return strings.Contains(url, "github.com") || strings.Contains(url, "raw.githubusercontent.com")
}
