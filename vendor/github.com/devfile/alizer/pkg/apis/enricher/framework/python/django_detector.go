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

	"github.com/devfile/alizer/pkg/apis/model"
	"github.com/devfile/alizer/pkg/utils"
)

type DjangoDetector struct{}

func (d DjangoDetector) GetSupportedFrameworks() []string {
	return []string{"Django"}
}

func (d DjangoDetector) GetApplicationFileInfos(componentPath string, ctx *context.Context) []model.ApplicationFileInfo {
	return []model.ApplicationFileInfo{
		{
			Context: ctx,
			Root:    componentPath,
			Dir:     "",
			File:    "manage.py",
		},
	}
}

func (d DjangoDetector) GetDjangoFilenames() []string {
	return []string{"manage.py", "urls.py", "wsgi.py", "asgi.py"}
}

func (d DjangoDetector) GetConfigDjangoFilenames() []string {
	return []string{"requirements.txt", "pyproject.toml"}
}

// DoFrameworkDetection uses a tag to check for the framework name
// with django files and django config files
func (d DjangoDetector) DoFrameworkDetection(language *model.Language, files *[]string) {
	var djangoFiles []string
	var configDjangoFiles []string

	for _, filename := range d.GetDjangoFilenames() {
		filePy := utils.GetFile(files, filename)
		utils.AddToArrayIfValueExist(&djangoFiles, filePy)
	}

	for _, filename := range d.GetConfigDjangoFilenames() {
		configFile := utils.GetFile(files, filename)
		utils.AddToArrayIfValueExist(&configDjangoFiles, configFile)
	}

	if hasFramework(&djangoFiles, "from django.") || hasFramework(&configDjangoFiles, "django") || hasFramework(&configDjangoFiles, "Django") {
		language.Frameworks = append(language.Frameworks, "Django")
	}
}

// DoPortsDetection searches for the port in /manage.py
func (d DjangoDetector) DoPortsDetection(component *model.Component, ctx *context.Context) {
	ports := []int{}
	appFileInfos := d.GetApplicationFileInfos(component.Path, ctx)
	if len(appFileInfos) == 0 {
		return
	}

	for _, appFileInfo := range appFileInfos {
		fileBytes, err := utils.GetApplicationFileBytes(appFileInfo)
		if err != nil {
			continue
		}

		re := regexp.MustCompile(`.default_port\s*=\s*"([^"]*)`)
		component.Ports = utils.FindAllPortsSubmatch(re, string(fileBytes), 1)
		if len(ports) > 0 {
			component.Ports = ports
			return
		}
	}

}
