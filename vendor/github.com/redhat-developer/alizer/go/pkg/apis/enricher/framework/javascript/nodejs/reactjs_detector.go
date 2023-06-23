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
	"os"

	"github.com/redhat-developer/alizer/go/pkg/apis/model"
	"github.com/redhat-developer/alizer/go/pkg/utils"
)

type ReactJsDetector struct{}

func (r ReactJsDetector) GetSupportedFrameworks() []string {
	return []string{"React"}
}

// DoFrameworkDetection uses a tag to check for the framework name
func (r ReactJsDetector) DoFrameworkDetection(language *model.Language, config string) {
	if hasFramework(config, "react") {
		language.Frameworks = append(language.Frameworks, "React")
	}
}

// DoPortsDetection searches for the port in the env var, .env file, and package.json
func (r ReactJsDetector) DoPortsDetection(component *model.Component, ctx *context.Context) {
	// check if port is set on env var
	portValue := os.Getenv("PORT")
	if port, err := utils.GetValidPort(portValue); err == nil {
		component.Ports = []int{port}
		return
	}
	// check if port is set on .env file
	port := utils.GetPortValueFromEnvFile(component.Path, `PORT=(\d*)`)
	if utils.IsValidPort(port) {
		component.Ports = []int{port}
		return
	}
	// check if port is set in start script in package.json
	port = getPortFromStartScript(component.Path, []string{`PORT=(\d*)`})
	if utils.IsValidPort(port) {
		component.Ports = []int{port}
		return
	}
}
