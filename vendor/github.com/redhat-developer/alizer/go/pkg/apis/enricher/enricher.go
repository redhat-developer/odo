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
package enricher

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"github.com/redhat-developer/alizer/go/pkg/apis/model"
	utils "github.com/redhat-developer/alizer/go/pkg/utils"
	"github.com/redhat-developer/alizer/go/pkg/utils/langfiles"
	"gopkg.in/yaml.v3"
)

type Enricher interface {
	GetSupportedLanguages() []string
	DoEnrichLanguage(language *model.Language, files *[]string)
	DoEnrichComponent(component *model.Component, settings model.DetectionSettings)
	IsConfigValidForComponentDetection(language string, configFile string) bool
}

type FrameworkDetectorWithConfigFile interface {
	GetSupportedFrameworks() []string
	DoFrameworkDetection(language *model.Language, config string)
	DoPortsDetection(component *model.Component)
}

type FrameworkDetectorWithoutConfigFile interface {
	GetSupportedFrameworks() []string
	DoFrameworkDetection(language *model.Language, files *[]string)
	DoPortsDetection(component *model.Component)
}

/*
	IsConfigurationValidForLanguage check whether the configuration file is valid for current language.
									For example when analyzing a nodejs project, we could find a package.json
									within the node_modules folder. That is not to be considered valid
									for component detection.
	Paramenters:
		language: language name
		file: configuration file name
	Returns:
		bool: true if config file is valid for current language

*/
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

/*
	isFolderNameIncludedInPath check if fullpath contains potentialSubFolderName
	Parameters:
		fullPath: 				complete path of a file
		potentialSubFolderName: folder name
	Returns:
		bool: true if potentialSubFolderName is included in fullPath
*/
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
	}
}

func GetEnricherByLanguage(language string) Enricher {
	for _, enricher := range getEnrichers() {
		if isLanguageSupportedByEnricher(language, enricher) {
			return enricher
		}
	}
	return nil
}

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

func GetPortsFromDockerFile(root string) []int {
	file, err := os.Open(filepath.Join(root, "Dockerfile"))
	if err == nil {
		defer file.Close()
		return getPortsFromReader(file)
	}
	return []int{}
}

func getPortsFromReader(file io.Reader) []int {
	ports := []int{}
	res, err := parser.Parse(file)
	if err != nil {
		return ports
	}

	for _, child := range res.AST.Children {
		if strings.ToLower(child.Value) == "expose" {
			for n := child.Next; n != nil; n = n.Next {
				if port, err := strconv.Atoi(n.Value); err == nil {
					ports = append(ports, port)
				}

			}
		}
	}
	return ports
}

func GetPortsFromDockerComposeFile(componentPath string, settings model.DetectionSettings) []int {
	ports := []int{}
	bytes, err := getDockerComposeFileBytes(settings.BasePath)
	if err != nil {
		return ports
	}
	ports = getComponentPortsFromDockerComposeFileBytes(bytes, componentPath, settings.BasePath)
	if len(ports) > 0 || componentPath == settings.BasePath {
		return ports
	}

	// we already performed a search in the real root where the detection originally started. No compose file was there so we try to look for
	// one in the actual component root
	bytes, err = getDockerComposeFileBytes(componentPath)
	if err != nil {
		return ports
	}
	return getComponentPortsFromDockerComposeFileBytes(bytes, componentPath, settings.BasePath)
}

func getDockerComposeFileBytes(root string) ([]byte, error) {
	return utils.ReadAnyApplicationFile(root, []model.ApplicationFileInfo{
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

func getComponentPortsFromDockerComposeFileBytes(bytes []byte, componentPath string, basePath string) []int {
	ports := []int{}
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
