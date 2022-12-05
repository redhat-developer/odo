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
	"regexp"

	"github.com/redhat-developer/alizer/go/pkg/apis/model"
	"github.com/redhat-developer/alizer/go/pkg/utils"
	"golang.org/x/mod/modfile"
)

type BeegoDetector struct{}

func (b BeegoDetector) GetSupportedFrameworks() []string {
	return []string{"Beego"}
}

func (b BeegoDetector) DoFrameworkDetection(language *model.Language, goMod *modfile.File) {
	if hasFramework(goMod.Require, "github.com/beego/beego") {
		language.Frameworks = append(language.Frameworks, "Beego")
	}
}

type ApplicationPropertiesFile struct {
	Dir  string
	File string
}

func (b BeegoDetector) DoPortsDetection(component *model.Component) {
	bytes, err := utils.ReadAnyApplicationFile(component.Path, []model.ApplicationFileInfo{
		{
			Dir:  "conf",
			File: "app.conf",
		},
	})
	if err != nil {
		return
	}
	re := regexp.MustCompile(`httpport\s*=\s*(\d+)`)
	component.Ports = utils.FindAllPortsSubmatch(re, string(bytes), 1)
}
