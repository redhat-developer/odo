package component

import (
	"fmt"
	"github.com/devfile/api/v2/pkg/devfile"
	"github.com/pkg/errors"
	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/redhat-developer/odo/pkg/util"
	"os"
	"path/filepath"
	"strings"
)

// GetComponentDir returns source repo name
// Parameters:
//		path: source path
// Returns: directory name
func GetComponentDir(path string) (string, error) {
	retVal := ""
	if path != "" {
		retVal = filepath.Base(path)
	} else {
		currDir, err := os.Getwd()
		if err != nil {
			return "", errors.Wrapf(err, "unable to generate a random name as getting current directory failed")
		}
		retVal = filepath.Base(currDir)
	}
	retVal = strings.TrimSpace(util.GetDNS1123Name(strings.ToLower(retVal)))
	return retVal, nil
}

// GetComponentTypeFromDevfileMetadata returns component type from the devfile metadata;
// it could either be projectType or language, if neither of them are set, return 'Not available'
func GetComponentTypeFromDevfileMetadata(metadata devfile.DevfileMetadata) string {
	var componentType string
	if metadata.ProjectType != "" {
		componentType = metadata.ProjectType
	} else if metadata.Language != "" {
		componentType = metadata.Language
	} else {
		componentType = NotAvailable
	}
	return componentType
}

// GetProjectTypeFromDevfileMetadata returns component type from the devfile metadata
func GetProjectTypeFromDevfileMetadata(metadata devfile.DevfileMetadata) string {
	var projectType string
	if metadata.ProjectType != "" {
		projectType = metadata.ProjectType
	} else {
		projectType = NotAvailable
	}
	return projectType
}

// GetLanguageFromDevfileMetadata returns component type from the devfile metadata
func GetLanguageFromDevfileMetadata(metadata devfile.DevfileMetadata) string {
	var language string
	if metadata.Language != "" {
		language = metadata.Language
	} else {
		language = NotAvailable
	}
	return language
}

// GetDefaultComponentName generates a unique component name
// Parameters: desired default component name(w/o prefix) and slice of existing component names
// Returns: Unique component name and error if any
func GetDefaultComponentName(cfg preference.Client, componentPath string, componentType string, existingComponentList ComponentList) (string, error) {
	var prefix string

	// Get component names from component list
	var existingComponentNames []string
	for _, component := range existingComponentList.Items {
		existingComponentNames = append(existingComponentNames, component.Name)
	}
	var err error
	// If there's no prefix in config file, or its value is empty string use safe default - the current directory along with component type
	if cfg.NamePrefix() == nil || *cfg.NamePrefix() == "" {
		prefix, err = GetComponentDir(componentPath)
		if err != nil {
			return "", errors.Wrap(err, "unable to generate random component name")
		}
		prefix = util.TruncateString(prefix, componentRandomNamePartsMaxLen)
	} else {
		// Set the required prefix into componentName
		prefix = *cfg.NamePrefix()
	}

	// Generate unique name for the component using prefix and unique random suffix
	componentName, err := util.GetRandomName(
		fmt.Sprintf("%s-%s", componentType, prefix),
		componentNameMaxLen,
		existingComponentNames,
		componentNameMaxRetries,
	)
	if err != nil {
		return "", errors.Wrap(err, "unable to generate random component name")
	}

	return util.GetDNS1123Name(componentName), nil
}
