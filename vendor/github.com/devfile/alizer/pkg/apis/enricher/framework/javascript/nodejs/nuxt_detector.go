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

package enricher

import (
	"context"
	"regexp"

	"github.com/devfile/alizer/pkg/apis/model"
	"github.com/devfile/alizer/pkg/utils"
)

type NuxtDetector struct{}

func (n NuxtDetector) GetSupportedFrameworks() []string {
	return []string{"Nuxt"}
}

func (n NuxtDetector) GetApplicationFileInfos(componentPath string, ctx *context.Context) []model.ApplicationFileInfo {
	return []model.ApplicationFileInfo{
		{
			Context: ctx,
			Root:    componentPath,
			Dir:     "",
			File:    "nuxt.config.js",
		},
	}
}

// DoFrameworkDetection uses a tag to check for the framework name
func (n NuxtDetector) DoFrameworkDetection(language *model.Language, config string) {
	if hasFramework(config, "nuxt") {
		language.Frameworks = append(language.Frameworks, "Nuxt", "Nuxt.js")
	}
}

// DoPortsDetection searches for the port in package.json, and nuxt.config.js
func (n NuxtDetector) DoPortsDetection(component *model.Component, ctx *context.Context) {
	ports := []int{}
	regexes := []string{`--port=(\d*)`}
	// check if port is set in start script in package.json
	port := getPortFromStartScript(component.Path, regexes)
	if utils.IsValidPort(port) {
		component.Ports = []int{port}
		return
	}

	// check if port is set in dev script in package.json
	port = getPortFromDevScript(component.Path, regexes)
	if utils.IsValidPort(port) {
		component.Ports = []int{port}
		return
	}

	//check inside the nuxt.config.js file
	appFileInfos := n.GetApplicationFileInfos(component.Path, ctx)
	if len(appFileInfos) == 0 {
		return
	}

	for _, appFileInfo := range appFileInfos {
		fileBytes, err := utils.GetApplicationFileBytes(appFileInfo)
		if err != nil {
			continue
		}

		re := regexp.MustCompile(`port:\s*(\d+)*`)
		ports = utils.FindAllPortsSubmatch(re, string(fileBytes), 1)
		if len(ports) > 0 {
			component.Ports = ports
			return
		}
	}
}
