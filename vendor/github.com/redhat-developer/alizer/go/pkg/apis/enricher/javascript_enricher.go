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
	"path/filepath"

	framework "github.com/redhat-developer/alizer/go/pkg/apis/enricher/framework/javascript/nodejs"
	"github.com/redhat-developer/alizer/go/pkg/apis/model"
	"github.com/redhat-developer/alizer/go/pkg/utils"
	langfile "github.com/redhat-developer/alizer/go/pkg/utils/langfiles"
)

type JavaScriptEnricher struct{}

func getJavaScriptFrameworkDetectors() []FrameworkDetectorWithConfigFile {
	return []FrameworkDetectorWithConfigFile{
		&framework.AngularDetector{},
		&framework.ExpressDetector{},
		&framework.NextDetector{},
		&framework.NuxtDetector{},
		&framework.ReactJsDetector{},
		&framework.SvelteDetector{},
		&framework.VueDetector{},
	}
}

func (j JavaScriptEnricher) GetSupportedLanguages() []string {
	return []string{"javascript", "typescript"}
}

// DoEnrichLanguage runs DoFrameworkDetection with found javascript project files.
// javascript project files: package.json
func (j JavaScriptEnricher) DoEnrichLanguage(language *model.Language, files *[]string) {
	packageJson := utils.GetFile(files, "package.json")

	if packageJson != "" {
		language.Tools = []string{"NodeJs", "Node.js"}
		var targetLanguage string
		if utils.IsTagInPackageJsonFile(packageJson, "typescript") || utils.IsTagInPackageJsonFile(packageJson, "tslib") {
			targetLanguage = "TypeScript"
		} else {
			targetLanguage = "JavaScript"
		}
		lang, err := langfile.Get().GetLanguageByName(targetLanguage)
		if err == nil {
			language.Name = lang.Name
			language.Aliases = lang.Aliases
		}
		detectJavaScriptFrameworks(language, packageJson)
	}
}

// DoEnrichComponent checks for the port number using a Dockerfile, Compose file, or Source strategy
func (j JavaScriptEnricher) DoEnrichComponent(component *model.Component, settings model.DetectionSettings, ctx *context.Context) {
	projectName := ""
	packageJsonPath := filepath.Join(component.Path, "package.json")
	if _, err := os.Stat(packageJsonPath); err == nil {
		packageJson, err := utils.GetPackageJsonSchemaFromFile(packageJsonPath)
		if err == nil {
			projectName = packageJson.Name
		}
	}
	if projectName == "" {
		projectName = GetDefaultProjectName(component.Path)
	}
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
				for _, detector := range getJavaScriptFrameworkDetectors() {
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

func (j JavaScriptEnricher) IsConfigValidForComponentDetection(language string, config string) bool {
	return IsConfigurationValidForLanguage(language, config)
}

func detectJavaScriptFrameworks(language *model.Language, configFile string) {
	for _, detector := range getJavaScriptFrameworkDetectors() {
		detector.DoFrameworkDetection(language, configFile)
	}
}
