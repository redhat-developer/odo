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
	"github.com/redhat-developer/alizer/go/pkg/utils"
)

type FlaskDetector struct{}

func (d FlaskDetector) GetSupportedFrameworks() []string {
	return []string{"Flask"}
}

func (d FlaskDetector) DoFrameworkDetection(language *model.Language, files *[]string) {
	appPy := utils.GetFile(files, "app.py")
	wsgiPy := utils.GetFile(files, "wsgi.py")
	requirementsTxt := utils.GetFile(files, "requirements.txt")
	projectToml := utils.GetFile(files, "pyproject.toml")

	flaskFiles := []string{}
	configFlaskFiles := []string{}
	utils.AddToArrayIfValueExist(&flaskFiles, appPy)
	utils.AddToArrayIfValueExist(&flaskFiles, wsgiPy)
	utils.AddToArrayIfValueExist(&configFlaskFiles, requirementsTxt)
	utils.AddToArrayIfValueExist(&configFlaskFiles, projectToml)

	if hasFramework(&flaskFiles, "from flask ") || hasFramework(&configFlaskFiles, "Flask") || hasFramework(&configFlaskFiles, "flask") {
		language.Frameworks = append(language.Frameworks, "Flask")
	}
}

func (d FlaskDetector) DoPortsDetection(component *model.Component, ctx *context.Context) {
	// Not implemented yet
}
