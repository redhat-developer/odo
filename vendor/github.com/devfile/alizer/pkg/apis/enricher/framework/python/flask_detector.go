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

func (d FlaskDetector) GetSupportedFrameworks() []string {
	return []string{"Flask"}
}

// DoFrameworkDetection uses a tag to check for the framework name
// with flask files and flask config files
func (d FlaskDetector) DoFrameworkDetection(language *model.Language, files *[]string) {
	appPy := utils.GetFile(files, "app.py")
	wsgiPy := utils.GetFile(files, "wsgi.py")
	requirementsTxt := utils.GetFile(files, "requirements.txt")
	projectToml := utils.GetFile(files, "pyproject.toml")

	flaskFiles := []string{}
	configFlaskFiles := []string{}
	utils.AddToArrayIfValueExist(&flaskFiles, appPy)
	utils.AddToArrayIfValueExist(&flaskFiles, wsgiPy)
	utils.AddToArrayIfValueExist(&configFlaskFiles, requirementsTxt)
	utils.AddToArrayIfValueExist(&configFlaskFiles, projectToml)

	if hasFramework(&flaskFiles, "from flask ") || hasFramework(&configFlaskFiles, "Flask") || hasFramework(&configFlaskFiles, "flask") {
		language.Frameworks = append(language.Frameworks, "Flask")
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

// DoPortsDetection searches for the port in app/__init__.py, app.py or /wsgi.py
func (d FlaskDetector) DoPortsDetection(component *model.Component, ctx *context.Context) {
	bytes, err := utils.ReadAnyApplicationFile(component.Path, []model.ApplicationFileInfo{
		{
			Dir:  "",
			File: "app.py",
		},
		{
			Dir:  "",
			File: "wsgi.py",
		},
		{
			Dir:  "app",
			File: "__init__.py",
		},
	}, ctx)

	matchIndexRegexes := []model.PortMatchRule{
		{
			Regex:     regexp.MustCompile(`.run\([^)]*`),
			ToReplace: ".run(",
		},
	}
	if err != nil {
		return
	}
	ports := getPortFromFileFlask(matchIndexRegexes, string(bytes))
	if len(ports) > 0 {
		component.Ports = ports
		return
	}
}
