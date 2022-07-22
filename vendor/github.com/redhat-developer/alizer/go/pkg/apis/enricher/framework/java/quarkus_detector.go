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
package recognizer

import "github.com/redhat-developer/alizer/go/pkg/apis/model"

type QuarkusDetector struct{}

func (q QuarkusDetector) DoFrameworkDetection(language *model.Language, config string) {
	if hasFwk, _ := hasFramework(config, "io.quarkus"); hasFwk {
		language.Frameworks = append(language.Frameworks, "Quarkus")
	}
}
