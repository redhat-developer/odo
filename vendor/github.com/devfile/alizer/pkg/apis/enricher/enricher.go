/*******************************************************************************
 * Copyright (c) 2021 Red Hat, Inc.
 * Distributed under license by Red Hat, Inc. All rights reserved.
 * This program is made available under the terms of the
 * Eclipse Public License v2.0 which accompanies this distribution,
 * and is available at http://www.eclipse.org/legal/epl-v20.html
 *
 * Contributors:
 * Red Hat, Inc.
 ******************************************************************************/

// Package enricher implements functions that detect the name and ports of a component.
// Uses three general strategies: Dockerfile, Compose, and Source.
// Dockerfile consists of using a dockerfile to extract information.
// Compose consists of using a compose file to extract information.
// Source consists of searching for specific statements of function invocations inside the source code.
package enricher

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/devfile/alizer/pkg/apis/model"
	"github.com/devfile/alizer/pkg/utils"
	"github.com/devfile/alizer/pkg/utils/langfiles"
	"gopkg.in/yaml.v3"
)

type Enricher interface {
	GetSupportedLanguages() []string
	DoEnrichLanguage(language *model.Language, files *[]string)
	DoEnrichComponent(component *model.Component, settings model.DetectionSettings, ctx *context.Context)
	IsConfigValidForComponentDetection(language string, configFile string) bool
}

type FrameworkDetectorWithConfigFile interface {
	GetSupportedFrameworks() []string
	DoFrameworkDetection(language *model.Language, config string)
	DoPortsDetection(component *model.Component, ctx *context.Context)
}

type FrameworkDetectorWithoutConfigFile interface {
	GetSupportedFrameworks() []string
	DoFrameworkDetection(language *model.Language, files *[]string)
	DoPortsDetection(component *model.Component, ctx *context.Context)
}

// IsConfigurationValidForLanguage checks whether the file is valid for the language.
//
// For example when analyzing a nodejs project, we could find a package.json
// within the node_modules folder. That is not to be considered valid
// for component detection.
func IsConfigurationValidForLanguage(language string, file string) bool {
	languageItem, err := langfiles.Get().GetLanguageByName(language)
	if err != nil {
		return false
	}
	for _, excludeFolder := range languageItem.ExcludeFolders {
		if isFolderNameIncludedInPath(file, excludeFolder) {
			return false
		}
	}
	return true
}

// isFolderNameIncludedInPath checks if fullPath contains potentialSubFolderName.
func isFolderNameIncludedInPath(fullPath string, potentialSubFolderName string) bool {
	pathSeparator := fmt.Sprintf("%c", os.PathSeparator)
	dir, _ := filepath.Split(fullPath)

	subDirectories := strings.Split(dir, pathSeparator)
	for _, subDir := range subDirectories {
		if strings.EqualFold(subDir, potentialSubFolderName) {
			return true
		}
	}
	return false
}

func getEnrichers() []Enricher {
	return []Enricher{
		&JavaEnricher{},
		&JavaScriptEnricher{},
		&PythonEnricher{},
		&DotNetEnricher{},
		&GoEnricher{},
		&PHPEnricher{},
		&DockerEnricher{},
	}
}

// GetEnricherByLanguage returns an enricher.
func GetEnricherByLanguage(language string) Enricher {
	for _, enricher := range getEnrichers() {
		// check the supported enricher languages
		if isLanguageSupportedByEnricher(language, enricher) {
			return enricher
		}
	}
	return nil
}

// isLanguageSupportedByEnricher checks the language has an enricher.
func isLanguageSupportedByEnricher(nameLanguage string, enricher Enricher) bool {
	for _, language := range enricher.GetSupportedLanguages() {
		if strings.EqualFold(language, nameLanguage) {
			return true
		}
	}
	return false
}

func GetDefaultProjectName(path string) string {
	return filepath.Base(path)
}

