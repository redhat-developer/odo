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
	"encoding/xml"

	"github.com/devfile/alizer/pkg/apis/model"
	"github.com/devfile/alizer/pkg/utils"
)

type OpenLibertyDetector struct{}

func (o OpenLibertyDetector) GetSupportedFrameworks() []string {
	return []string{"OpenLiberty"}
}

func (o OpenLibertyDetector) GetApplicationFileInfos(componentPath string, ctx *context.Context) []model.ApplicationFileInfo {
	return []model.ApplicationFileInfo{
		{
			Context: ctx,
			Root:    componentPath,
			Dir:     "",
			File:    "server.xml",
		},
		{
			Context: ctx,
			Root:    componentPath,
			Dir:     "src/main/liberty/config",
			File:    "server.xml",
		},
	}
}

// DoFrameworkDetection uses the groupId to check for the framework name
func (o OpenLibertyDetector) DoFrameworkDetection(language *model.Language, config string) {
	if hasFwk, _ := hasFramework(config, "io.openliberty", ""); hasFwk {
		language.Frameworks = append(language.Frameworks, "OpenLiberty")
	}
}

// DoPortsDetection searches for the port in src/main/liberty/config/server.xml and /server.xml
func (o OpenLibertyDetector) DoPortsDetection(component *model.Component, ctx *context.Context) {
	appFileInfos := o.GetApplicationFileInfos(component.Path, ctx)
	if len(appFileInfos) == 0 {
		return
	}

	for _, appFileInfo := range appFileInfos {
		fileBytes, err := utils.GetApplicationFileBytes(appFileInfo)
		if err != nil {
			continue
		}

		var data model.OpenLibertyServerXml
		err = xml.Unmarshal(fileBytes, &data)
		if err != nil {
			continue
		}
		ports := utils.GetValidPorts([]string{data.HttpEndpoint.HttpPort, data.HttpEndpoint.HttpsPort})
		if len(ports) > 0 {
			component.Ports = ports
			return
		}
	}
}
