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

type WildFlyDetector struct{}

func (w WildFlyDetector) GetSupportedFrameworks() []string {
	return []string{"WildFly"}
}

func (w WildFlyDetector) GetApplicationFileInfos(componentPath string, ctx *context.Context) []model.ApplicationFileInfo {
	files, err := utils.GetCachedFilePathsFromRoot(componentPath, ctx)
	if err != nil {
		return []model.ApplicationFileInfo{}
	}
	pomXML := utils.GetFile(&files, "pom.xml")
	return utils.GenerateApplicationFileFromFilters([]string{pomXML}, componentPath, "", ctx)
}

// DoFrameworkDetection uses the groupId and artifactId to check for the framework name
func (w WildFlyDetector) DoFrameworkDetection(language *model.Language, config string) {
	if hasFwk, _ := hasFramework(config, "org.wildfly.plugins", "wildfly-maven-plugin"); hasFwk {
		language.Frameworks = append(language.Frameworks, "WildFly")
	}
}

// DoPortsDetection for wildfly fetches the pom.xml and tries to find any javaOpts under
// the wildfly-maven-plugin profiles. If there is one it looks if jboss.http.port is defined.
func (w WildFlyDetector) DoPortsDetection(component *model.Component, ctx *context.Context) {
	ports := []int{}
	// Fetch the content of xml for this component
	appFileInfos := w.GetApplicationFileInfos(component.Path, ctx)
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

		portPlaceholder := GetPortsForJBossFrameworks(pom, "wildfly-maven-plugin", "org.wildfly.plugins")
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
