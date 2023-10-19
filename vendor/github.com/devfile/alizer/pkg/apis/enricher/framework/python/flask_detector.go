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
	"regexp"
	"strings"

	"github.com/devfile/alizer/pkg/apis/model"
	"github.com/devfile/alizer/pkg/utils"
)

type FlaskDetector struct{}

func (f FlaskDetector) GetSupportedFrameworks() []string {
	return []string{"Flask"}
}

func (f FlaskDetector) GetApplicationFileInfos(componentPath string, ctx *context.Context) []model.ApplicationFileInfo {
	return []model.ApplicationFileInfo{
		{
			Context: ctx,
			Root:    componentPath,
			Dir:     "",
			File:    "app.py",
		},
		{
			Context: ctx,
			Root:    componentPath,
			Dir:     "",
			File:    "wsgi.py",
		},
		{
			Context: ctx,
			Root:    componentPath,
			Dir:     "app",
			File:    "__init__.py",
		},
	}
}

func (f FlaskDetector) GetFlaskFilenames() []string {
	return []string{"app.py", "wsgi.py"}
}

func (f FlaskDetector) GetConfigFlaskFilenames() []string {
	return []string{"requirements.txt", "pyproject.toml"}
}

// DoFrameworkDetection uses a tag to check for the framework name
// with flask files and flask config files
func (f FlaskDetector) DoFrameworkDetection(language *model.Language, files *[]string) {
	var flaskFiles []string
	var configFlaskFiles []string

	for _, filename := range f.GetFlaskFilenames() {
		filePy := utils.GetFile(files, filename)
		utils.AddToArrayIfValueExist(&flaskFiles, filePy)
	}

	for _, filename := range f.GetConfigFlaskFilenames() {
		configFile := utils.GetFile(files, filename)
		utils.AddToArrayIfValueExist(&configFlaskFiles, configFile)
	}

	if hasFramework(&flaskFiles, "from flask ") || hasFramework(&configFlaskFiles, "Flask") || hasFramework(&configFlaskFiles, "flask") {
		language.Frameworks = append(language.Frameworks, "Flask")
	}
}

// DoPortsDetection searches for the port in app/__init__.py, app.py or /wsgi.py
func (f FlaskDetector) DoPortsDetection(component *model.Component, ctx *context.Context) {
	appFileInfos := f.GetApplicationFileInfos(component.Path, ctx)
	if len(appFileInfos) == 0 {
		return
	}

	for _, appFileInfo := range appFileInfos {
		fileBytes, err := utils.GetApplicationFileBytes(appFileInfo)
		if err != nil {
			continue
		}

		matchIndexRegexes := []model.PortMatchRule{
			{
				Regex:     regexp.MustCompile(`.run\([^)]*`),
				ToReplace: ".run(",
			},
		}
		if err != nil {
			continue
		}
		ports := getPortFromFileFlask(matchIndexRegexes, string(fileBytes))
		if len(ports) > 0 {
			component.Ports = ports
			return
		}
	}
}

// getPortFromFileFlask tries to find a port configuration inside a given file content
func getPortFromFileFlask(matchIndexRegexes []model.PortMatchRule, text string) []int {
	var ports []int
	for _, matchIndexRegex := range matchIndexRegexes {
		matchIndexesSlice := matchIndexRegex.Regex.FindAllStringSubmatchIndex(text, -1)
		for _, matchIndexes := range matchIndexesSlice {
			if len(matchIndexes) > 1 {
				port := getPortWithMatchIndexesFlask(text, matchIndexes, matchIndexRegex.ToReplace)
				if port != -1 {
					ports = append(ports, port)
				}
			}
		}
	}

	return ports
}

func getPortWithMatchIndexesFlask(content string, matchIndexes []int, toBeReplaced string) int {
	// select the correct range for placeholder and remove unnecessary strings
	portPlaceholder := content[matchIndexes[0]:matchIndexes[1]]
	portPlaceholder = strings.Replace(portPlaceholder, toBeReplaced, "", -1)
	// try first to check for hardcoded ports inside the app.run command. e.g port=3001
	re, err := regexp.Compile(`port=*(\d+)`)
	if err != nil {
		return -1
	}
	if port := utils.FindPortSubmatch(re, portPlaceholder, 1); port != -1 {
		return port
	}

	// get the value of kwarg "port" inside app.run and check if there is
	// any variable defined with its value
	contentBeforeMatch := content[0:matchIndexes[0]]
	portPlaceholder = strings.Replace(portPlaceholder, "port=", "", -1)
	re, err = regexp.Compile(portPlaceholder + `\s=\s*(\d+)`)
	if err != nil {
		return -1
	}
	matches := re.FindStringSubmatch(contentBeforeMatch)
	if len(matches) > 0 {
		portValue := matches[len(matches)-1]
		re, err = regexp.Compile(`:*(\d+)$`)
		if err != nil {
			return -1
		}
		if port := utils.FindPortSubmatch(re, portValue, 1); port != -1 {
			return port
		}
	}

	return -1
}
