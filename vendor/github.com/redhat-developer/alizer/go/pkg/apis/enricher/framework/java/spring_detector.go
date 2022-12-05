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
	"errors"
	"io/ioutil"
	"path/filepath"

	"github.com/redhat-developer/alizer/go/pkg/apis/model"
	utils "github.com/redhat-developer/alizer/go/pkg/utils"
	"gopkg.in/yaml.v3"
)

type SpringDetector struct{}

type ApplicationProsServer struct {
	Server struct {
		Port int `yaml:"port,omitempty"`
		Http struct {
			Port int `yaml:"port,omitempty"`
		} `yaml:"http,omitempty"`
	} `yaml:"server,omitempty"`
}

func (s SpringDetector) GetSupportedFrameworks() []string {
	return []string{"Spring"}
}

func (s SpringDetector) DoFrameworkDetection(language *model.Language, config string) {
	if hasFwk, _ := hasFramework(config, "org.springframework"); hasFwk {
		language.Frameworks = append(language.Frameworks, "Spring")
	}
}

func (s SpringDetector) DoPortsDetection(component *model.Component) {
	// check if port is set on env var
	ports := getSpringPortsFromEnvs()
	if len(ports) > 0 {
		component.Ports = ports
		return
	}

	applicationFile := utils.GetAnyApplicationFilePath(component.Path, []model.ApplicationFileInfo{
		{
			Dir:  "src/main/resources",
			File: "application.properties",
		},
		{
			Dir:  "src/main/resources",
			File: "application.yml",
		},
		{
			Dir:  "src/main/resources",
			File: "application.yaml",
		},
	})
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
	ports := []int{}
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
	yamlFile, err := ioutil.ReadFile(file)
	if err != nil {
		return []int{}, err
	}
	var data ApplicationProsServer
	yaml.Unmarshal(yamlFile, &data)
	ports := []int{}
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
