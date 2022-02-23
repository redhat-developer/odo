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
	"path/filepath"
	"strings"

	framework "github.com/redhat-developer/alizer/go/pkg/apis/enricher/framework/java"
	"github.com/redhat-developer/alizer/go/pkg/apis/language"
	utils "github.com/redhat-developer/alizer/go/pkg/utils"
)

type JavaEnricher struct{}

func getJavaFrameworkDetectors() []FrameworkDetectorWithConfigFile {
	return []FrameworkDetectorWithConfigFile{
		&framework.MicronautDetector{},
		&framework.OpenLibertyDetector{},
		&framework.QuarkusDetector{},
		&framework.SpringDetector{},
		&framework.VertxDetector{},
	}
}

func (j JavaEnricher) GetSupportedLanguages() []string {
	return []string{"java"}
}

func (j JavaEnricher) DoEnrichLanguage(language *language.Language, files *[]string) {
	gradle := utils.GetFile(files, "build.gradle")
	maven := utils.GetFile(files, "pom.xml")
	ant := utils.GetFile(files, "build.xml")

	if gradle != "" {
		language.Tools = []string{"Gradle"}
		detectJavaFrameworks(language, gradle)
	} else if maven != "" {
		language.Tools = []string{"Maven"}
		detectJavaFrameworks(language, maven)
	} else if ant != "" {
		language.Tools = []string{"Ant"}
	}
}

func (j JavaEnricher) IsConfigValidForComponentDetection(language string, config string) bool {
	return IsConfigurationValidForLanguage(language, config) && !isParentModuleMaven(config)
}

/*
	isParentModuleMaven checks if configuration file is a parent pom.xml
	Parameters:
		configPath: configuration file path
	Returns:
		bool: true if config file is parent
*/
func isParentModuleMaven(configPath string) bool {
	_, file := filepath.Split(configPath)
	if !strings.EqualFold(file, "pom.xml") {
		return false
	}

	hasTag, _ := utils.IsTagInPomXMLFile(configPath, "modules")
	return hasTag
}

func detectJavaFrameworks(language *language.Language, configFile string) {
	for _, detector := range getJavaFrameworkDetectors() {
		detector.DoFrameworkDetection(language, configFile)
	}
}
