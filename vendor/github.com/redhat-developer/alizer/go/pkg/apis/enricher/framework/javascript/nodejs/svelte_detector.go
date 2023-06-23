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

	"github.com/redhat-developer/alizer/go/pkg/apis/model"
	"github.com/redhat-developer/alizer/go/pkg/utils"
)

type SvelteDetector struct{}

func (s SvelteDetector) GetSupportedFrameworks() []string {
	return []string{"Svelte"}
}

// DoFrameworkDetection uses a tag to check for the framework name
func (s SvelteDetector) DoFrameworkDetection(language *model.Language, config string) {
	if hasFramework(config, "svelte") {
		language.Frameworks = append(language.Frameworks, "Svelte")
	}
}

// DoPortsDetection searches for the port in package.json
func (s SvelteDetector) DoPortsDetection(component *model.Component, ctx *context.Context) {
	// check if port is set in start script in package.json
	port := getPortFromDevScript(component.Path, []string{`--port (\d*)`, `PORT=(\d*)`})
	if utils.IsValidPort(port) {
		component.Ports = []int{port}
	}
}
