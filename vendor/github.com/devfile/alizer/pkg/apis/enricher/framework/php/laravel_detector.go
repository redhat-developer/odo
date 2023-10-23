/*******************************************************************************
 * Copyright (c) 2023 Red Hat, Inc.
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

	"github.com/devfile/alizer/pkg/apis/model"
	"github.com/devfile/alizer/pkg/utils"
)

type LaravelDetector struct{}

func (d LaravelDetector) GetSupportedFrameworks() []string {
	return []string{"Laravel"}
}

func (d LaravelDetector) GetApplicationFileInfos(componentPath string, ctx *context.Context) []model.ApplicationFileInfo {
	// laravel enricher does not apply source code detection.
	// It only detects ports declared as env vars
	return nil
}

// DoFrameworkDetection uses a tag to check for the framework name
func (d LaravelDetector) DoFrameworkDetection(language *model.Language, config string) {
	if hasFramework(config, "laravel") {
		language.Frameworks = append(language.Frameworks, "Laravel")
	}
}

// DoPortsDetection for Laravel will check if there is any .env file inside the component
// configuring the APP_PORT variable which is dedicated to port configuration.
func (d LaravelDetector) DoPortsDetection(component *model.Component, ctx *context.Context) {
	regexes := []string{`APP_PORT=(\d*)`}
	// Case ENV file
	ports := utils.GetPortValuesFromEnvFile(component.Path, regexes)
	if len(ports) > 0 {
		component.Ports = ports
		return
	}
	// Case env var defined inside dockerfile
	ports, err := utils.GetEnvVarPortValueFromDockerfile(component.Path, []string{"APP_PORT"})
	if len(ports) > 0 && err != nil {
		component.Ports = ports
		return
	}

}
