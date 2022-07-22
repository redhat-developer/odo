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
package recognizer

import (
	framework "github.com/redhat-developer/alizer/go/pkg/apis/enricher/framework/dotnet"
	"github.com/redhat-developer/alizer/go/pkg/apis/model"
	utils "github.com/redhat-developer/alizer/go/pkg/utils"
)

type DotNetEnricher struct{}

func (j DotNetEnricher) GetSupportedLanguages() []string {
	return []string{"c#", "f#", "visual basic .net"}
}

func getDotNetFrameworkDetectors() []FrameworkDetectorWithConfigFile {
	return []FrameworkDetectorWithConfigFile{
		&framework.DotNetDetector{},
	}
}

func (j DotNetEnricher) DoEnrichLanguage(language *model.Language, files *[]string) {
	configFiles := utils.GetFilesByRegex(files, ".*\\.\\w+proj")
	for _, configFile := range configFiles {
		getDotNetFrameworks(language, configFile)
	}
}

func (j DotNetEnricher) DoEnrichComponent(component *model.Component) {
	projectName := GetDefaultProjectName(component.Path)
	component.Name = projectName
}

func (j DotNetEnricher) IsConfigValidForComponentDetection(language string, config string) bool {
	return IsConfigurationValidForLanguage(language, config)
}

func getDotNetFrameworks(language *model.Language, configFile string) {
	for _, detector := range getDotNetFrameworkDetectors() {
		detector.DoFrameworkDetection(language, configFile)
	}
}
