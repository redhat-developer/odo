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
	"encoding/json"

	"github.com/devfile/alizer/pkg/apis/model"
	"github.com/devfile/alizer/pkg/utils"
)

type VertxDetector struct{}

func (v VertxDetector) GetSupportedFrameworks() []string {
	return []string{"Vertx"}
}

func (v VertxDetector) GetApplicationFileInfos(componentPath string, ctx *context.Context) []model.ApplicationFileInfo {
	return []model.ApplicationFileInfo{
		{
			Context: ctx,
			Root:    componentPath,
			Dir:     "src/main/conf",
			File:    ".*.json",
		},
	}
}

// DoFrameworkDetection uses the groupId to check for the framework name
func (v VertxDetector) DoFrameworkDetection(language *model.Language, config string) {
	if hasFwk, _ := hasFramework(config, "io.vertx", ""); hasFwk {
		language.Frameworks = append(language.Frameworks, "Vertx")
	}
}

// DoPortsDetection searches for the port in json files under src/main/conf/
func (v VertxDetector) DoPortsDetection(component *model.Component, ctx *context.Context) {
	appFileInfos := v.GetApplicationFileInfos(component.Path, ctx)
	if len(appFileInfos) == 0 {
		return
	}

	for _, appFileInfo := range appFileInfos {
		fileBytes, err := utils.GetApplicationFileBytes(appFileInfo)
		if err != nil {
			continue
		}

		var data model.VertxConf
		err = json.Unmarshal(fileBytes, &data)
		if err != nil {
			continue
		}

		if utils.IsValidPort(data.Port) {
			component.Ports = []int{data.Port}
			return
		}

		if utils.IsValidPort(data.ServerConfig.Port) {
			component.Ports = []int{data.ServerConfig.Port}
			return
		}
	}
}
