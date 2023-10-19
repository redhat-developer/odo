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

type QuarkusDetector struct{}

func (q QuarkusDetector) GetSupportedFrameworks() []string {
	return []string{"Quarkus"}
}

func (q QuarkusDetector) GetApplicationFileInfos(componentPath string, ctx *context.Context) []model.ApplicationFileInfo {
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
func (q QuarkusDetector) DoFrameworkDetection(language *model.Language, config string) {
	if hasFwk, _ := hasFramework(config, "io.quarkus", ""); hasFwk {
		language.Frameworks = append(language.Frameworks, "Quarkus")
	}
}

// DoPortsDetection searches for ports in the env var, .env file, and
// src/main/resources/application.properties, or src/main/resources/application.yaml
func (q QuarkusDetector) DoPortsDetection(component *model.Component, ctx *context.Context) {
	// check if port is set on env var
	ports := getQuarkusPortsFromEnvs()
	if len(ports) > 0 {
		component.Ports = ports
		return
	}

	// check if port is set on env var of a dockerfile
	ports = getQuarkusPortsFromEnvDockerfile(component.Path)
	if len(ports) > 0 {
		component.Ports = ports
		return
	}

	// check if port is set on .env file
	insecureRequestEnabled := utils.GetStringValueFromEnvFile(component.Path, `QUARKUS_HTTP_INSECURE_REQUESTS=(\w*)`)
	regexes := []string{`QUARKUS_HTTP_SSL_PORT=(\d*)`}
	if insecureRequestEnabled != "disabled" {
		regexes = append(regexes, `QUARKUS_HTTP_PORT=(\d*)`)
	}
	ports = utils.GetPortValuesFromEnvFile(component.Path, regexes)
	if len(ports) > 0 {
		component.Ports = ports
		return
	}

	appFileInfos := q.GetApplicationFileInfos(component.Path, ctx)
	if len(appFileInfos) == 0 {
		return
	}

	// case: no port found as env var. Look into source code.
	applicationFile := utils.GetAnyApplicationFilePath(component.Path, appFileInfos, ctx)
	if applicationFile == "" {
		return
	}

	var err error
	if filepath.Ext(applicationFile) == ".yml" || filepath.Ext(applicationFile) == ".yaml" {
		ports, err = getServerPortsFromQuarkusApplicationYamlFile(applicationFile)
	} else {
		ports, err = getServerPortsFromQuarkusPropertiesFile(applicationFile)
	}
	if err != nil {
		return
	}
	component.Ports = ports
}

func getQuarkusPortsFromEnvs() []int {
	insecureRequestEnabled := os.Getenv("QUARKUS_HTTP_INSECURE_REQUESTS")
	envs := []string{"QUARKUS_HTTP_SSL_PORT"}
	if insecureRequestEnabled != "disabled" {
		envs = append(envs, "QUARKUS_HTTP_PORT")
	}
	return utils.GetValidPortsFromEnvs(envs)
}

func getQuarkusPortsFromEnvDockerfile(path string) []int {
	envVars, err := utils.GetEnvVarsFromDockerFile(path)
	if err != nil {
		return nil
	}
	insecureRequestEnabled := ""
	envs := []string{"QUARKUS_HTTP_SSL_PORT"}
	for _, envVar := range envVars {
		if envVar.Name == "QUARKUS_HTTP_INSECURE_REQUESTS" {
			insecureRequestEnabled = envVar.Value
			break
		}
	}

	if insecureRequestEnabled == "true" {
		envs = append(envs, "QUARKUS_HTTP_PORT")
	}

	return utils.GetValidPortsFromEnvDockerfile(envs, envVars)
}

func getServerPortsFromQuarkusPropertiesFile(file string) ([]int, error) {
	var ports []int
	props, err := utils.ConvertPropertiesFileAsPathToMap(file)
	if err != nil {
		return ports, err
	}
	if portSSLValue, exists := props["quarkus.http.ssl-port"]; exists {
		if port, err := utils.GetValidPort(portSSLValue); err == nil {
			ports = append(ports, port)
		}
	}
	if insecureValue, exists := props["quarkus.http.insecure-requests"]; !exists || insecureValue != "disabled" {
		if portValue, exists := props["quarkus.http.port"]; exists {
			if port, err := utils.GetValidPort(portValue); err == nil {
				ports = append(ports, port)
			}
		}
	}
	if len(ports) > 0 {
		return ports, nil
	}
	return []int{}, errors.New("no port found")
}

func getServerPortsFromQuarkusApplicationYamlFile(file string) ([]int, error) {
	yamlFile, err := os.ReadFile(filepath.Clean(file))
	if err != nil {
		return []int{}, err
	}
	var data model.QuarkusApplicationYaml
	err = yaml.Unmarshal(yamlFile, &data)
	if err != nil {
		return []int{}, err
	}
	var ports []int
	if data.Quarkus.Http.SSLPort > 0 {
		ports = append(ports, data.Quarkus.Http.SSLPort)
	}
	if data.Quarkus.Http.InsecureRequests != "disabled" {
		if data.Quarkus.Http.Port > 0 {
			ports = append(ports, data.Quarkus.Http.Port)
		}
	}
	if len(ports) > 0 {
		return ports, nil
	}
	return []int{}, errors.New("no port found")
}
