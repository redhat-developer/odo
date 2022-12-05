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
	"os"
	"path/filepath"

	"github.com/redhat-developer/alizer/go/pkg/apis/model"
	utils "github.com/redhat-developer/alizer/go/pkg/utils"
	"gopkg.in/yaml.v3"
)

type QuarkusDetector struct{}

type QuarkusApplicationYaml struct {
	Quarkus QuarkusHttp `yaml:"quarkus,omitempty"`
}

type QuarkusHttp struct {
	Http QuarkusHttpPort `yaml:"http,omitempty"`
}

type QuarkusHttpPort struct {
	Port             int    `yaml:"port,omitempty"`
	InsecureRequests string `yaml:"insecure-requests,omitempty"`
	SSLPort          int    `yaml:"ssl-port,omitempty"`
}

func (q QuarkusDetector) GetSupportedFrameworks() []string {
	return []string{"Quarkus"}
}

func (q QuarkusDetector) DoFrameworkDetection(language *model.Language, config string) {
	if hasFwk, _ := hasFramework(config, "io.quarkus"); hasFwk {
		language.Frameworks = append(language.Frameworks, "Quarkus")
	}
}

func (q QuarkusDetector) DoPortsDetection(component *model.Component) {
	// check if port is set on env var
	ports := getQuarkusPortsFromEnvs()
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

func getServerPortsFromQuarkusPropertiesFile(file string) ([]int, error) {
	ports := []int{}
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
	yamlFile, err := ioutil.ReadFile(file)
	if err != nil {
		return []int{}, err
	}
	var data QuarkusApplicationYaml
	yaml.Unmarshal(yamlFile, &data)
	ports := []int{}
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
