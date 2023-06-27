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
	"github.com/redhat-developer/alizer/go/pkg/apis/model"
)

type LaravelDetector struct{}

func (d LaravelDetector) GetSupportedFrameworks() []string {
	return []string{"Laravel"}
}

// DoFrameworkDetection uses a tag to check for the framework name
func (d LaravelDetector) DoFrameworkDetection(language *model.Language, config string) {
	if hasFramework(config, "laravel") {
		language.Frameworks = append(language.Frameworks, "Laravel")
	}
}

func (d LaravelDetector) DoPortsDetection(component *model.Component, ctx *context.Context) {
	// Not implemented yet
}
