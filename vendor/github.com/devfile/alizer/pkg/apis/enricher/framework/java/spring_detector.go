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
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/devfile/alizer/pkg/apis/model"
	"github.com/devfile/alizer/pkg/utils"
	"gopkg.in/yaml.v3"
)

type SpringDetector struct{}

func (s SpringDetector) GetSupportedFrameworks() []string {
	return []string{"Spring", "Spring Boot"}
}

func (s SpringDetector) GetApplicationFileInfos(componentPath string, ctx *context.Context) []model.ApplicationFileInfo {
	return []model.ApplicationFileInfo{
		{
			Context: ctx,
			Root:    componentPath,
			Dir:     "src/main/resources",
			File:    "application.properties",
		},
		{
			Context: ctx,
			Root:    componentPath,
			Dir:     "src/main/resources",
			File:    "application.yml",
		},
		{
			Context: ctx,
			Root:    componentPath,
			Dir:     "src/main/resources",
			File:    "application.yaml",
		},
	}
}

// DoFrameworkDetection uses the groupId to check for the framework name
func (s SpringDetector) DoFrameworkDetection(language *model.Language, config string) {
	if hasFwk, _ := hasFramework(config, "org.springframework", ""); hasFwk {
		language.Frameworks = append(language.Frameworks, s.GetSupportedFrameworks()...)
	}
}

// DoPortsDetection searches for ports in the env var and
// src/main/resources/application.properties, or src/main/resources/application.yaml
func (s SpringDetector) DoPortsDetection(component *model.Component, ctx *context.Context) {
	// case: port is set on env var
	ports := getSpringPortsFromEnvs()
	if len(ports) > 0 {
		component.Ports = ports
		return
	}

	// check if port is set on env var of dockerfile
	ports = getSpringPortsFromEnvDockerfile(component.Path)
	if len(ports) > 0 {
		component.Ports = ports
		return
	}

	// check if port is set inside application file
	appFileInfos := s.GetApplicationFileInfos(component.Path, ctx)
	if len(appFileInfos) == 0 {
		return
	}

	applicationFile := utils.GetAnyApplicationFilePath(component.Path, appFileInfos, ctx)
	if applicationFile == "" {
		return
	}

	var err error
	if filepath.Ext(applicationFile) == ".yml" || filepath.Ext(applicationFile) == ".yaml" {
		ports, err = getServerPortsFromYamlFile(applicationFile)
	} else {
		ports, err = getServerPortsFromPropertiesFile(applicationFile)
	}
	if err != nil {
		return
	}
	component.Ports = ports

}

func getSpringPortsFromEnvs() []int {
	return utils.GetValidPortsFromEnvs([]string{"SERVER_PORT", "SERVER_HTTP_PORT"})
}

func getSpringPortsFromEnvDockerfile(path string) []int {
	envVars, err := utils.GetEnvVarsFromDockerFile(path)
	if err != nil {
		return nil
	}
	envs := []string{"SERVER_PORT", "SERVER_HTTP_PORT"}
	return utils.GetValidPortsFromEnvDockerfile(envs, envVars)
}

func getServerPortsFromPropertiesFile(file string) ([]int, error) {
	props, err := utils.ConvertPropertiesFileAsPathToMap(file)
	if err != nil {
		return []int{}, err
	}

	ports := getPortsFromMap(props, []string{"server.port", "server.http.port"})
	if len(ports) > 0 {
		return ports, nil
	}
	return []int{}, errors.New("no port found")
}

func getPortsFromMap(props map[string]string, keys []string) []int {
	var ports []int
	for _, key := range keys {
		port := getPortFromMap(props, key)
		if port != -1 {
			ports = append(ports, port)
		}
	}
	return ports
}

func getPortFromMap(props map[string]string, key string) int {
	if portValue, exists := props[key]; exists {
		if port, err := utils.GetValidPort(portValue); err == nil {
			return port
		}
	}
	return -1
}

func getServerPortsFromYamlFile(file string) ([]int, error) {
	yamlFile, err := os.ReadFile(filepath.Clean(file))
	if err != nil {
		return []int{}, err
	}
	var data model.SpringApplicationProsServer
	err = yaml.Unmarshal(yamlFile, &data)
	if err != nil {
		return []int{}, err
	}
	var ports []int
	if data.Server.Port > 0 {
		ports = append(ports, data.Server.Port)
	}
	if data.Server.Http.Port > 0 {
		ports = append(ports, data.Server.Http.Port)
	}
	if len(ports) > 0 {
		return ports, nil
	}
	return []int{}, errors.New("no port found")
}
