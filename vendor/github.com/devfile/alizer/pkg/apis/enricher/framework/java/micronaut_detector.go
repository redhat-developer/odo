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
	"os"

	"github.com/devfile/alizer/pkg/apis/model"
	"github.com/devfile/alizer/pkg/utils"
	"gopkg.in/yaml.v3"
)

type MicronautDetector struct{}

func (m MicronautDetector) GetSupportedFrameworks() []string {
	return []string{"Micronaut"}
}

func (m MicronautDetector) GetApplicationFileInfos(componentPath string, ctx *context.Context) []model.ApplicationFileInfo {
	return []model.ApplicationFileInfo{
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
func (m MicronautDetector) DoFrameworkDetection(language *model.Language, config string) {
	if hasFwk, _ := hasFramework(config, "io.micronaut", ""); hasFwk {
		language.Frameworks = append(language.Frameworks, "Micronaut")
	}
}

// DoPortsDetection searches for the port in src/main/resources/application.yaml
func (m MicronautDetector) DoPortsDetection(component *model.Component, ctx *context.Context) {
	// check if port is set on env var
	ports := getMicronautPortsFromEnvs()
	if len(ports) > 0 {
		component.Ports = ports
		return
	}

	// check if port is set on dockerfile as env var
	ports = getMicronautPortsFromEnvDockerfile(component.Path)
	if len(ports) > 0 {
		component.Ports = ports
		return
	}

	// check source code
	appFileInfos := m.GetApplicationFileInfos(component.Path, ctx)
	if len(appFileInfos) == 0 {
		return
	}

	for _, appFileInfo := range appFileInfos {
		fileBytes, err := utils.GetApplicationFileBytes(appFileInfo)
		if err != nil {
			continue
		}

		ports = getMicronautPortsFromBytes(fileBytes)
		if len(ports) > 0 {
			component.Ports = ports
			return
		}
	}
}

func getMicronautPortsFromBytes(bytes []byte) []int {
	var ports []int
	var data model.MicronautApplicationProps
	err := yaml.Unmarshal(bytes, &data)
	if err != nil {
		return []int{}
	}
	if data.Micronaut.Server.SSL.Enabled && utils.IsValidPort(data.Micronaut.Server.SSL.Port) {
		ports = append(ports, data.Micronaut.Server.SSL.Port)
	}
	if utils.IsValidPort(data.Micronaut.Server.Port) {
		ports = append(ports, data.Micronaut.Server.Port)
	}
	return ports
}

func getMicronautPortsFromEnvs() []int {
	sslEnabled := os.Getenv("MICRONAUT_SERVER_SSL_ENABLED")
	envs := []string{"MICRONAUT_SERVER_PORT"}
	if sslEnabled == "true" {
		envs = append(envs, "MICRONAUT_SERVER_SSL_PORT")
	}
	return utils.GetValidPortsFromEnvs(envs)
}

func getMicronautPortsFromEnvDockerfile(path string) []int {
	envVars, err := utils.GetEnvVarsFromDockerFile(path)
	if err != nil {
		return nil
	}
	sslEnabled := ""
	envs := []string{"MICRONAUT_SERVER_PORT"}
	for _, envVar := range envVars {
		if envVar.Name == "MICRONAUT_SERVER_SSL_ENABLED" {
			sslEnabled = envVar.Value
			break
		}
	}

	if sslEnabled == "true" {
		envs = append(envs, "MICRONAUT_SERVER_SSL_PORT")
	}

	return utils.GetValidPortsFromEnvDockerfile(envs, envVars)
}
