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
	"regexp"
	"strings"

	"github.com/devfile/alizer/pkg/apis/model"
	"github.com/devfile/alizer/pkg/utils"
)

type ExpressDetector struct{}

func (e ExpressDetector) GetSupportedFrameworks() []string {
	return []string{"Express"}
}

// DoFrameworkDetection uses a tag to check for the framework name
func (e ExpressDetector) DoFrameworkDetection(language *model.Language, config string) {
	if hasFramework(config, "express") {
		language.Frameworks = append(language.Frameworks, "Express")
	}
}

func (e ExpressDetector) DoPortsDetection(component *model.Component, ctx *context.Context) {
	files, err := utils.GetCachedFilePathsFromRoot(component.Path, ctx)
	if err != nil {
		return
	}

	re := regexp.MustCompile(`\.listen\([^,)]*`)
	var ports []int
	for _, file := range files {
		cleanFile := filepath.Clean(file)
		bytes, err := os.ReadFile(cleanFile)
		if err != nil {
			continue
		}
		content := string(bytes)
		matchesIndexes := re.FindAllStringSubmatchIndex(content, -1)
		for _, matchIndexes := range matchesIndexes {
			port := getPort(content, matchIndexes)
			if port != -1 {
				ports = append(ports, port)
			}
		}
		if len(ports) > 0 {
			component.Ports = ports
			return
		}
	}
}

func getPortGroup(content string, matchIndexes []int, portPlaceholder string) string {
	contentBeforeMatch := content[0:matchIndexes[0]]
	re, err := regexp.Compile(`(let|const|var)\s+` + portPlaceholder + `\s*=\s*([^;]*)`)
	if err != nil {
		return ""
	}
	return utils.FindPotentialPortGroup(re, contentBeforeMatch, 2)
}

func GetEnvPort(envPlaceholder string) int {
	envPlaceholder = strings.Replace(envPlaceholder, "process.env.", "", -1)
	portValue := os.Getenv(envPlaceholder)
	if port, err := utils.GetValidPort(portValue); err == nil {
		return port
	}
	return -1
}

func getPort(content string, matchIndexes []int) int {
	// Express configures its port with app.listen()
	portPlaceholder := content[matchIndexes[0]:matchIndexes[1]]
	portPlaceholder = strings.Replace(portPlaceholder, ".listen(", "", -1)

	// Case: Raw port value -> return it directly
	if port, err := utils.GetValidPort(portPlaceholder); err == nil {
		return port
	}

	// Case: Env var given as value in app.listen -> Get env value
	// example: app.listen(process.env.PORT...
	re := regexp.MustCompile(`process.env.[^ ,)]+`)
	envMatchIndexes := re.FindStringSubmatchIndex(portPlaceholder)
	envPortValue := portPlaceholder
	// If no match was found check if port is a variable assigned elsewhere in the code
	if len(envMatchIndexes) == 0 {
		// Case: Var Port with env var as value
		potentialPortGroup := getPortGroup(content, matchIndexes, portPlaceholder)
		if potentialPortGroup != "" {
			// Takes into account cases like -> var PORT = process.env.PORT || 8080
			portValues := strings.Split(potentialPortGroup, " || ")
			for _, portValue := range portValues {
				re = regexp.MustCompile(`process.env.[^ ,)]+`)
				tmpMatchIndexes := re.FindStringSubmatchIndex(portValue)
				// If there is any matching update the env values
				if len(tmpMatchIndexes) > 1 {
					envMatchIndexes = tmpMatchIndexes
					envPortValue = portValue
				}
			}
		}
	}
	// After double-checking for env vars try to get the value of this port
	if len(envMatchIndexes) > 1 {
		envPlaceholder := envPortValue[envMatchIndexes[0]:envMatchIndexes[1]]
		port := GetEnvPort(envPlaceholder)
		// The port will be return only if a value was found for the given env var
		if port > 0 {
			return port
		}
	}
	// Case: No env var or raw value found -> check for raw value into a var
	potentialPortGroup := getPortGroup(content, matchIndexes, portPlaceholder)
	if potentialPortGroup != "" {
		// Takes into account cases like -> var PORT = process.env.PORT || 8080
		portValues := strings.Split(potentialPortGroup, " || ")
		for _, portValue := range portValues {
			if port, err := utils.GetValidPort(portValue); err == nil {
				return port
			}
		}
	}
	return -1
}
