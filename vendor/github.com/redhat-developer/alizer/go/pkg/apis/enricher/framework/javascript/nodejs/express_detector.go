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
	"os"
	"regexp"
	"strings"

	"github.com/redhat-developer/alizer/go/pkg/apis/model"
	"github.com/redhat-developer/alizer/go/pkg/utils"
)

type ExpressDetector struct{}

func (e ExpressDetector) GetSupportedFrameworks() []string {
	return []string{"Express"}
}

func (e ExpressDetector) DoFrameworkDetection(language *model.Language, config string) {
	if hasFramework(config, "express") {
		language.Frameworks = append(language.Frameworks, "Express")
	}
}

func (e ExpressDetector) DoPortsDetection(component *model.Component) {
	files, err := utils.GetFilePathsFromRoot(component.Path)
	if err != nil {
		return
	}

	re := regexp.MustCompile(`\.listen\([^,)]*`)
	ports := []int{}
	for _, file := range files {
		bytes, err := os.ReadFile(file)
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

func getPort(content string, matchIndexes []int) int {
	portPlaceholder := content[matchIndexes[0]:matchIndexes[1]]
	//we should end up with something like ".listen(PORT"
	portPlaceholder = strings.Replace(portPlaceholder, ".listen(", "", -1)
	// if we are lucky enough portPlaceholder contains a real PORT otherwise it would be a variable/expression
	if port, err := utils.GetValidPort(portPlaceholder); err == nil {
		return port
	}
	// of course we are unlucky ... is it an env variable?
	re := regexp.MustCompile(`process.env.[^ ,)]+`)
	envMatchIndexes := re.FindStringSubmatchIndex(portPlaceholder)
	if len(envMatchIndexes) > 1 {
		envPlaceholder := portPlaceholder[envMatchIndexes[0]:envMatchIndexes[1]]
		// we should end up with something like process.env.PORT
		envPlaceholder = strings.Replace(envPlaceholder, "process.env.", "", -1)
		//envPlaceholder should contain the name of the env variable
		portValue := os.Getenv(envPlaceholder)
		if port, err := utils.GetValidPort(portValue); err == nil {
			return port
		}
	} else {
		// we are not dealing with an env variable, let's try to find a variable set before the listen function
		contentBeforeMatch := content[0:matchIndexes[0]]
		re = regexp.MustCompile(`(let|const|var)\s+` + portPlaceholder + `\s*=\s*([^;]*)`)
		return utils.FindPortSubmatch(re, contentBeforeMatch, 2)
	}
	return -1
}
