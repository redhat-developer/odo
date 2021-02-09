package config

import (
	"github.com/openshift/odo/pkg/localConfigProvider"
)

// GetPorts returns the ports stored in the config for the component
// returns default i.e nil if nil
func (lc *LocalConfig) GetPorts() []string {
	if lc.componentSettings.Ports == nil {
		return nil
	}
	return *lc.componentSettings.Ports
}

// ListURLs returns the ConfigURL, returns default if nil
func (lc *LocalConfig) ListURLs() []localConfigProvider.LocalURL {
	if lc.componentSettings.URL == nil {
		return []localConfigProvider.LocalURL{}
	}
	var resultURLs []localConfigProvider.LocalURL
	for _, url := range *lc.componentSettings.URL {
		resultURLs = append(resultURLs, localConfigProvider.LocalURL{
			Name:   url.Name,
			Port:   url.Port,
			Secure: url.Secure,
			Host:   url.Host,
			Path:   "/",
			Kind:   localConfigProvider.ROUTE,
		})
	}
	return resultURLs
}
