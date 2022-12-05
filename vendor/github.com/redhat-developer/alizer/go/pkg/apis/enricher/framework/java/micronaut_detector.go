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
	"os"

	"github.com/redhat-developer/alizer/go/pkg/apis/model"
	utils "github.com/redhat-developer/alizer/go/pkg/utils"
	"gopkg.in/yaml.v3"
)

type MicronautDetector struct{}

type MicronautApplicationProps struct {
	Micronaut struct {
		Server struct {
			Port int `yaml:"port,omitempty"`
			SSL  struct {
				Enabled bool `yaml:"enabled,omitempty"`
				Port    int  `yaml:"port,omitempty"`
			} `yaml:"ssl,omitempty"`
		} `yaml:"server,omitempty"`
	} `yaml:"micronaut,omitempty"`
}

func (m MicronautDetector) GetSupportedFrameworks() []string {
	return []string{"Micronaut"}
}

func (m MicronautDetector) DoFrameworkDetection(language *model.Language, config string) {
	if hasFwk, _ := hasFramework(config, "io.micronaut"); hasFwk {
		language.Frameworks = append(language.Frameworks, "Micronaut")
	}
}

func (m MicronautDetector) DoPortsDetection(component *model.Component) {
	// check if port is set on env var
	ports := getMicronautPortsFromEnvs()
	if len(ports) > 0 {
		component.Ports = ports
		return
	}

	bytes, err := utils.ReadAnyApplicationFile(component.Path, []model.ApplicationFileInfo{
		{
			Dir:  "src/main/resources",
			File: "application.yml",
		},
		{
			Dir:  "src/main/resources",
			File: "application.yaml",
		},
	})
	if err != nil {
		return
	}
	ports = getMicronautPortsFromBytes(bytes)
	if len(ports) > 0 {
		component.Ports = ports
	}
}

func getMicronautPortsFromBytes(bytes []byte) []int {
	ports := []int{}
	var data MicronautApplicationProps
	yaml.Unmarshal(bytes, &data)
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
