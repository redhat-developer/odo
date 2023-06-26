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

	"github.com/redhat-developer/alizer/go/pkg/apis/model"
	"github.com/redhat-developer/alizer/go/pkg/utils"
	"golang.org/x/mod/modfile"
)

type GinDetector struct{}

func (g GinDetector) GetSupportedFrameworks() []string {
	return []string{"Gin"}
}

// DoFrameworkDetection uses a tag to check for the framework name
func (g GinDetector) DoFrameworkDetection(language *model.Language, goMod *modfile.File) {
	if hasFramework(goMod.Require, "github.com/gin-gonic/gin") {
		language.Frameworks = append(language.Frameworks, "Gin")
	}
}

func (g GinDetector) DoPortsDetection(component *model.Component, ctx *context.Context) {
	files, err := utils.GetCachedFilePathsFromRoot(component.Path, ctx)
	if err != nil {
		return
	}

	matchRegexRules := model.PortMatchRules{
		MatchIndexRegexes: []model.PortMatchRule{
			{
				Regex:     regexp.MustCompile(`.Run\(([^,)]*)`),
				ToReplace: ".Run(",
			},
		},
	}

	ports := GetPortFromFilesGo(matchRegexRules, files)
	if len(ports) > 0 {
		component.Ports = ports
	}
}
