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
package recognizer

import (
	"github.com/redhat-developer/alizer/go/pkg/apis/model"
	"golang.org/x/mod/modfile"
)

type GoFiberDetector struct{}

func (e GoFiberDetector) DoFrameworkDetection(language *model.Language, goMod *modfile.File) {
	if hasFramework(goMod.Require, "github.com/gofiber/fiber") {
		language.Frameworks = append(language.Frameworks, "GoFiber")
	}
}
