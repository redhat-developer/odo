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
	"encoding/json"

	"github.com/devfile/alizer/pkg/apis/model"
	"github.com/devfile/alizer/pkg/utils"
)

type AngularDetector struct{}

func (a AngularDetector) GetSupportedFrameworks() []string {
	return []string{"Angular"}
}

func (a AngularDetector) GetApplicationFileInfos(componentPath string, ctx *context.Context) []model.ApplicationFileInfo {
	return []model.ApplicationFileInfo{
		{
			Context: ctx,
			Root:    componentPath,
			Dir:     "",
			File:    "angular.json",
		},
		{
			Context: ctx,
			Root:    componentPath,
			Dir:     "",
			File:    "angular-cli.json",
		},
	}
}

// DoFrameworkDetection uses a tag to check for the framework name
func (a AngularDetector) DoFrameworkDetection(language *model.Language, config string) {
	if hasFramework(config, "angular") {
		language.Frameworks = append(language.Frameworks, "Angular")
	}
}

// DoPortsDetection searches for the port in angular.json, package.json, and angular-cli.json
func (a AngularDetector) DoPortsDetection(component *model.Component, ctx *context.Context) {
	// check if port is set on angular.json file
	appFileInfos := a.GetApplicationFileInfos(component.Path, ctx)
	if len(appFileInfos) == 0 {
		return
	}

	appFileInfo, err := utils.GetApplicationFileInfo(appFileInfos, "angular.json")
	if err != nil {
		return
	}

	fileBytes, err := utils.GetApplicationFileBytes(appFileInfo)
	if err != nil {
		return
	}

	if err != nil {
		return
	}
	var data model.AngularJson
	err = json.Unmarshal(fileBytes, &data)
	if err != nil {
		return
	}

	if projectBody, exists := data.Projects[component.Name]; exists {
		port := projectBody.Architect.Serve.Options.Port
		if utils.IsValidPort(port) {
			component.Ports = []int{port}
			return
		}
	}

	// check if port is set in start script in package.json
	port := getPortFromStartScript(component.Path, []string{`--port (\d*)`})
	if utils.IsValidPort(port) {
		component.Ports = []int{port}
		return
	}

	// check if port is set on angular-cli.json file
	appFileInfoCli, err := utils.GetApplicationFileInfo(appFileInfos, "angular-cli.json")
	if err != nil {
		return
	}

	fileBytesCli, err := utils.GetApplicationFileBytes(appFileInfoCli)
	if err != nil {
		return
	}

	var dataCli model.AngularCliJson
	err = json.Unmarshal(fileBytesCli, &dataCli)
	if err != nil {
		return
	}

	port = dataCli.Defaults.Serve.Port
	if utils.IsValidPort(port) {
		component.Ports = []int{port}
		return
	}
}
