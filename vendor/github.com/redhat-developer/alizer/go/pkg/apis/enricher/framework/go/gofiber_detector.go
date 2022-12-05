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
	"os"
	"regexp"

	"github.com/redhat-developer/alizer/go/pkg/apis/model"
	"github.com/redhat-developer/alizer/go/pkg/utils"
	"golang.org/x/mod/modfile"
)

type GoFiberDetector struct{}

func (g GoFiberDetector) GetSupportedFrameworks() []string {
	return []string{"GoFiber"}
}

func (g GoFiberDetector) DoFrameworkDetection(language *model.Language, goMod *modfile.File) {
	if hasFramework(goMod.Require, "github.com/gofiber/fiber") {
		language.Frameworks = append(language.Frameworks, "GoFiber")
	}
}

func (g GoFiberDetector) DoPortsDetection(component *model.Component) {
	files, err := utils.GetFilePathsFromRoot(component.Path)
	if err != nil {
		return
	}

	matchRegexRule := model.PortMatchRules{
		MatchIndexRegexes: []model.PortMatchRule{
			{
				Regex:     regexp.MustCompile(`.Listen\(([^,)]*)`),
				ToReplace: ".Listen(",
			},
		},
	}
	for _, file := range files {
		bytes, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		ports := GetPortFromFileGo(matchRegexRule, string(bytes))
		if len(ports) > 0 {
			component.Ports = ports
			return
		}
	}
}