// GetPortsFromDockerFile returns a slice of port numbers from Dockerfiles in the given directory.
func GetPortsFromDockerFile(root string) []int {
	locations := utils.GetLocations(root)
	for _, location := range locations {
		filePath := filepath.Join(root, location)
		cleanFilePath := filepath.Clean(filePath)
		file, err := os.Open(cleanFilePath)
		if err == nil {
			defer func() error {
				if err := file.Close(); err != nil {
					return fmt.Errorf("error closing file: %s", err)
				}
				return nil
			}()
			return utils.ReadPortsFromDockerfile(file)
		}
	}
	return []int{}
}

// GetPortsFromDockerComposeFile returns a slice of port numbers from a compose file.
func GetPortsFromDockerComposeFile(componentPath string, settings model.DetectionSettings) []int {
	var ports []int
	bytes, err := getDockerComposeFileBytes(settings.BasePath)
	if err != nil {
		return ports
	}
	ports = getComponentPortsFromDockerComposeFileBytes(bytes, componentPath, settings.BasePath)
	if len(ports) > 0 || componentPath == settings.BasePath {
		return ports
	}

	// we already performed a search in the real root where the detection originally started. No compose file was there, so we try to look for
	// one in the actual component root
	bytes, err = getDockerComposeFileBytes(componentPath)
	if err != nil {
		return ports
	}
	return getComponentPortsFromDockerComposeFileBytes(bytes, componentPath, settings.BasePath)
}

// getDockerComposeFileBytes returns a byte slice of the compose file if found in the given directory.
func getDockerComposeFileBytes(root string) ([]byte, error) {
	return utils.ReadAnyApplicationFileExactMatch(root, []model.ApplicationFileInfo{
		{
			Dir:  "",
			File: "docker-compose.yml",
		},
		{
			Dir:  "",
			File: "docker-compose.yaml",
		},
		{
			Dir:  "",
			File: "compose.yml",
		},
		{
			Dir:  "",
			File: "compose.yaml",
		},
	})
}

// getComponentPortsFromDockerComposeFileBytes returns a slice of port numbers.
func getComponentPortsFromDockerComposeFileBytes(bytes []byte, componentPath string, basePath string) []int {
	var ports []int
	composeMap := make(map[string]interface{})
	err := yaml.Unmarshal(bytes, &composeMap)
	if err != nil {
		return ports
	}

	servicesField, hasServicesField := composeMap["services"].(map[string]interface{})
	if !hasServicesField {
		return ports
	}

	for _, serviceItem := range servicesField {
		serviceField, hasServiceField := serviceItem.(map[string]interface{})
		if !hasServiceField {
			continue
		}
		build, hasBuild := serviceField["build"].(string)
		if !hasBuild {
			continue
		}
		if build == "." || filepath.Join(basePath, build) == filepath.Clean(componentPath) {
			portsField, hasPortsField := serviceField["ports"].([]interface{})
			exposeField, hasExposeField := serviceField["expose"].([]interface{})
			if hasPortsField {
				re := regexp.MustCompile(`(\d+)\/*\w*$`) // ports syntax [HOST:]CONTAINER[/PROTOCOL] or map[string]interface
				for _, portInterface := range portsField {
					port := -1
					switch portInterfaceValue := portInterface.(type) {
					case string:
						port = utils.FindPortSubmatch(re, portInterfaceValue, 1)
					case map[string]interface{}:
						if targetInterface, exists := portInterfaceValue["target"]; exists {
							switch targetInterfaceValue := targetInterface.(type) {
							case int:
								if utils.IsValidPort(targetInterfaceValue) {
									port = targetInterfaceValue
								}
							case string:
								portValue, err := utils.GetValidPort(portInterfaceValue["target"].(string))
								if err == nil {
									port = portValue
								}
							default:
								break
							}
						}
					default:
						break
					}
					if port != -1 {
						ports = append(ports, port)
					}
				}
			}
			if hasExposeField {
				for _, portInterface := range exposeField {
					if portValue, ok := portInterface.(string); ok {
						port, err := utils.GetValidPort(portValue)
						if err == nil {
							ports = append(ports, port)
						}
					}
				}
			}
			break
		}
	}

	return ports
}
