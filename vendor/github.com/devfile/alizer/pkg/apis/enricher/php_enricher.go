/*******************************************************************************
 * Copyright (c) 2023 Red Hat, Inc.
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
	framework "github.com/devfile/alizer/pkg/apis/enricher/framework/php"
	"github.com/devfile/alizer/pkg/apis/model"
	"github.com/devfile/alizer/pkg/utils"
	langfile "github.com/devfile/alizer/pkg/utils/langfiles"
)

type PHPEnricher struct{}

func getPHPFrameworkDetectors() []FrameworkDetectorWithConfigFile {
	return []FrameworkDetectorWithConfigFile{
		&framework.LaravelDetector{},
	}
}

func (p PHPEnricher) GetSupportedLanguages() []string {
	return []string{"php"}
}

// DoEnrichLanguage runs DoFrameworkDetection with found php project files.
// php project files: composer.json
func (p PHPEnricher) DoEnrichLanguage(language *model.Language, files *[]string) {
	composerJson := utils.GetFile(files, "composer.json")

	if composerJson != "" {
		var targetLanguage string
		if utils.IsTagInComposerJsonFile(composerJson, "php") {
			targetLanguage = "PHP"
		}
		lang, err := langfile.Get().GetLanguageByName(targetLanguage)
		if err == nil {
			language.Name = lang.Name
			language.Aliases = lang.Aliases
		}
		detectPHPFrameworks(language, composerJson)
	}
}

// DoEnrichComponent checks for the port number using a Dockerfile, Compose file, or Source strategy
func (p PHPEnricher) DoEnrichComponent(component *model.Component, settings model.DetectionSettings, ctx *context.Context) {
	projectName := GetDefaultProjectName(component.Path)
	component.Name = projectName

	for _, algorithm := range settings.PortDetectionStrategy {
		var ports []int
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
				for _, detector := range getPHPFrameworkDetectors() {
					for _, framework := range component.Languages[0].Frameworks {
						if utils.Contains(detector.GetSupportedFrameworks(), framework) {
							detector.DoPortsDetection(component, ctx)
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

func (p PHPEnricher) IsConfigValidForComponentDetection(language string, config string) bool {
	return IsConfigurationValidForLanguage(language, config)
}

func detectPHPFrameworks(language *model.Language, configFile string) {
	for _, detector := range getPHPFrameworkDetectors() {
		detector.DoFrameworkDetection(language, configFile)
	}
}
