/*******************************************************************************
 * Copyright (c) 2022 Red Hat, Inc.
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

	framework "github.com/redhat-developer/alizer/go/pkg/apis/enricher/framework/go"
	"github.com/redhat-developer/alizer/go/pkg/apis/model"
	utils "github.com/redhat-developer/alizer/go/pkg/utils"
	"golang.org/x/mod/modfile"
)

type GoEnricher struct{}

type GoFrameworkDetector interface {
	GetSupportedFrameworks() []string
	DoFrameworkDetection(language *model.Language, goMod *modfile.File)
	DoPortsDetection(component *model.Component)
}

func getGoFrameworkDetectors() []GoFrameworkDetector {
	return []GoFrameworkDetector{
		&framework.GinDetector{},
		&framework.BeegoDetector{},
		&framework.EchoDetector{},
		&framework.FastHttpDetector{},
		&framework.GoFiberDetector{},
		&framework.MuxDetector{},
	}
}

func (j GoEnricher) GetSupportedLanguages() []string {
	return []string{"go"}
}

func (j GoEnricher) DoEnrichLanguage(language *model.Language, files *[]string) {
	goModPath := utils.GetFile(files, "go.mod")

	if goModPath != "" {
		goModFile, err := getGoModFile(goModPath)
		if err != nil {
			return
		}
		if goModFile.Go != nil {
			language.Tools = []string{goModFile.Go.Version}
		}
		detectGoFrameworks(language, goModFile)
	}
}

func (j GoEnricher) DoEnrichComponent(component *model.Component, settings model.DetectionSettings) {
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
				for _, detector := range getGoFrameworkDetectors() {
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

func (j GoEnricher) IsConfigValidForComponentDetection(language string, config string) bool {
	return IsConfigurationValidForLanguage(language, config)
}

func getGoModFile(filePath string) (*modfile.File, error) {
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, errors.New("unable to read go.mod file")
	}
	return modfile.Parse(filePath, b, nil)
}

func detectGoFrameworks(language *model.Language, configFile *modfile.File) {
	for _, detector := range getGoFrameworkDetectors() {
		detector.DoFrameworkDetection(language, configFile)
	}
}
