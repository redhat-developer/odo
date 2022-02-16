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
	"encoding/xml"
	"io/ioutil"
	"os"
	"strings"

	"github.com/redhat-developer/alizer/go/pkg/apis/language"
	"github.com/redhat-developer/alizer/go/pkg/schema"
	"github.com/redhat-developer/alizer/go/pkg/utils"
)

type DotNetDetector struct{}

func (m DotNetDetector) DoFrameworkDetection(language *language.Language, configFilePath string) {
	framework := getFrameworks(configFilePath)
	if framework == "" {
		return
	}
	var frameworks []string
	if strings.Contains(framework, ";") {
		frameworks = strings.Split(framework, ";")
	} else {
		frameworks = []string{framework}
	}

	for _, frm := range frameworks {
		if !utils.Contains(language.Frameworks, frm) {
			language.Frameworks = append(language.Frameworks, frm)
		}
	}
}

func getFrameworks(configFilePath string) string {
	xmlFile, err := os.Open(configFilePath)
	if err != nil {
		return ""
	}
	byteValue, _ := ioutil.ReadAll(xmlFile)

	var proj schema.DotNetProject
	xml.Unmarshal(byteValue, &proj)

	defer xmlFile.Close()
	if proj.PropertyGroup.TargetFramework != "" {
		return proj.PropertyGroup.TargetFramework
	} else if proj.PropertyGroup.TargetFrameworkVersion != "" {
		return proj.PropertyGroup.TargetFrameworkVersion
	} else if proj.PropertyGroup.TargetFrameworks != "" {
		return proj.PropertyGroup.TargetFrameworks
	}
	return ""
}
