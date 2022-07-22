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
	"regexp"
	"strings"

	framework "github.com/redhat-developer/alizer/go/pkg/apis/enricher/framework/java"
	"github.com/redhat-developer/alizer/go/pkg/apis/model"
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

func (j JavaEnricher) DoEnrichLanguage(language *model.Language, files *[]string) {
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

func (j JavaEnricher) DoEnrichComponent(component *model.Component) {
	projectName := getProjectNameMaven(component.Path)
	if projectName == "" {
		projectName = getProjectNameGradle(component.Path)
	}
	if projectName == "" {
		projectName = GetDefaultProjectName(component.Path)
	}
	component.Name = projectName
}

func getProjectNameGradle(root string) string {
	settingsGradlePath := filepath.Join(root, "settings.gradle")
	if _, err := os.Stat(settingsGradlePath); err == nil {
		re := regexp.MustCompile(`rootProject.name\s*=\s*(.*)`)
		bytes, err := os.ReadFile(settingsGradlePath)
		if err != nil {
			return ""
		}
		content := string(bytes)
		matchProjectName := re.FindStringSubmatch(content)
		if len(matchProjectName) > 0 && matchProjectName[1] != "" {
			projectName := strings.TrimLeft(matchProjectName[1], "\"'")
			projectName = strings.TrimRight(projectName, "\"' ")
			return projectName
		}
	}
	return ""
}

func getProjectNameMaven(root string) string {
	pomXMLPath := filepath.Join(root, "pom.xml")
	if _, err := os.Stat(pomXMLPath); err == nil {
		pomXML, err := utils.GetPomFileContent(pomXMLPath)
		if err == nil {
			return pomXML.ArtifactId
		}
	}
	return ""
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

func detectJavaFrameworks(language *model.Language, configFile string) {
	for _, detector := range getJavaFrameworkDetectors() {
		detector.DoFrameworkDetection(language, configFile)
	}
}
