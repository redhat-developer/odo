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

import (
	"github.com/redhat-developer/alizer/go/pkg/apis/language"
	"github.com/redhat-developer/alizer/go/pkg/utils"
)

type DjangoDetector struct{}

func (d DjangoDetector) DoFrameworkDetection(language *language.Language, files *[]string) {
	managePy := utils.GetFile(files, "manage.py")
	urlsPy := utils.GetFile(files, "urls.py")
	wsgiPy := utils.GetFile(files, "wsgi.py")
	asgiPy := utils.GetFile(files, "asgi.py")

	djangoFiles := []string{}
	utils.AddToArrayIfValueExist(&djangoFiles, managePy)
	utils.AddToArrayIfValueExist(&djangoFiles, urlsPy)
	utils.AddToArrayIfValueExist(&djangoFiles, wsgiPy)
	utils.AddToArrayIfValueExist(&djangoFiles, asgiPy)

	if hasFramework(files, "from django.") {
		language.Frameworks = append(language.Frameworks, "Django")
	}
}
