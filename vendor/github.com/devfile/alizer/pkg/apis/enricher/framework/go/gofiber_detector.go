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
	"context"
	"regexp"

	"github.com/devfile/alizer/pkg/apis/model"
	"github.com/devfile/alizer/pkg/utils"
	"golang.org/x/mod/modfile"
)

type GoFiberDetector struct{}

func (g GoFiberDetector) GetSupportedFrameworks() []string {
	return []string{"GoFiber"}
}

// DoFrameworkDetection uses a tag to check for the framework name
func (g GoFiberDetector) DoFrameworkDetection(language *model.Language, goMod *modfile.File) {
	if hasFramework(goMod.Require, "github.com/gofiber/fiber") {
		language.Frameworks = append(language.Frameworks, "GoFiber")
	}
}

func (g GoFiberDetector) DoPortsDetection(component *model.Component, ctx *context.Context) {
	files, err := utils.GetCachedFilePathsFromRoot(component.Path, ctx)
	if err != nil {
		return
	}

	matchRegexRules := model.PortMatchRules{
		MatchIndexRegexes: []model.PortMatchRule{
			{
				Regex:     regexp.MustCompile(`.Listen\(([^,)]*)`),
				ToReplace: ".Listen(",
			},
		},
	}
	ports := GetPortFromFilesGo(matchRegexRules, files)
	if len(ports) > 0 {
		component.Ports = ports
	}
}
