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

	"github.com/redhat-developer/alizer/go/pkg/apis/model"
)

type WildFlyDetector struct{}

func (o WildFlyDetector) GetSupportedFrameworks() []string {
	return []string{"WildFly"}
}

// DoFrameworkDetection uses the groupId and artifactId to check for the framework name
func (o WildFlyDetector) DoFrameworkDetection(language *model.Language, config string) {
	if hasFwk, _ := hasFramework(config, "org.wildfly.plugins", "wildfly-maven-plugin"); hasFwk {
		language.Frameworks = append(language.Frameworks, "WildFly")
	}
}

func (o WildFlyDetector) DoPortsDetection(component *model.Component, ctx *context.Context) {
	// Not implemented
}
