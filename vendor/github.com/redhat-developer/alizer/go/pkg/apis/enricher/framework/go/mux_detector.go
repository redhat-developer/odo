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
	"github.com/redhat-developer/alizer/go/pkg/apis/language"
	"golang.org/x/mod/modfile"
)

type MuxDetector struct{}

func (e MuxDetector) DoFrameworkDetection(language *language.Language, goMod *modfile.File) {
	if hasFramework(goMod.Require, "github.com/gorilla/mux") {
		language.Frameworks = append(language.Frameworks, "Mux")
	}
}
