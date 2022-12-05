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
	framework "github.com/redhat-developer/alizer/go/pkg/apis/enricher/framework/python"
	"github.com/redhat-developer/alizer/go/pkg/apis/model"
	utils "github.com/redhat-developer/alizer/go/pkg/utils"
)

type PythonEnricher struct{}

func getPythonFrameworkDetectors() []FrameworkDetectorWithoutConfigFile {
	return []FrameworkDetectorWithoutConfigFile{
		&framework.DjangoDetector{},
	}
}

func (p PythonEnricher) GetSupportedLanguages() []string {
	return []string{"python"}
}

func (p PythonEnricher) DoEnrichLanguage(language *model.Language, files *[]string) {
	language.Tools = []string{}
	detectPythonFrameworks(language, files)
}

func (j PythonEnricher) DoEnrichComponent(component *model.Component, settings model.DetectionSettings) {
	projectName := GetDefaultProjectName(component.Path)
	component.Name = projectName

	for _, algorithm := range settings.PortDetectionStrategy {
		ports := []int{}
		switch algorithm {
		case model.DockerFile:
			{
				ports = GetPortsFromDockerFile(component.Path)
				break
			}
		case model.Compose:
			{
				ports = GetPortsFromDockerComposeFile(component.Path, settings)
				break
			}
		case model.Source:
			{
				for _, detector := range getPythonFrameworkDetectors() {
					for _, framework := range component.Languages[0].Frameworks {
						if utils.Contains(detector.GetSupportedFrameworks(), framework) {
							detector.DoPortsDetection(component)
						}
					}
				}
			}
		}
		if len(ports) > 0 {
			component.Ports = ports
		}
		if len(component.Ports) > 0 {
			return
		}
	}
}

func (p PythonEnricher) IsConfigValidForComponentDetection(language string, config string) bool {
	return IsConfigurationValidForLanguage(language, config)
}

func detectPythonFrameworks(language *model.Language, files *[]string) {
	for _, detector := range getPythonFrameworkDetectors() {
		detector.DoFrameworkDetection(language, files)
	}
}
