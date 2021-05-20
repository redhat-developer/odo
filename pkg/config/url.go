package config

import (
	"fmt"
	"strings"

	"github.com/openshift/odo/pkg/localConfigProvider"
	"github.com/openshift/odo/pkg/util"
)

// GetPorts returns the ports stored in the config for the component
// returns default i.e nil if nil
func (lc *LocalConfig) GetPorts(container string) ([]string, error) {
	if lc.componentSettings.Ports == nil {
		return nil, nil
	}
	return *lc.componentSettings.Ports, nil
}

// CompleteURL completes the given URL with default values
func (lc *LocalConfig) CompleteURL(url *localConfigProvider.LocalURL) error {
	var err error

	url.Kind = localConfigProvider.ROUTE

	ports, err := lc.GetPorts("")
	if err != nil {
		return err
	}
	url.Port, err = util.GetValidPortNumber(lc.GetName(), url.Port, ports)
	if err != nil {
		return err
	}

	// get the name
	if len(url.Name) == 0 {
		url.Name = util.GetURLName(lc.GetName(), url.Port)
	}

	return nil
}

// ValidateURL validates the given URL
func (lc *LocalConfig) ValidateURL(url localConfigProvider.LocalURL) error {
	errorList := make([]string, 0)

	urls, err := lc.ListURLs()
	if err != nil {
		return err
	}
	for _, localURL := range urls {
		if url.Name == localURL.Name {
			errorList = append(errorList, fmt.Sprintf("URL %s already exists in application: %s", url.Name, lc.GetApplication()))
		}
	}

	if len(errorList) > 0 {
		return fmt.Errorf(strings.Join(errorList, "\n"))
	}

	return nil
}

// GetURL gets the given url localConfig
func (lc *LocalConfig) GetURL(name string) (*localConfigProvider.LocalURL, error) {
	urls, err := lc.ListURLs()
	if err != nil {
		return nil, err
	}
	for _, url := range urls {
		if name == url.Name {
			return &url, nil
		}
	}
	return nil, nil
}

// CreateURL writes the given url to the localConfig
func (lci *LocalConfigInfo) CreateURL(url localConfigProvider.LocalURL) error {
	return lci.SetConfiguration("url", localConfigProvider.LocalURL{Name: url.Name, Port: url.Port, Secure: url.Secure})
}

// ListURLs returns the ConfigURL, returns default if nil
func (lc *LocalConfig) ListURLs() ([]localConfigProvider.LocalURL, error) {
	if lc.componentSettings.URL == nil {
		return []localConfigProvider.LocalURL{}, nil
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
	return resultURLs, nil
}

// DeleteURL is used to delete config from local odo config
func (lci *LocalConfigInfo) DeleteURL(parameter string) error {
	for i, url := range *lci.componentSettings.URL {
		if url.Name == parameter {
			s := *lci.componentSettings.URL
			s = append(s[:i], s[i+1:]...)
			lci.componentSettings.URL = &s
		}
	}
	return lci.writeToFile()
}
