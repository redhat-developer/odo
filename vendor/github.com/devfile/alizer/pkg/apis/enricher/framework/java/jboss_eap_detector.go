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
	"github.com/devfile/alizer/pkg/schema"
	"github.com/devfile/alizer/pkg/utils"
)

type JBossEAPDetector struct{}

func (o JBossEAPDetector) GetSupportedFrameworks() []string {
	return []string{"JBoss EAP"}
}

func (o JBossEAPDetector) GetApplicationFileInfos(componentPath string, ctx *context.Context) []model.ApplicationFileInfo {
	return []model.ApplicationFileInfo{
		{
			Context: ctx,
			Root:    componentPath,
			Dir:     "",
			File:    "pom.xml",
		},
	}
}

// DoFrameworkDetection uses the groupId and artifactId to check for the framework name
func (o JBossEAPDetector) DoFrameworkDetection(language *model.Language, config string) {
	if hasFwk, _ := hasFramework(config, "org.jboss.eap.plugins", "eap-maven-plugin"); hasFwk {
		language.Frameworks = append(language.Frameworks, "JBoss EAP")
	}
}

func (o JBossEAPDetector) DoPortsDetection(component *model.Component, ctx *context.Context) {
	ports := []int{}
	// Fetch the content of xml for this component
	appFileInfos := o.GetApplicationFileInfos(component.Path, ctx)
	if len(appFileInfos) == 0 {
		return
	}

	for _, appFileInfo := range appFileInfos {
		fileBytes, err := utils.GetApplicationFileBytes(appFileInfo)
		if err != nil {
			continue
		}

		var pom schema.Pom
		err = xml.Unmarshal(fileBytes, &pom)
		if err != nil {
			continue
		}

		portPlaceholder := GetPortsForJBossFrameworks(pom, "eap-maven-plugin", "org.jboss.eap.plugins")
		if portPlaceholder == "" {
			continue
		}

		if port, err := utils.GetValidPort(portPlaceholder); err == nil {
			ports = append(ports, port)
		}

		if len(ports) > 0 {
			component.Ports = ports
			return
		}
	}
}
