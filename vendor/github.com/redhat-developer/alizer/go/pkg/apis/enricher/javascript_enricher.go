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
package recognizer

import (
	"os"
	"path/filepath"

	framework "github.com/redhat-developer/alizer/go/pkg/apis/enricher/framework/javascript/nodejs"
	"github.com/redhat-developer/alizer/go/pkg/apis/model"
	utils "github.com/redhat-developer/alizer/go/pkg/utils"
)

type JavaScriptEnricher struct{}

func getJavaScriptFrameworkDetectors() []FrameworkDetectorWithConfigFile {
	return []FrameworkDetectorWithConfigFile{
		&framework.ExpressDetector{},
		&framework.ReactJsDetector{},
	}
}

func (j JavaScriptEnricher) GetSupportedLanguages() []string {
	return []string{"javascript", "typescript"}
}

func (j JavaScriptEnricher) DoEnrichLanguage(language *model.Language, files *[]string) {
	packageJson := utils.GetFile(files, "package.json")

	if packageJson != "" {
		language.Tools = []string{"NodeJs"}
		detectJavaScriptFrameworks(language, packageJson)
	}
}

func (j JavaScriptEnricher) DoEnrichComponent(component *model.Component) {
	projectName := ""
	packageJsonPath := filepath.Join(component.Path, "package.json")
	if _, err := os.Stat(packageJsonPath); err == nil {
		packageJson, err := utils.GetPackageJsonFile(packageJsonPath)
		if err == nil {
			projectName = packageJson.Name
		}
	}
	if projectName == "" {
		projectName = GetDefaultProjectName(component.Path)
	}
	component.Name = projectName
}

func (j JavaScriptEnricher) IsConfigValidForComponentDetection(language string, config string) bool {
	return IsConfigurationValidForLanguage(language, config)
}

func detectJavaScriptFrameworks(language *model.Language, configFile string) {
	for _, detector := range getJavaScriptFrameworkDetectors() {
		detector.DoFrameworkDetection(language, configFile)
	}
}
