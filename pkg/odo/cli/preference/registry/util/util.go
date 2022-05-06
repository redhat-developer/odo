package util

import (
	// odo packages

	"errors"
	"fmt"
	"strings"

	"github.com/redhat-developer/odo/pkg/preference"
	url2 "net/url"
)

const (
	RegistryUser = "default"
)

var ErrGithubRegistryNotSupported = errors.New("github based registries are no longer supported, use OCI based registries instead, see https://github.com/devfile/registry-support")

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

func IsGithubBasedRegistry(url string) (bool, error) {
	pu, err := url2.Parse(url)
	if err != nil {
		return false, fmt.Errorf("unable to parse registry url %w", err)
	}
	for _, d := range []string{"github.com", "raw.githubusercontent.com"} {
		if pu.Host == d || strings.HasSuffix(pu.Host, "."+d) {
			return true, nil
		}
	}
	return false, nil
}
