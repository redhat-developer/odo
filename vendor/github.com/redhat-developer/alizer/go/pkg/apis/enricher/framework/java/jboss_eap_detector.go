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

type JBossEAPDetector struct{}

func (o JBossEAPDetector) GetSupportedFrameworks() []string {
	return []string{"JBoss EAP"}
}

// DoFrameworkDetection uses the groupId and artifactId to check for the framework name
func (o JBossEAPDetector) DoFrameworkDetection(language *model.Language, config string) {
	if hasFwk, _ := hasFramework(config, "org.jboss.eap.plugins", "eap-maven-plugin"); hasFwk {
		language.Frameworks = append(language.Frameworks, "JBoss EAP")
	}
}

func (o JBossEAPDetector) DoPortsDetection(component *model.Component, ctx *context.Context) {
	// Not implemented
}
