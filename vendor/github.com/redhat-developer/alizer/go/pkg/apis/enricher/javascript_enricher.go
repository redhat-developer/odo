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
	framework "github.com/redhat-developer/alizer/go/pkg/apis/enricher/framework/javascript/nodejs"
	"github.com/redhat-developer/alizer/go/pkg/apis/language"
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

func (j JavaScriptEnricher) DoEnrichLanguage(language *language.Language, files *[]string) {
	packageJson := utils.GetFile(files, "package.json")

	if packageJson != "" {
		language.Tools = []string{"NodeJs"}
		detectJavaScriptFrameworks(language, packageJson)
	}
}

func (j JavaScriptEnricher) IsConfigValidForComponentDetection(language string, config string) bool {
	return IsConfigurationValidForLanguage(language, config)
}

func detectJavaScriptFrameworks(language *language.Language, configFile string) {
	for _, detector := range getJavaScriptFrameworkDetectors() {
		detector.DoFrameworkDetection(language, configFile)
	}
}
